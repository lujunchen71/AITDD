# MCP 接口路径验证方案设计

> 创建时间：2026-03-06  
> 状态：草稿

---

## 一、`compile_static` 当前路径处理分析

### 1.1 现有实现（已有基础保护）

[`handleCompileStaticImpl()`](../backend/internal/mcp/tools_other.go:1683) 已有路径验证逻辑：

```go
// tools_other.go:1684-1700
pathName, ok := getParam(request, "pathName")
currentProjectPathName := s.configManager.GetProjectPathName()
if !ok || pathName == "" {
    pathName = currentProjectPathName
}
if pathName == "" {
    return mcp.NewToolResultText("未配置项目路径名称，请先使用 init_project 设置项目"), nil
}

// 验证 pathName 是否属于当前项目
if currentProjectPathName != "" {
    pathParts := strings.Split(pathName, "/")
    if pathParts[0] != currentProjectPathName {
        return mcp.NewToolResultText(fmt.Sprintf("路径 [%s] 不属于当前项目 [%s]，...", pathName, currentProjectPathName)), nil
    }
}
```

### 1.2 `compile_static` 路径验证的现有问题

**问题1：项目名含空格时的 pathName 比较不严谨**

当前项目 `pathName` 为 `"Plants vs Zombies Web Game"`（含空格）。
验证逻辑是用 `pathParts[0] != currentProjectPathName`。
- 如果 `pathName = "Plants vs Zombies Web Game"` (项目级)，则 `pathParts[0] = "Plants"` ≠ `"Plants vs Zombies Web Game"` ✗
- **根本原因**：`strings.Split(pathName, "/")` 对于项目名含空格的情况，`pathParts[0]` 只取到第一个单词 "Plants"，而 `currentProjectPathName` 是完整的 "Plants vs Zombies Web Game"。

> **注意**：项目名（`pathName`）本身不含 `/`，是整个项目的顶层标识。但当项目名含有空格时，当前逻辑 `pathParts[0] != currentProjectPathName` 不会出错，因为 `strings.Split("Plants vs Zombies Web Game", "/")` 返回 `["Plants vs Zombies Web Game"]`，`pathParts[0]` 就是完整的项目名。此处原本是正确的。

**问题2：`compile_static` 参数在 `server.go` 中被标记为 Required，但实现中允许为空**

在 [`server.go:343`](../backend/internal/mcp/server.go:343)：
```go
mcp.WithString("pathName", mcp.Description("项目或模块 pathName"), mcp.Required()),
```
但实现中：
```go
if !ok || pathName == "" {
    pathName = currentProjectPathName  // 允许为空，使用当前项目默认值
}
```
这形成了设计矛盾：接口标记为 Required，但实现却 fallback 到默认值。行为上没有问题，但语义不一致。

**问题3：`compile_static` 的路径验证不严谨（潜在越界）**

```go
pathParts := strings.Split(pathName, "/")
if pathParts[0] != currentProjectPathName {
```
如果 `pathName = ""` 但此时 `currentProjectPathName` 也是空字符串，则 `pathParts[0] = ""` == `""` 通过验证，后续继续执行但路径为空，可能导致 API 调用出现问题。

但实际上，前面已经检查 `pathName == ""` 后 fallback，所以这个场景不会发生。

**结论：`compile_static` 路径验证基本正确，无严重 Bug。** 主要问题在于其他大量 MCP 接口**缺少路径归属验证**。

---

## 二、`project.json` 的 `pathName` 字段确认

### 2.1 字段位置

配置文件路径（通过 `findProjectConfigPath()` 向上查找）：`.aitdd/project.json`

当前测试环境内容（[`.aitdd/project.json`](../.aitdd/project.json)）：
```json
{
  "projectName": "Plants vs Zombies Web Game",
  "pathName": "Plants vs Zombies Web Game",
  "apiBaseUrl": "http://localhost:34567/api/v1"
}
```

### 2.2 `Config` 结构体定义

[`config.go:14-18`](../backend/internal/mcp/config.go:14)：
```go
type Config struct {
    ProjectName string `json:"projectName"`
    PathName    string `json:"pathName"`    // ← 当前项目的 pathName
    ApiBaseUrl  string `json:"apiBaseUrl"`
}
```

### 2.3 获取接口

- [`GetProjectPathName()`](../backend/internal/mcp/config.go:174)：返回 `config.PathName`
- [`IsConfigured()`](../backend/internal/mcp/config.go:200)：检查 `config.PathName != ""`

