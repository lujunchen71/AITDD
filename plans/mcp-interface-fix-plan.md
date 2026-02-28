# MCP 接口修复方案

## 概述

本文档详细描述了三个 MCP 接口的问题分析和修复方案。

---

## 1. `init_project` 接口修复

### 1.1 问题描述

- **期望行为**：不填参数时应该列出所有项目供选择，格式为简洁的 YAML
- **当前行为**：返回的是冗长的文本格式，不是期望的 YAML 格式

### 1.2 期望输出格式

**列表模式（未传参数）**：
```yaml
projects:
  - name: "AITDD"
    pathName: "aitdd"
    description: "AI驱动的设计与开发工具"
  - name: "PVZ"
    pathName: "pvz"
    description: "植物大战僵尸游戏"
```

**设置模式（传入 pathName）**：
```yaml
success: true
projectName: "AITDD"
pathName: "aitdd"
```

### 1.3 问题根因分析

查看 [`handleInitProjectImpl`](backend/internal/mcp/tools_get.go:84) 函数：

```go
// 当前代码（第 102-117 行）
var sb strings.Builder
sb.WriteString("可用项目列表：\n\n")
if data, ok := result["data"].([]interface{}); ok {
    for _, item := range data {
        if project, ok := item.(map[string]interface{}); ok {
            id, _ := project["id"].(string)
            name, _ := project["name"].(string)
            desc, _ := project["description"].(string)
            pn, _ := project["pathName"].(string)
            sb.WriteString(fmt.Sprintf("- ID: %s\n  名称: %s\n  路径: %s\n  简介: %s\n\n", id, name, pn, desc))
        }
    }
}
```

**问题**：
1. API 返回的数据结构是 `{"projects": [...], "total": N}`，但代码尝试从 `result["data"]` 获取
2. 输出格式不是期望的 YAML 格式

### 1.4 修复方案

**修改位置**：[`backend/internal/mcp/tools_get.go`](backend/internal/mcp/tools_get.go) 第 84-165 行

**修复后代码**：

```go
// handleInitProjectImpl 初始化项目配置
func (s *MCPServer) handleInitProjectImpl(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
    projectId, _ := getParam(request, "projectId")
    projectName, _ := getParam(request, "projectName")
    pathName, _ := getParam(request, "pathName")

    // 如果没有提供任何参数，返回项目列表
    if projectId == "" && projectName == "" && pathName == "" {
        resp, err := http.Get(fmt.Sprintf("%s/projects", s.getApiURL()))
        if err != nil {
            return mcp.NewToolResultText("请求失败：" + err.Error()), nil
        }
        defer resp.Body.Close()

        var result map[string]interface{}
        if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
            return mcp.NewToolResultText("解析失败：" + err.Error()), nil
        }

        // 格式化项目列表为 YAML 格式
        var sb strings.Builder
        sb.WriteString("projects:\n")
        
        // API 返回的结构是 {"projects": [...], "total": N}
        if projects, ok := result["projects"].([]interface{}); ok {
            for _, item := range projects {
                if project, ok := item.(map[string]interface{}); ok {
                    name, _ := project["name"].(string)
                    pn, _ := project["pathName"].(string)
                    desc, _ := project["constitution"].(string) // 使用 constitution 作为描述
                    if desc == "" {
                        desc, _ = project["description"].(string)
                    }
                    sb.WriteString(fmt.Sprintf("  - name: %q\n    pathName: %q\n    description: %q\n", name, pn, desc))
                }
            }
        }
        return mcp.NewToolResultText(sb.String()), nil
    }

    // 如果提供了 pathName，直接通过 pathName 获取项目
    if pathName != "" {
        resp, err := http.Get(fmt.Sprintf("%s/projects/by-path/%s", s.getApiURL(), pathName))
        if err != nil {
            return mcp.NewToolResultText("获取项目失败：" + err.Error()), nil
        }
        defer resp.Body.Close()

        if resp.StatusCode == 404 {
            return mcp.NewToolResultText(fmt.Sprintf("未找到项目: %s", pathName)), nil
        }

        var result map[string]interface{}
        if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
            return mcp.NewToolResultText("解析项目失败：" + err.Error()), nil
        }

        project, ok := result["project"].(map[string]interface{})
        if !ok {
            return mcp.NewToolResultText("项目数据格式错误"), nil
        }

        // 保存配置
        config := map[string]interface{}{
            "projectId":   project["id"],
            "projectName": project["name"],
            "pathName":    project["pathName"],
        }

        if err := s.configManager.SaveConfig(config); err != nil {
            return mcp.NewToolResultText("保存配置失败：" + err.Error()), nil
        }

        return mcp.NewToolResultText(fmt.Sprintf("success: true\nprojectName: %q\npathName: %q\n", project["name"], project["pathName"])), nil
    }

    // 如果提供了 projectId 或 projectName（兼容旧逻辑）
    if projectId != "" || projectName != "" {
        var projectURL string
        if projectId != "" {
            projectURL = fmt.Sprintf("%s/projects/%s", s.getApiURL(), projectId)
        } else {
            // 需要添加按名称查询的 API 支持
            projectURL = fmt.Sprintf("%s/projects/by-name/%s", s.getApiURL(), projectName)
        }

        resp, err := http.Get(projectURL)
        if err != nil {
            return mcp.NewToolResultText("获取项目失败：" + err.Error()), nil
        }
        defer resp.Body.Close()

        if resp.StatusCode == 404 {
            return mcp.NewToolResultText(fmt.Sprintf("未找到项目: %s", projectId+projectName)), nil
        }

        var result map[string]interface{}
        if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
            return mcp.NewToolResultText("解析项目失败：" + err.Error()), nil
        }

        // API 返回 {"project": {...}}
        project, ok := result["project"].(map[string]interface{})
        if !ok {
            return mcp.NewToolResultText("项目数据格式错误"), nil
        }

        // 保存配置
        config := map[string]interface{}{
            "projectId":   project["id"],
            "projectName": project["name"],
            "pathName":    project["pathName"],
        }

        if err := s.configManager.SaveConfig(config); err != nil {
            return mcp.NewToolResultText("保存配置失败：" + err.Error()), nil
        }

        return mcp.NewToolResultText(fmt.Sprintf("success: true\nprojectName: %q\npathName: %q\n", project["name"], project["pathName"])), nil
    }

    return mcp.NewToolResultText("请提供 projectId、projectName 或 pathName"), nil
}
```

