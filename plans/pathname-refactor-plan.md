# AITDD pathName 字段问题分析和修复计划

## 1. 问题发现

### 1.1 当前问题
查询模块数据时发现 pathName 值不正确：
```json
{
  "name": "主界面模块",
  "pathName": "主界面模块",// 错误！应该是 "PYQT6Calculator/主界面模块"
  "projectId": "83c4c9d9-3c8e-4844-8f85-1a09398d708f"
}
```

### 1.2 正确的 pathName 格式设计
根据 `services/module_service.go` 和 `services/task_service.go` 中的代码：

| 资源类型 | pathName 格式 | 示例 |
|---------|--------------|------|
| 项目 | `{项目名称}` | `PYQT6Calculator` |
| 模块 | `{项目pathName}/{模块名称}` | `PYQT6Calculator/主界面模块` |
| 任务 | `{模块pathName}/{任务名称}` | `PYQT6Calculator/主界面模块/初始化UI` |

## 2. 问题根源分析

### 2.1 代码对比

**Service 层（正确实现）** - [`module_service.go:67`](backend/internal/services/module_service.go:67):
```go
basePathName := fmt.Sprintf("%s/%s", project.PathName, name)
pathName := basePathName
```

**Handler 层（问题代码）** - [`module.go:77-91`](backend/internal/api/handlers/module.go:77):
```go
// 生成唯一的 pathName
pathName := req.PathName
if pathName == "" {
    pathName = req.Name// 问题：只用模块名，没有项目前缀！
}
```

### 2.2 数据流问题
```
导入脚本 → API Handler (module.go) → 数据库
     ↓
   使用错误的 pathName 生成逻辑（只用模块名）
```

Service 层有正确的 pathName 生成逻辑，但 API Handler 没有调用 Service 层，而是直接操作数据库。

## 3. 受影响的代码位置

### 3.1 Module Handler - [`backend/internal/api/handlers/module.go`](backend/internal/api/handlers/module.go)

| 函数 | 行号 | 问题 |
|-----|------|------|
| `CreateModule` | 77-91 | pathName 只用模块名，缺少项目前缀 |
| `UpdateModule` | - | 重命名时需要级联更新 pathName |

### 3.2 Task Handler - [`backend/internal/api/handlers/task.go`](backend/internal/api/handlers/task.go)

| 函数 | 行号 | 问题 |
|-----|------|------|
| `CreateTask` | 99-114 | pathName 只用任务名，缺少模块前缀 |
| `UpdateTask` | - | 重命名时需要级联更新 pathName |

### 3.3 Project Handler - [`backend/internal/api/handlers/project.go`](backend/internal/api/handlers/project.go)

| 函数 | 行号 | 状态 |
|-----|------|------|
| `CreateProject` | 47-57 | ✅ 正确（项目 pathName = 项目名） |
| `UpdateProject` | 201-213 | ✅ 有级联更新逻辑 |

## 4. 修复方案

### 4.1 Module Handler 修复

**文件**: [`backend/internal/api/handlers/module.go`](backend/internal/api/handlers/module.go)

**CreateModule 函数修复** (行 77-91):
```go
// 修复前
pathName := req.PathName
if pathName == "" {
    pathName = req.Name
}

// 修复后
// 获取项目信息以生成正确的 pathName
var project models.Project
if err := database.DB.First(&project, "id = ?", req.ProjectID).Error; err != nil {
    BadRequest(c, "项目不存在")
    return
}

basePathName := fmt.Sprintf("%s/%s", project.PathName, req.Name)
pathName := basePathName
suffix := 1
for {
    var count int64
    database.DB.Model(&models.Module{}).Where("path_name = ?", pathName).Count(&count)
    if count == 0 {
        break
    }
    pathName = fmt.Sprintf("%s-%d", basePathName, suffix)
    suffix++
}
```

### 4.2 Task Handler 修复

**文件**: [`backend/internal/api/handlers/task.go`](backend/internal/api/handlers/task.go)

**CreateTask 函数修复** (行 99-114):
```go
// 修复前
pathName := req.PathName
if pathName == "" {
    pathName = req.Name
}

// 修复后
// 获取模块信息以生成正确的 pathName
var module models.Module
if err := database.DB.First(&module, "id = ?", req.ModuleID).Error; err != nil {
    BadRequest(c, "模块不存在")
    return
}

basePathName := fmt.Sprintf("%s/%s", module.PathName, req.Name)
pathName := basePathName
suffix := 1
for {
    var count int64
    database.DB.Model(&models.Task{}).Where("path_name = ?", pathName).Count(&count)
    if count == 0 {
        break
    }
    pathName = fmt.Sprintf("%s-%d", basePathName, suffix)
    suffix++
}
```