---

## 三、路径合法性验证函数设计

### 3.1 验证函数位置建议

**推荐：在 `config.go` 中添加 `ValidatePathBelongsToProject()` 方法**，或新建 `path_validator.go` 文件。

考虑到代码组织和可测试性，建议在 [`backend/internal/mcp/config.go`](../backend/internal/mcp/config.go) 中扩展 `ConfigManager`，或新建 `path_validator.go`。

### 3.2 验证函数签名设计

```go
// ValidatePathBelongsToProject 验证给定 pathName 是否属于当前项目
// 规则：pathName 的第一个路径段（"/"之前）必须等于 currentProjectPathName
// 特殊：若 currentProjectPathName 为空，则跳过验证（视为未配置项目）
//
// 返回值：
//   - "" 表示验证通过
//   - 非空字符串表示错误消息（可直接返回给调用方）
func (cm *ConfigManager) ValidatePathBelongsToProject(pathName string) string {
    cm.mu.RLock()
    defer cm.mu.RUnlock()
    
    currentProjectPathName := cm.config.PathName
    if currentProjectPathName == "" {
        return "" // 未配置项目，跳过验证
    }
    if pathName == "" {
        return "路径名称不能为空"
    }
    
    // 获取 pathName 的第一个段
    slashIdx := strings.Index(pathName, "/")
    var projectPart string
    if slashIdx == -1 {
        // 无斜杠，整个 pathName 就是项目名
        projectPart = pathName
    } else {
        projectPart = pathName[:slashIdx]
    }
    
    if projectPart != currentProjectPathName {
        return fmt.Sprintf(
            "路径 [%s] 不属于当前项目 [%s]，请先切换项目（使用 init_project）或使用正确的路径",
            pathName, currentProjectPathName,
        )
    }
    return ""
}
```

### 3.3 批量路径验证（用于 operations 数组场景）

```go
// ValidatePathsInOperations 验证操作列表中的所有路径
// fieldName 是 operation 中路径字段的名称（如 "pathName"、"parentPath"）
// 返回第一个不合法的路径和对应错误消息
func (cm *ConfigManager) ValidatePathsInOperations(
    operations []map[string]interface{},
    fieldNames ...string,
) string {
    for _, op := range operations {
        for _, fieldName := range fieldNames {
            if val, ok := op[fieldName].(string); ok && val != "" {
                if errMsg := cm.ValidatePathBelongsToProject(val); errMsg != "" {
                    return errMsg
                }
            }
        }
    }
    return ""
}
```

---

## 四、需要修改的 MCP 接口列表

### 4.1 分类说明

路径验证的核心逻辑：**pathName 的第一段（项目名部分）必须等于 `config.PathName`**。

验证时机分三类：
- **A 类（单路径参数）**：直接验证参数值
- **B 类（多路径参数）**：验证多个路径参数
- **C 类（operations 批量）**：遍历 operations 数组验证每个路径

### 4.2 写操作接口（必须验证）

| 工具名 | 文件 | 参数名 | 分类 | 当前状态 |
|--------|------|--------|------|----------|
| `create_module` | [`tools_modify.go:1041`](../backend/internal/mcp/tools_modify.go:1041) | operations[].parentPath | C | ❌ 无验证 |
| `create_task` | [`tools_modify.go:1242`](../backend/internal/mcp/tools_modify.go:1242) | operations[].parentPath | C | ❌ 无验证 |
| `modify_module` | [`tools_modify.go:461`](../backend/internal/mcp/tools_modify.go:461) | operations[].pathName | C | ❌ 无验证 |
| `modify_task` | [`tools_modify.go:737`](../backend/internal/mcp/tools_modify.go:737) | operations[].pathName | C | ❌ 无验证 |
| `delete_node` | [`tools_modify.go:355`](../backend/internal/mcp/tools_modify.go:355) | path | A | ❌ 无验证 |
| `create_dependency` | [`tools_other.go:637`](../backend/internal/mcp/tools_other.go:637) | upstreamPath, downstreamPath | B | ❌ 无验证 |
| `delete_dependency` | [`tools_other.go:805`](../backend/internal/mcp/tools_other.go:805) | upstreamPath, downstreamPath | B | ❌ 无验证 |
| `acquire_lock` | [`tools_other.go:1397`](../backend/internal/mcp/tools_other.go:1397) | pathName | A | ❌ 无验证 |
| `release_lock` | [`tools_other.go:1427`](../backend/internal/mcp/tools_other.go:1427) | pathName | A | ❌ 无验证 |
| `create_issue` | [`tools_other.go:1162`](../backend/internal/mcp/tools_other.go:1162) | fromTaskPathName, toTaskPathName | B | ❌ 无验证 |
| `reply_issue` | [`tools_other.go:1204`](../backend/internal/mcp/tools_other.go:1204) | fromTaskPathName, toTaskPathName | B | ❌ 无验证 |
| `resolve_issue` | [`tools_other.go:1244`](../backend/internal/mcp/tools_other.go:1244) | fromTaskPathName, toTaskPathName | B | ❌ 无验证 |
| `compile_static` | [`tools_other.go:1683`](../backend/internal/mcp/tools_other.go:1683) | pathName | A | ✅ 已有验证（可重构用公共函数） |
| `compile_dynamic` | [`tools_other.go:3033`](../backend/internal/mcp/tools_other.go:3033) | compilePathName | A | ❌ 无验证 |