### 1.5 需要修改的文件

| 文件 | 修改内容 |
|------|----------|
| `backend/internal/mcp/tools_get.go` | 修改 `handleInitProjectImpl` 函数 |

---

## 2. `query_project_index_tree` 接口修复

### 2.1 问题描述

- **期望行为**：返回完整的项目→模块→任务层级树
- **当前行为**：只返回一行内容或空内容

### 2.2 期望输出格式

```
aitdd                              # AITDD, project, -
├── auth                           # 认证模块, module, developing
│   ├── login                      # 用户登录, task, completed
│   ├── register                   # 用户注册, task, in_progress
│   └── logout                     # 用户登出, task, ready
├── core                           # 核心模块, module, designing
│   ├── engine                     # 引擎任务, task, ready
│   └── compiler                   # 编译器任务, task, ready
└── ui                             # 界面模块, module, designing
```

### 2.3 问题根因分析

查看 [`handleQueryProjectIndexTreeImpl`](backend/internal/mcp/tools_get.go:272) 函数：

```go
// 当前代码（第 302-303 行）
modulesURL := fmt.Sprintf("%s/modules?projectPathName=%s", s.getApiURL(), pathName)
```

**问题**：
1. **API 参数不匹配**：MCP 调用使用 `projectPathName` 参数，但 API 的 `GetModules` 只支持 `projectId` 参数
2. **API 返回数据结构**：API 返回 `{"modules": [...], "total": N}`，但代码尝试从 `result["data"]` 获取
3. **任务获取问题**：获取任务时使用 `action=tasks`，返回的是 `{"tasks": [...]}` 而非 `{"data": [...]}`

### 2.4 修复方案

有两种修复方案：

#### 方案 A：修改 MCP 层（推荐）

在 MCP 层先通过 pathName 获取项目 ID，再用项目 ID 查询模块。

**修改位置**：[`backend/internal/mcp/tools_get.go`](backend/internal/mcp/tools_get.go) 第 272-326 行

**修复后代码**：

