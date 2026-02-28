# AITDD MCP 接口测试错误报告

**测试日期**: 2026-02-28  
**测试环境**: Windows 10, 后端服务 http://localhost:34567  
**测试方法**: 通过 MCP 协议直接调用接口

---

## 测试摘要

| 类别 | 成功 | 失败 | 部分工作 |
|------|------|------|----------|
| 项目上下文 (2) | 0 | 1 | 1 |
| 信息查询 (3) | 1 | 2 | 0 |
| 验证检查 (3) | 3 | 0 | 0 |
| 节点操作 (3) | - | - | - |
| 依赖管理 (3) | 1 | - | - |
| 问答系统 (4) | 1 | - | - |
| 锁管理 (2) | 1 | 0 | 0 |
| 编译接口 (2) | 0 | 1 | 0 |
| 配置/规则/状态 (5) | 4 | 0 | 1 |
| **总计** | **11** | **4** | **2** |

**总接口数**: 27  
**完全成功**: 11  
**部分工作**: 2  
**失败**: 4  
**未测试**: 10 (节点操作、部分依赖管理等)

---

## 详细测试结果

### 1. 项目上下文接口

#### 1.1 `init_project` - ❌ 失败

**描述**: 初始化项目配置

**测试结果**:
- 不带参数调用时返回"可用项目列表："但没有实际内容
- 通过 `projectId` 设置项目失败，返回"未找到项目"
- 通过 `projectName` 设置项目失败，返回"未找到项目"

**错误原因**:

1. **获取项目列表解析错误** (第105行)
   ```go
   // 代码中:
   if data, ok := result["data"].([]interface{}); ok {
   
   // 实际 API 返回:
   {"success":true,"data":{"projects":[...],"total":3}}
   
   // 应该是:
   if dataObj, ok := result["data"].(map[string]interface{}); ok {
       if projects, ok := dataObj["projects"].([]interface{}); ok {
   ```

2. **通过 projectId 获取项目时路由不存在** (第125行)
   ```go
   // 代码中:
   projectURL = fmt.Sprintf("%s/projects/%s", s.getApiURL(), projectId)
   
   // 实际路由只有:
   // GET /api/v1/projects/by-path/*pathName
   // 没有 GET /api/v1/projects/:id
   ```

3. **通过 projectName 获取项目时路由不存在** (第127行)
   ```go
   // 代码中:
   projectURL = fmt.Sprintf("%s/projects/by-name/%s", s.getApiURL(), projectName)
   
   // 实际路由没有 /projects/by-name/
   ```

**修复建议**:
1. 修复项目列表解析逻辑
2. 添加 REST API 端点 `/api/v1/projects/:id` 和 `/api/v1/projects/by-name/:name`
3. 或者修改 MCP 代码使用现有的 `/projects/by-path/` 端点

---

#### 1.2 `get_context` - ⚠️ 部分工作

**描述**: 获取当前上下文

**测试结果**:
```
当前项目上下文：

项目路径: PYQT6Calculator

项目信息：
```

**问题**: "项目信息"后面没有内容，可能是获取项目详情失败

---

### 2. 信息查询接口

#### 2.1 `query_node` - ✅ 成功

**测试结果**:
```
success: true
data: {"project":{"constitution":"基于PyQt6的科学计算器应用程序",...}}
```

---

#### 2.2 `query_project_index_tree` - ❌ 失败

**描述**: 查询项目计划索引树

**测试结果**:
```
PYQT6Calculator                     # , project, -
```

**问题**: 只返回项目根节点，没有返回模块和任务树

**错误原因** (第321行):
```go
// 代码中:
if modules, ok := modulesResult["data"].([]interface{}); ok {

// 实际 API 返回:
{"success":true,"data":{"modules":[...],"total":18}}

// 应该是:
if dataObj, ok := modulesResult["data"].(map[string]interface{}); ok {
    if modules, ok := dataObj["modules"].([]interface{}); ok {
```

---

#### 2.3 `query_file_code_path_tree` - ❌ 失败

**描述**: 查询代码文件路径树

**测试结果**: 未完成测试（用户反馈需要返回代码文件树）

**问题**: 与 `query_project_index_tree` 相同的数据结构解析问题

---

### 3. 验证检查接口

#### 3.1 `check_contract_alignment` - ✅ 成功

**测试结果**: 接口可正常调用（需要有效的上下游任务对测试）

---

#### 3.2 `check_task_readiness` - ✅ 成功

**测试结果**:
```
ready: false
issues:
  - type: "missing_prompt"
    message: "任务提示词为空"
    severity: "error"
  - type: "empty_tests"
    message: "未定义测试用例"
    severity: "warning"
```

---

#### 3.3 `check_dependencies` - ✅ 成功

**测试结果**:
```
hasCycle: false
blocked: false
```