### 4.3 读操作接口（建议验证，防止信息泄露）

| 工具名 | 文件 | 参数名 | 分类 | 优先级 |
|--------|------|--------|------|--------|
| `query_module` | [`tools_get.go:1162`](../backend/internal/mcp/tools_get.go:1162) | queries[].pathName | C | 中 |
| `query_task` | [`tools_get.go:1325`](../backend/internal/mcp/tools_get.go:1325) | queries[].pathName | C | 中 |
| `query_project_index_tree` | [`tools_get.go:547`](../backend/internal/mcp/tools_get.go:547) | pathName (可选) | A | 低 |
| `query_file_code_path_tree` | [`tools_get.go:708`](../backend/internal/mcp/tools_get.go:708) | pathName (可选) | A | 低 |
| `check_contract_alignment` | [`tools_other.go:22`](../backend/internal/mcp/tools_other.go:22) | upstreamPathName, downstreamPathName | B | 中 |
| `check_task_readiness` | [`tools_other.go:215`](../backend/internal/mcp/tools_other.go:215) | pathName | A | 中 |
| `check_dependencies` | [`tools_other.go:331`](../backend/internal/mcp/tools_other.go:331) | pathName | A | 中 |
| `query_dependencies` | [`tools_other.go:967`](../backend/internal/mcp/tools_other.go:967) | pathName | A | 中 |
| `query_issues` | [`tools_other.go:1282`](../backend/internal/mcp/tools_other.go:1282) | taskPathName (可选) | A | 低 |
| `get_status` | [`tools_other.go:3531`](../backend/internal/mcp/tools_other.go:3531) | pathName (可选) | A | 低 |

### 4.4 不需要验证的接口

| 工具名 | 原因 |
|--------|------|
| `init_project` | 本身用于设置项目，特殊接口 |
| `get_context` | 无路径参数 |
| `get_config` | 无路径参数 |
| `update_config` | 无路径参数 |
| `get_rule` | 无路径参数 |
| `update_rule` | 无路径参数 |
| `create_project` | 创建新项目，pathName 不含斜杠，允许创建任意项目 |

---

## 五、实现方案

### 5.1 整体策略

**推荐：验证失败返回友好文本提示，而非 error**

原因：与现有 `compile_static` 验证失败的处理方式保持一致，返回 `mcp.NewToolResultText("错误信息")` 让 AI 能读懂。

### 5.2 实现步骤

#### 步骤1：在 `config.go` 中添加验证方法

在 [`backend/internal/mcp/config.go`](../backend/internal/mcp/config.go) 添加 `ValidatePathBelongsToProject()` 方法（见 §3.2）。

#### 步骤2：修改写操作接口（高优先级）

**A 类接口修改示例**（以 `handleDeleteNodeImpl` 为例）：

```go
// tools_modify.go:355
func (s *MCPServer) handleDeleteNodeImpl(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
    path, ok := getParam(request, "path")
    if !ok || path == "" {
        return mcp.NewToolResultText("参数 path 不能为空"), nil
    }
    // 新增：路径归属验证
    if errMsg := s.configManager.ValidatePathBelongsToProject(path); errMsg != "" {
        return mcp.NewToolResultText(errMsg), nil
    }
    // ... 原有逻辑
}
```

**C 类接口修改示例**（以 `handleModifyTaskImpl` 为例）：