```go
// handleQueryProjectIndexTreeImpl 查询项目计划索引树实现
func (s *MCPServer) handleQueryProjectIndexTreeImpl(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
    pathName, _ := getParam(request, "pathName")

    // 如果没有提供 pathName，使用当前配置的项目
    if pathName == "" {
        pathName = s.configManager.GetProjectPathName()
    }
    if pathName == "" {
        return mcp.NewToolResultText("未配置项目路径名称，请先使用 init_project 设置项目，或提供 pathName"), nil
    }

    // 获取项目信息
    projectURL := fmt.Sprintf("%s/projects/by-path/%s", s.getApiURL(), pathName)
    projectResp, err := http.Get(projectURL)
    if err != nil {
        return mcp.NewToolResultText("获取项目失败：" + err.Error()), nil
    }
    defer projectResp.Body.Close()

    if projectResp.StatusCode == 404 {
        return mcp.NewToolResultText(fmt.Sprintf("未找到项目: %s", pathName)), nil
    }

    var projectResult map[string]interface{}
    if err := json.NewDecoder(projectResp.Body).Decode(&projectResult); err != nil {
        return mcp.NewToolResultText("解析项目失败：" + err.Error()), nil
    }

    project, ok := projectResult["project"].(map[string]interface{})
    if !ok {
        return mcp.NewToolResultText("项目数据格式错误"), nil
    }

    projectId, _ := project["id"].(string)
    projectName, _ := project["name"].(string)

    // 使用 projectId 获取模块列表
    modulesURL := fmt.Sprintf("%s/modules?projectId=%s", s.getApiURL(), projectId)
    modulesResp, err := http.Get(modulesURL)
    if err != nil {
        return mcp.NewToolResultText("获取模块失败：" + err.Error()), nil
    }
    defer modulesResp.Body.Close()

    var modulesResult map[string]interface{}
    if err := json.NewDecoder(modulesResp.Body).Decode(&modulesResult); err != nil {
        return mcp.NewToolResultText("解析模块失败：" + err.Error()), nil
    }

    // 构建树形结构
    var sb strings.Builder
    sb.WriteString(fmt.Sprintf("%-35s # %s, project, -\n", pathName, projectName))

    // 处理模块 - API 返回 {"modules": [...]}
    if modules, ok := modulesResult["modules"].([]interface{}); ok {
        buildModuleTree(&sb, modules, "", s.getApiURL())
    }

    return mcp.NewToolResultText(sb.String()), nil
}

// buildModuleTree 递归构建模块树（修复版）
func buildModuleTree(sb *strings.Builder, modules []interface{}, prefix string, apiURL string) {
    total := len(modules)
    for i, module := range modules {
        moduleMap, ok := module.(map[string]interface{})
        if !ok {
            continue
        }

        isLast := i == total-1
        var connector, childPrefix string
        if isLast {
            connector = "└── "
            childPrefix = "    "
        } else {
            connector = "├── "
            childPrefix = "│   "
        }

        modulePathName, _ := moduleMap["pathName"].(string)
        moduleName, _ := moduleMap["name"].(string)
        moduleStatus, _ := moduleMap["status"].(string)
        if moduleStatus == "" {
            moduleStatus = "-"
        }

        sb.WriteString(fmt.Sprintf("%s%s%-30s # %s, module, %s\n", prefix, connector, modulePathName, moduleName, moduleStatus))

        // 获取该模块的任务
        tasksURL := fmt.Sprintf("%s/modules/by-path/%s?action=tasks", apiURL, modulePathName)
        tasksResp, err := http.Get(tasksURL)
        if err == nil {
            defer tasksResp.Body.Close()
            var tasksResult map[string]interface{}
            if err := json.NewDecoder(tasksResp.Body).Decode(&tasksResult); err == nil {
                // API 返回 {"tasks": [...]}
                if tasks, ok := tasksResult["tasks"].([]interface{}); ok {
                    for j, task := range tasks {
                        taskMap, ok := task.(map[string]interface{})
                        if !ok {
                            continue
                        }
                        taskIsLast := j == len(tasks)-1
                        var taskConnector string
                        if taskIsLast {
                            taskConnector = "└── "
                        } else {
                            taskConnector = "├── "
                        }

                        taskPathName, _ := taskMap["pathName"].(string)
                        taskName, _ := taskMap["name"].(string)
                        taskStatus, _ := taskMap["status"].(string)
                        if taskStatus == "" {
                            taskStatus = "ready"
                        }

                        sb.WriteString(fmt.Sprintf("%s%s%s%-30s # %s, task, %s\n", prefix, childPrefix, taskConnector, taskPathName, taskName, taskStatus))
                    }
                }
            }
        }
    }
}
```