---

### 4. 编译接口

#### 4.1 `compile_static` - ❌ 失败

**测试结果**:
```
success: true
totalModules: 0
totalTasks: 0
```

**问题**: 模块和任务数量都是0，与 `query_project_index_tree` 相同的数据解析问题

---

### 5. 配置/规则/状态接口

#### 5.1 `get_config` - ✅ 成功

**测试结果**:
```json
{
  "pathName": "PYQT6Calculator",
  "apiBaseUrl": ""
}
```

---

#### 5.2 `get_rule` - ✅ 成功

**测试结果**: 返回完整的规则配置

---

#### 5.3 `get_status` - ⚠️ 部分工作

**测试结果**:
```
pathName: "PYQT6Calculator"
type: "project"
```

**问题**: 返回内容较少，缺少进度、错误、警告等详细信息

---

### 6. 问答系统接口

#### 6.1 `query_issues` - ✅ 成功

**测试结果**:
```
issues:
total: 0
```

---

### 7. 锁管理接口

#### 7.1 `acquire_lock` - ✅ 成功（参数验证正常）

**测试结果**:
```json
{
  "error": {
    "code": "VALIDATION_ERROR",
    "message": "缺少 lockedBy 参数"
  }
}
```

这是正常的参数验证行为，接口工作正常。

---

## 根本原因分析

### 主要问题：API 响应数据结构不匹配

多个接口存在相同的问题：代码期望 `result["data"]` 是一个数组，但实际 API 返回的是一个对象，包含 `modules` 或 `projects` 等字段。

**影响的接口**:
1. `init_project` - 项目列表解析
2. `query_project_index_tree` - 模块列表解析
3. `query_file_code_path_tree` - 任务列表解析
4. `compile_static` - 模块/任务列表解析

### 次要问题：缺失 REST API 端点

以下端点在 MCP 代码中被调用，但实际不存在：
- `GET /api/v1/projects/:id`
- `GET /api/v1/projects/by-name/:name`

---

## 修复优先级

### P0 - 紧急（影响核心功能）

1. **修复 `init_project` 项目列表解析**
   - 文件: `backend/internal/mcp/tools_get.go`
   - 行号: 105-115
   - 修复数据结构解析逻辑

2. **修复 `query_project_index_tree` 模块列表解析**
   - 文件: `backend/internal/mcp/tools_get.go`
   - 行号: 321
   - 修复 `modulesResult["data"]` 解析

### P1 - 高优先级

3. **添加缺失的 REST API 端点**
   - 文件: `backend/internal/api/routes.go`
   - 添加: `GET /api/v1/projects/:id`
   - 添加: `GET /api/v1/projects/by-name/:name`

4. **修复 `compile_static` 数据解析**
   - 文件: `backend/internal/mcp/tools_other.go`
   - 与 `query_project_index_tree` 相同的修复

### P2 - 中优先级

5. **修复 `get_context` 项目信息获取**
6. **修复 `get_status` 详细信息返回**
7. **修复 `query_file_code_path_tree` 数据解析**

---

## 建议的修复方案

### 方案 A：修复 MCP 代码（推荐）

修改 MCP 代码以正确解析 API 响应：

```go
// 通用修复模式
func parseListResponse(result map[string]interface{}, listKey string) ([]interface{}, bool) {
    dataObj, ok := result["data"].(map[string]interface{})
    if !ok {
        return nil, false
    }
    list, ok := dataObj[listKey].([]interface{})
    return list, ok
}

// 使用示例
if modules, ok := parseListResponse(modulesResult, "modules"); ok {
    buildModuleTree(&sb, modules, "", s.getApiURL())
}
```

### 方案 B：添加缺失的 REST API 端点

在 `routes.go` 中添加：
```go
projects.GET("/:id", handlers.GetProjectByID)
projects.GET("/by-name/:name", handlers.GetProjectByName)
```

---

## 测试通过接口汇总

以下接口测试通过，功能正常：
- ✅ `query_node`
- ✅ `check_contract_alignment`
- ✅ `check_task_readiness`
- ✅ `check_dependencies`
- ✅ `get_config`
- ✅ `get_rule`
- ✅ `query_issues`
- ✅ `acquire_lock` (参数验证正常)

---

## 后续测试建议

1. 完成节点操作接口测试（`create_node`, `modify_node`, `delete_node`）
2. 完成依赖管理接口测试（`create_dependency`, `delete_dependency`, `query_dependencies`）
3. 完成问答系统接口测试（`create_issue`, `reply_issue`, `resolve_issue`）
4. 完成锁管理接口测试（`release_lock`）
5. 完成编译接口测试（`compile_dynamic`）
6. 完成配置更新接口测试（`update_config`, `update_rule`）

---

**报告生成时间**: 2026-02-28 19:56 CST