```go
// tools_modify.go:737
func (s *MCPServer) handleModifyTaskImpl(...) {
    // 获取 operations
    // ...
    
    // 新增：批量验证所有 pathName
    for _, op := range operations {
        if errMsg := s.configManager.ValidatePathBelongsToProject(op.PathName); errMsg != "" {
            return mcp.NewToolResultText(fmt.Sprintf("操作 [%s] 失败：%s", op.PathName, errMsg)), nil
        }
    }
    // ... 原有逻辑
}
```

**B 类接口修改示例**（以 `handleCreateDependencyImpl` 为例）：

```go
// tools_other.go:637
func (s *MCPServer) handleCreateDependencyImpl(...) {
    upstreamPath, _ := getParam(request, "upstreamPath")
    downstreamPath, _ := getParam(request, "downstreamPath")
    
    // 新增：验证两个路径均属于当前项目
    if errMsg := s.configManager.ValidatePathBelongsToProject(upstreamPath); errMsg != "" {
        return mcp.NewToolResultText("上游路径: " + errMsg), nil
    }
    if errMsg := s.configManager.ValidatePathBelongsToProject(downstreamPath); errMsg != "" {
        return mcp.NewToolResultText("下游路径: " + errMsg), nil
    }
    // ... 原有逻辑
}
```

#### 步骤3：将 `compile_static` 现有验证重构为使用公共函数

将 [`tools_other.go:1693-1700`](../backend/internal/mcp/tools_other.go:1693) 的手写验证替换为：

```go
// 新实现
if errMsg := s.configManager.ValidatePathBelongsToProject(pathName); errMsg != "" {
    return mcp.NewToolResultText(errMsg), nil
}
```

#### 步骤4：处理 `create_module` / `create_task` 的 parentPath 验证

`create_module` 和 `create_task` 中，`parentPath` 是关键参数：

```go
// tools_modify.go，handleCreateModuleImpl 的单个模块创建逻辑
func (s *MCPServer) createSingleModule(op CreateOperation) CreateResult {
    // 新增验证
    if errMsg := s.configManager.ValidatePathBelongsToProject(op.ParentPath); errMsg != "" {
        result.Status = "failed"
        result.Error = "parentPath " + errMsg
        return result
    }
    // ... 原有逻辑
}
```

### 5.3 `compile_static` 的 `pathName` 参数设计修正

建议将 [`server.go:343`](../backend/internal/mcp/server.go:343) 的 `Required()` 改为可选：

```go
// 修改前
mcp.WithString("pathName", mcp.Description("项目或模块 pathName"), mcp.Required()),

// 修改后（与实现行为一致）
mcp.WithString("pathName", mcp.Description("项目或模块 pathName，不传则使用当前项目")),
```

### 5.4 错误消息格式统一

所有路径验证失败统一使用以下格式：
```
路径 [<pathName>] 不属于当前项目 [<currentProjectPathName>]，
请先切换项目（使用 init_project）或使用正确的路径。
当前项目：<currentProjectPathName>
```

---

## 六、具体代码修改计划

### 6.1 新增代码

**文件**：[`backend/internal/mcp/config.go`](../backend/internal/mcp/config.go)

添加 `ValidatePathBelongsToProject()` 和 `ValidatePathsInOperations()` 两个方法（约 30 行）。

### 6.2 修改 `tools_other.go`

| 函数 | 修改位置 | 操作 |
|------|----------|------|
| `handleCompileStaticImpl` | 第 1693-1700 行 | 替换为 `ValidatePathBelongsToProject()` 调用 |
| `handleCheckContractAlignmentImpl` | 第 22 行后 | 新增对 upstreamPathName、downstreamPathName 的验证 |
| `handleCheckTaskReadinessImpl` | 第 216 行后 | 新增对 pathName 的验证 |
| `handleCheckDependenciesImpl` | 第 332 行后 | 新增对 pathName 的验证 |
| `handleCreateDependencyImpl` | 第 641 行后 | 新增对 upstreamPath、downstreamPath 的验证 |
| `handleDeleteDependencyImpl` | 第 808 行后 | 新增对 upstreamPath、downstreamPath 的验证 |
| `handleQueryDependenciesImpl` | 第 969 行后 | 新增对 pathName 的验证 |
| `handleCreateIssueImpl` | 第 1165 行后 | 新增对 fromTaskPathName、toTaskPathName 的验证 |
| `handleReplyIssueImpl` | 第 1207 行后 | 新增对 fromTaskPathName、toTaskPathName 的验证 |
| `handleResolveIssueImpl` | 第 1247 行后 | 新增对 fromTaskPathName、toTaskPathName 的验证 |
| `handleLockResourceImpl` | 第 1400 行后 | 新增对 pathName 的验证 |
| `handleUnlockResourceImpl` | 第 1430 行后 | 新增对 pathName 的验证 |
| `handleCompileDynamicImpl` | 第 3035 行后 | 新增对 compilePathName 的验证 |