#### 方案 B：修改 API 层

在 API 的 `GetModules` 函数中添加 `projectPathName` 参数支持。

**修改位置**：[`backend/internal/api/handlers/module.go`](backend/internal/api/handlers/module.go) 第 14-44 行

```go
// GetModules 获取模块列表
func GetModules(c *gin.Context) {
    var modules []models.Module

    query := database.DB.Model(&models.Module{})

    // 过滤条件 - 支持多个项目ID（逗号分隔）
    if projectIDs := c.Query("projectIds"); projectIDs != "" {
        query = query.Where("project_id IN ?", strings.Split(projectIDs, ","))
    } else if projectID := c.Query("projectId"); projectID != "" {
        query = query.Where("project_id = ?", projectID)
    } else if projectPathName := c.Query("projectPathName"); projectPathName != "" {
        // 新增：支持通过项目 pathName 查询
        var project models.Project
        if err := database.DB.First(&project, "path_name = ?", projectPathName).Error; err == nil {
            query = query.Where("project_id = ?", project.ID)
        }
    }
    // ... 其余代码不变
}
```

### 2.5 推荐方案

**推荐方案 A**（修改 MCP 层），原因：
1. 保持 API 层的简洁性
2. MCP 层已经有获取项目信息的能力
3. 修改范围更小，只涉及一个文件

### 2.6 需要修改的文件

| 文件 | 修改内容 |
|------|----------|
| `backend/internal/mcp/tools_get.go` | 修改 `handleQueryProjectIndexTreeImpl` 和 `buildModuleTree` 函数 |

---

## 3. `query_file_code_path_tree` 接口修复

### 3.1 问题描述

- **期望行为**：返回完整的代码文件路径树
- **当前行为**：只返回一行内容或空内容

### 3.2 期望输出格式

**查询模块**：
```
aitdd/auth                         # 认证模块, module
├── src/auth/login.ts              # 用户登录 (aitdd/auth/login)
├── src/auth/login.test.ts         # 用户登录 (aitdd/auth/login)
├── src/auth/register.ts           # 用户注册 (aitdd/auth/register)
└── src/auth/logout.ts             # 用户登出 (aitdd/auth/logout)
totalFiles: 4
```

**查询单个任务**：
```
aitdd/auth/login                   # 用户登录, task
├── src/auth/login.ts              # 用户登录
├── src/auth/login.test.ts         # 用户登录
totalFiles: 2
modulePathName: aitdd/auth
```

### 3.3 问题根因分析

查看 [`handleQueryFileCodePathTreeImpl`](backend/internal/mcp/tools_get.go:392) 函数：

**问题**：
1. **任务列表获取问题**：与 `query_project_index_tree` 相同，API 返回 `{"tasks": [...]}` 而非 `{"data": [...]}`
2. **数据解析错误**：代码尝试从 `tasksResult["data"]` 获取任务列表

```go
// 当前代码（第 439 行）
if tasks, ok := tasksResult["data"].([]interface{}); ok {
```

### 3.4 修复方案

**修改位置**：[`backend/internal/mcp/tools_get.go`](backend/internal/mcp/tools_get.go) 第 392-499 行

**修复后代码**：