## 5. MCP 工具 pathName 参数使用分析

### 5.1 使用 pathName 的 MCP 工具

| 工具名称 | 参数名 | 描述 | API 调用 |
|---------|-------|------|---------|
| `get_project_info` | `pathName` | 项目路径名称 | `/projects/by-path/{pathName}` |
| `get_all_modules` | `pathName` | 项目路径名称 | `/modules?projectPathName={pathName}` |
| `get_module_overview` | `pathName` | 模块路径名称 | `/modules/by-path/{pathName}` |
| `get_module_tasks` | `pathName` | 模块路径名称 | `/modules/by-path/{pathName}?action=tasks` |
| `get_task_detail` | `pathName` | 任务路径名称 | `/tasks/by-path/{pathName}` |
| `delete_module` | `pathName` | 模块路径名称 | `DELETE /modules/by-path/{pathName}` |
| `create_task` | `pathName` | 所属模块路径名称 | `POST /tasks` (需要先查模块ID) |
| `update_task` | `pathName` | 任务路径名称 | `PUT /tasks/by-path/{pathName}` |
| `delete_task` | `pathName` | 任务路径名称 | `DELETE /tasks/by-path/{pathName}` |
| `lock_resource` | `pathName` | 资源路径名称 | `POST /lock/by-path/{pathName}` |
| `unlock_resource` | `pathName` | 资源路径名称 | `POST /lock/unlock-by-path/{pathName}` |
| `create_module_dependency` | `pathName`, `dependsOnPathName` | 两个模块路径 | 需要先查模块ID |
| `create_task_dependency` | `upstreamPathName`, `downstreamPathName` | 两个任务路径 | 需要先查任务ID |

### 5.2 MCP 工具的 API 调用流程

```
MCP 工具接收 pathName
    ↓
通过 /xxx/by-path/{pathName} 获取资源 ID
    ↓
使用 ID 调用其他 API
```

**示例** - `create_task` 工具流程：
```go
// 1. 通过 pathName 获取模块 ID
moduleResp, err := http.Get(fmt.Sprintf("%s/modules/by-path/%s", s.getApiURL(), modulePathName))

// 2. 提取模块 ID
moduleID := moduleData["id"]

// 3. 创建任务（使用模块 ID）
createResp, err := http.Post(fmt.Sprintf("%s/tasks", s.getApiURL()), taskData)
```

## 6. 修复步骤

### Phase 1: 修复 Handler 层 pathName 生成逻辑
1. [ ] 修复 `module.go` - `CreateModule` 函数
2. [ ] 修复 `task.go` - `CreateTask` 函数
3. [ ] 验证级联更新逻辑（重命名时 pathName 同步更新）

### Phase 2: 修复现有数据
1. [ ] 创建数据修复脚本，更新现有模块的 pathName
2. [ ] 创建数据修复脚本，更新现有任务的 pathName
3. [ ] 执行数据修复

### Phase 3: 测试验证
1. [ ] 测试创建新模块 - 验证 pathName 格式正确
2. [ ] 测试创建新任务 - 验证 pathName 格式正确
3. [ ] 测试 MCP 工具 - 验证 pathName 参数正确工作

## 7. 数据修复 SQL

### 7.1 修复模块 pathName
```sql
-- 更新模块的 pathName，添加项目前缀
UPDATE modules m
JOIN projects p ON m.project_id = p.id
SET m.path_name = CONCAT(p.path_name, '/', m.name)
WHERE m.path_name NOT LIKE CONCAT(p.path_name, '/%');
```

### 7.2 修复任务 pathName
```sql
-- 更新任务的 pathName，添加模块前缀
UPDATE tasks t
JOIN modules m ON t.module_id = m.id
SET t.path_name = CONCAT(m.path_name, '/', t.name)
WHERE t.path_name NOT LIKE CONCAT(m.path_name, '/%');
```

## 8. 风险评估

| 风险 | 影响 | 缓解措施 |
|-----|------|---------|
| 现有数据 pathName 不正确 | MCP 工具无法查询 | 执行数据修复 SQL |
| pathName 包含特殊字符 | URL 编码问题 | 前端/客户端需要正确编码 |
| 重命名后 pathName 变化 | 依赖 pathName 的引用失效 | 级联更新已实现 |

## 9. 文件修改清单

| 文件 | 修改内容 | 优先级 |
|-----|---------|-------|
| `backend/internal/api/handlers/module.go` | 修复 CreateModule pathName 生成 | 高 |
| `backend/internal/api/handlers/task.go` | 修复 CreateTask pathName 生成 | 高 |
| 数据库 | 执行数据修复 SQL | 高 |