### 6.3 修改 `tools_modify.go`

| 函数 | 修改位置 | 操作 |
|------|----------|------|
| `handleDeleteNodeImpl` | 第 357 行后 | 新增对 path 的验证 |
| `handleModifyModuleImpl` | 第 477 行后（解析 operations 后） | 新增批量验证 pathName |
| `handleModifyTaskImpl` | 第 753 行后（解析 operations 后） | 新增批量验证 pathName |
| `createSingleModule` | 第 1097 行后 | 新增对 op.ParentPath 的验证 |
| `createSingleTask` | 第 1300 行后 | 新增对 op.ParentPath 的验证 |

### 6.4 修改 `tools_get.go`（中低优先级）

| 函数 | 修改位置 | 操作 |
|------|----------|------|
| `handleQueryProjectIndexTreeImpl` | 第 548 行后 | 新增对非空 pathName 的验证 |
| `handleQueryFileCodePathTreeImpl` | 第 709 行后 | 新增对非空 pathName 的验证 |
| `handleQueryModuleImpl` | queries 解析后 | 新增批量验证 pathName |
| `handleQueryTaskImpl` | queries 解析后 | 新增批量验证 pathName |

---

## 七、数据流图

```mermaid
graph TD
    AI-->|调用 MCP 接口|MCPTool
    MCPTool-->|提取 pathName 参数|PathExtract
    PathExtract-->|调用验证|ValidatePathBelongsToProject
    ValidatePathBelongsToProject-->|读取|ConfigManager
    ConfigManager-->|返回 currentProjectPathName|ValidatePathBelongsToProject
    ValidatePathBelongsToProject-->|pathName[0] == currentProjectPathName?|DecisionNode
    DecisionNode-->|是|ProceedWithLogic
    DecisionNode-->|否|ReturnError
    ReturnError-->|返回提示文本|AI
    ProceedWithLogic-->|执行业务逻辑|APICall
```

---

## 八、测试计划

### 8.1 单元测试

在 [`backend/internal/mcp/config_test.go`](../backend/internal/mcp/config_test.go) 中为 `ValidatePathBelongsToProject()` 添加测试用例：

| 场景 | 输入 pathName | currentProjectPathName | 预期结果 |
|------|--------------|----------------------|----------|
| 正常项目级路径 | "MyProject" | "MyProject" | "" |
| 正常模块级路径 | "MyProject/模块A" | "MyProject" | "" |
| 正常任务级路径 | "MyProject/模块A/任务1" | "MyProject" | "" |
| 其他项目路径 | "OtherProject/模块B" | "MyProject" | 非空错误 |
| 未配置项目 | "AnyProject" | "" | "" |
| 含空格的项目名 | "Plants vs Zombies Web Game/模块" | "Plants vs Zombies Web Game" | "" |

### 8.2 集成测试

在 [`backend/internal/mcp/tools_other_test.go`](../backend/internal/mcp/tools_other_test.go) 中补充测试，验证 `compile_static` 在跨项目路径情况下的返回值。

---

## 九、实施优先级

| 优先级 | 接口 | 理由 |
|--------|------|------|
| P0（立即）| `compile_static` 代码重构 | 使用公共函数替换手写验证 |
| P0（立即）| `modify_module`, `modify_task` | 写操作，风险高 |
| P0（立即）| `delete_node` | 写操作，风险高 |
| P1（高）| `create_module`, `create_task` | 写操作 |
| P1（高）| `create_dependency`, `delete_dependency` | 写操作 |
| P1（高）| `acquire_lock`, `release_lock` | 写操作 |
| P2（中）| `create_issue`, `reply_issue`, `resolve_issue` | 写操作 |
| P2（中）| `compile_dynamic` | 写操作（影响代码执行） |
| P3（低）| `check_*`, `query_*` 读操作接口 | 防止信息泄露 |