```go
// handleQueryFileCodePathTreeImpl 查询代码文件路径树实现
func (s *MCPServer) handleQueryFileCodePathTreeImpl(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
    pathName, ok := getParam(request, "pathName")
    if !ok || pathName == "" {
        return mcp.NewToolResultText("缺少 pathName 参数"), nil
    }

    nodeType := s.detectNodeType(pathName)

    var sb strings.Builder
    var totalFiles int

    // 根据节点类型获取代码路径
    if nodeType == "module" {
        // 获取模块信息
        moduleURL := fmt.Sprintf("%s/modules/by-path/%s", s.getApiURL(), pathName)
        moduleResp, err := http.Get(moduleURL)
        if err != nil {
            return mcp.NewToolResultText("获取模块失败：" + err.Error()), nil
        }
        defer moduleResp.Body.Close()

        if moduleResp.StatusCode == 404 {
            return mcp.NewToolResultText(fmt.Sprintf("未找到模块: %s", pathName)), nil
        }

        var moduleResult map[string]interface{}
        if err := json.NewDecoder(moduleResp.Body).Decode(&moduleResult); err != nil {
            return mcp.NewToolResultText("解析模块失败：" + err.Error()), nil
        }

        module, ok := moduleResult["module"].(map[string]interface{})
        if !ok {
            return mcp.NewToolResultText("模块数据格式错误"), nil
        }

        moduleName, _ := module["name"].(string)
        sb.WriteString(fmt.Sprintf("%-35s # %s, module\n", pathName, moduleName))

        // 获取模块下所有任务的代码路径
        tasksURL := fmt.Sprintf("%s/modules/by-path/%s?action=tasks", s.getApiURL(), pathName)
        tasksResp, err := http.Get(tasksURL)
        if err != nil {
            return mcp.NewToolResultText("获取任务失败：" + err.Error()), nil
        }
        defer tasksResp.Body.Close()

        var tasksResult map[string]interface{}
        if err := json.NewDecoder(tasksResp.Body).Decode(&tasksResult); err != nil {
            return mcp.NewToolResultText("解析任务失败：" + err.Error()), nil
        }

        // 修复：API 返回 {"tasks": [...]}
        if tasks, ok := tasksResult["tasks"].([]interface{}); ok {
            for _, task := range tasks {
                taskMap, ok := task.(map[string]interface{})
                if !ok {
                    continue
                }
                taskPathName, _ := taskMap["pathName"].(string)
                taskName, _ := taskMap["name"].(string)
                codePaths, _ := taskMap["codePaths"].([]interface{})

                for _, codePath := range codePaths {
                    codePathStr, ok := codePath.(string)
                    if ok {
                        sb.WriteString(fmt.Sprintf("├── %-30s # %s (%s)\n", codePathStr, taskName, taskPathName))
                        totalFiles++
                    }
                }
            }
        }
    } else if nodeType == "task" {
        // 获取单个任务的代码路径
        taskURL := fmt.Sprintf("%s/tasks/by-path/%s", s.getApiURL(), pathName)
        taskResp, err := http.Get(taskURL)
        if err != nil {
            return mcp.NewToolResultText("获取任务失败：" + err.Error()), nil
        }
        defer taskResp.Body.Close()

        if taskResp.StatusCode == 404 {
            return mcp.NewToolResultText(fmt.Sprintf("未找到任务: %s", pathName)), nil
        }

        var taskResult map[string]interface{}
        if err := json.NewDecoder(taskResp.Body).Decode(&taskResult); err != nil {
            return mcp.NewToolResultText("解析任务失败：" + err.Error()), nil
        }

        task, ok := taskResult["task"].(map[string]interface{})
        if !ok {
            return mcp.NewToolResultText("任务数据格式错误"), nil
        }

        taskName, _ := task["name"].(string)
        modulePathName, _ := task["modulePathName"].(string)
        sb.WriteString(fmt.Sprintf("%-35s # %s, task\n", pathName, taskName))

        codePaths, _ := task["codePaths"].([]interface{})
        for _, codePath := range codePaths {
            codePathStr, ok := codePath.(string)
            if ok {
                sb.WriteString(fmt.Sprintf("├── %-30s # %s\n", codePathStr, taskName))
                totalFiles++
            }
        }

        // 显示所属模块信息
        if modulePathName != "" {
            sb.WriteString(fmt.Sprintf("modulePathName: %s\n", modulePathName))
        }
    } else {
        return mcp.NewToolResultText(fmt.Sprintf("不支持的节点类型: %s (仅支持 module 和 task)", nodeType)), nil
    }

    sb.WriteString(fmt.Sprintf("totalFiles: %d\n", totalFiles))
    return mcp.NewToolResultText(sb.String()), nil
}
```

### 3.5 需要修改的文件

| 文件 | 修改内容 |
|------|----------|
| `backend/internal/mcp/tools_get.go` | 修改 `handleQueryFileCodePathTreeImpl` 函数 |

---

## 4. 问题根因总结

### 4.1 数据结构不匹配

| API | 实际返回 | MCP 代码期望 |
|-----|----------|--------------|
| `GET /projects` | `{"projects": [...], "total": N}` | `{"data": [...]}` |
| `GET /modules` | `{"modules": [...], "total": N}` | `{"data": [...]}` |
| `GET /modules/:id?action=tasks` | `{"tasks": [...], "total": N}` | `{"data": [...]}` |
| `GET /modules/by-path/:pathName` | `{"module": {...}}` | `{...}` |
| `GET /tasks/by-path/:pathName` | `{"task": {...}}` | `{...}` |

### 4.2 查询参数不匹配

| MCP 调用 | API 支持 | 问题 |
|----------|----------|------|
| `?projectPathName=xxx` | `?projectId=xxx` | API 不支持 pathName 参数 |

---

## 5. 实施计划

### 5.1 修改顺序

1. **修复 `init_project` 接口**
   - 修改 `handleInitProjectImpl` 函数
   - 修正 API 返回数据解析
   - 调整输出格式为 YAML

2. **修复 `query_project_index_tree` 接口**
   - 修改 `handleQueryProjectIndexTreeImpl` 函数
   - 先获取项目 ID 再查询模块
   - 修正任务列表数据解析

3. **修复 `query_file_code_path_tree` 接口**
   - 修改 `handleQueryFileCodePathTreeImpl` 函数
   - 修正模块和任务数据解析

### 5.2 测试验证

修复后需要验证以下场景：

1. **`init_project` 测试**
   - 无参数调用，验证返回项目列表格式
   - 传入 `pathName`，验证设置成功

2. **`query_project_index_tree` 测试**
   - 验证返回完整的树形结构
   - 验证模块和任务都正确显示

3. **`query_file_code_path_tree` 测试**
   - 查询模块，验证返回所有任务的代码路径
   - 查询任务，验证返回该任务的代码路径

---

## 6. 代码变更汇总

| 文件 | 函数 | 变更类型 |
|------|------|----------|
| `backend/internal/mcp/tools_get.go` | `handleInitProjectImpl` | 修改 |
| `backend/internal/mcp/tools_get.go` | `handleQueryProjectIndexTreeImpl` | 修改 |
| `backend/internal/mcp/tools_get.go` | `buildModuleTree` | 修改 |
| `backend/internal/mcp/tools_get.go` | `handleQueryFileCodePathTreeImpl` | 修改 |

---

## 7. 附录：API 返回数据结构参考

### 7.1 项目相关

```json
// GET /api/v1/projects
{
  "projects": [
    {
      "id": "xxx",
      "name": "AITDD",
      "pathName": "aitdd",
      "constitution": "项目描述...",
      "createdAt": 1234567890,
      "updatedAt": 1234567890,
      "version": 1
    }
  ],
  "total": 1
}

// GET /api/v1/projects/by-path/:pathName
{
  "project": {
    "id": "xxx",
    "name": "AITDD",
    "pathName": "aitdd",
    ...
  }
}
```

### 7.2 模块相关

```json
// GET /api/v1/modules?projectId=xxx
{
  "modules": [
    {
      "id": "xxx",
      "projectId": "xxx",
      "name": "认证模块",
      "pathName": "aitdd/auth",
      "status": "developing",
      ...
    }
  ],
  "total": 1
}

// GET /api/v1/modules/by-path/:pathName
{
  "module": {
    "id": "xxx",
    "name": "认证模块",
    ...
  }
}

// GET /api/v1/modules/by-path/:pathName?action=tasks
{
  "tasks": [
    {
      "id": "xxx",
      "moduleId": "xxx",
      "name": "用户登录",
      "pathName": "aitdd/auth/login",
      "status": "completed",
      "codePaths": ["src/auth/login.ts", "src/auth/login.test.ts"],
      ...
    }
  ],
  "total": 1
}
```

### 7.3 任务相关

```json
// GET /api/v1/tasks/by-path/:pathName
{
  "task": {
    "id": "xxx",
    "moduleId": "xxx",
    "name": "用户登录",
    "pathName": "aitdd/auth/login",
    "modulePathName": "aitdd/auth",
    "status": "completed",
    "codePaths": ["src/auth/login.ts", "src/auth/login.test.ts"],
    ...
  }
}
```
