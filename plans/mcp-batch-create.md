# MCP 批量创建接口设计方案

## 1. 概述

本文档说明如何将现有的 `create_node` 函数重构为三个独立的批量创建函数：

- `create_module` - 批量创建模块
- `create_task` - 批量创建任务
- `create_project` - 批量创建项目

### 核心改进点

#### 统一批量创建模式
- 使用 `operations` 数组参数，每项包含创建所需的字段
- **每个操作独立执行**：创建操作不需要原子性保证（新资源之间无版本冲突）
- **部分成功支持**：返回每个操作的独立结果（成功/失败）
- 详细的错误信息（指明哪个操作失败、失败原因）

#### 与现有批量接口的一致性
- 参数格式与 `modify_module`/`modify_task` 的 `operations` 数组保持一致
- 输出格式参考批量修改的 YAML 风格
- 便于 AI 理解和调用

---

## 2. 当前接口分析

### 2.1 create_node

**当前定义** ([`server.go:162-169`](backend/internal/mcp/server.go:162)):
```go
s.server.AddTool(mcp.NewTool("create_node",
    mcp.WithDescription("统一创建节点接口。根据 type 创建模块或任务。type='module' 时创建模块，type='task' 时创建任务。"),
    mcp.WithString("type", mcp.Description("节点类型：module 或 task"), mcp.Required()),
    mcp.WithString("parentPath", mcp.Description("父节点 pathName。创建模块时为项目或父模块 pathName，创建任务时为所属模块 pathName"), mcp.Required()),
    mcp.WithString("name", mcp.Description("节点名称"), mcp.Required()),
    mcp.WithString("pathName", mcp.Description("节点 pathName（可选，不传则根据 parentPath 和 name 自动生成）")),
    mcp.WithObject("data", mcp.Description("类型相关的具体字段。module: description, prompt, upstreamContractSummary, downstreamContractSummary；task: description, prompt, upstreamContractDetail, downstreamContractDetail, tests, codePaths")),
), s.handleCreateNode)
```

**当前实现** ([`tools_modify.go:24-57`](backend/internal/mcp/tools_modify.go:24)):
- 接收单个创建请求
- 根据 `type` 参数分发到 `createModuleFromNode` 或 `createTaskFromNode`
- 调用 POST API `/modules` 或 `/tasks`
- 返回单个创建结果

**问题分析**:
1. **单次操作限制**：每次只能创建一个节点，批量创建需要多次调用
2. **类型混合**：module 和 task 使用同一接口，参数验证复杂
3. **缺少项目创建**：无法通过 MCP 创建项目
4. **返回信息不足**：失败时错误信息不够详细

---

## 3. 新接口设计

### 3.1 create_module - 批量创建模块

#### 参数设计

```json
{
  "operations": [
    {
      "parentPath": "project-a",
      "name": "认证模块",
      "pathName": "project-a/module-auth",
      "data": {
        "description": "用户认证和授权模块",
        "prompt": "实现JWT认证...",
        "upstreamContractSummary": "无上游依赖",
        "downstreamContractSummary": "提供登录/登出接口"
      }
    },
    {
      "parentPath": "project-a/module-auth",
      "name": "OAuth模块",
      "data": {
        "description": "第三方OAuth认证"
      }
    }
  ]
}
```

| 参数名 | 类型 | 必填 | 说明 |
|--------|------|------|------|
| operations | object[] | 是 | 批量操作数组 |
| operations[].parentPath | string | 是 | 父节点路径（项目路径或父模块路径） |
| operations[].name | string | 是 | 模块名称 |
| operations[].pathName | string | 否 | 指定 pathName，不传则自动生成（parentPath + "/" + name 的 slug 形式） |
| operations[].data | object | 否 | 模块属性字段 |
| operations[].data.description | string | 否 | 模块描述 |
| operations[].data.prompt | string | 否 | 模块提示词 |
| operations[].data.upstreamContractSummary | string | 否 | 上游契约摘要 |
| operations[].data.downstreamContractSummary | string | 否 | 下游契约摘要 |
| operations[].data.status | string | 否 | 初始状态，默认 "designing" |

#### 输出格式设计

**全部成功情况**：
```yaml
success: true
summary: "全部 2 个模块创建成功"
results:
  - pathName: "project-a/module-auth"
    name: "认证模块"
    status: "created"
  - pathName: "project-a/module-auth/oauth"
    name: "OAuth模块"
    status: "created"
```

**部分成功情况**：
```yaml
success: partial
summary: "2 个操作中 1 个成功，1 个失败"
results:
  - pathName: "project-a/module-auth"
    name: "认证模块"
    status: "created"
  - pathName: "project-a/module-auth/oauth"
    name: "OAuth模块"
    status: "failed"
    error: "父模块不存在: project-a/module-auth"
```

**全部失败情况**：
```yaml
success: false
summary: "全部 2 个操作失败"
results:
  - pathName: "project-a/module-auth"
    name: "认证模块"
    status: "failed"
    error: "项目不存在: project-a"
  - pathName: "project-a/module-api"
    name: "API模块"
    status: "failed"
    error: "项目不存在: project-a"
```

### 3.2 create_task - 批量创建任务

#### 参数设计

```json
{
  "operations": [
    {
      "parentPath": "project-a/module-auth",
      "name": "登录功能",
      "pathName": "project-a/module-auth/task-login",
      "data": {
        "description": "实现用户登录功能",
        "prompt": "创建登录表单和验证逻辑...",
        "upstreamContractDetail": {
          "title": "上游契约",
          "list": [
            {"label": "用户数据", "contract_api": "UserService.getUser()", "from": "module-user"}
          ]
        },
        "downstreamContractDetail": {
          "title": "下游契约",
          "list": [
            {"label": "登录接口", "contract_api": "POST /api/login", "from": "self"}
          ]
        },
        "tests": [
          {"target": "验证登录表单", "api": "test_login_form()"},
          {"target": "验证JWT生成", "api": "test_jwt_generation()"}
        ],
        "codePaths": ["src/auth/login.ts"]
      }
    }
  ]
}
```

| 参数名 | 类型 | 必填 | 说明 |
|--------|------|------|------|
| operations | object[] | 是 | 批量操作数组 |
| operations[].parentPath | string | 是 | 所属模块路径 |
| operations[].name | string | 是 | 任务名称 |
| operations[].pathName | string | 否 | 指定 pathName，不传则自动生成 |
| operations[].data | object | 否 | 任务属性字段 |
| operations[].data.description | string | 否 | 任务描述 |
| operations[].data.prompt | string | 否 | 任务提示词 |
| operations[].data.upstreamContractDetail | object | 否 | 上游契约详情 |
| operations[].data.downstreamContractDetail | object | 否 | 下游契约详情 |
| operations[].data.tests | array | 否 | 测试用例列表 |
| operations[].data.codePaths | string[] | 否 | 代码文件路径 |
| operations[].data.status | string | 否 | 初始状态，默认 "ready" |

#### upstreamContractDetail / downstreamContractDetail 结构

```json
{
  "title": "契约标题",
  "list": [
    {
      "label": "契约项名称",
      "contract_api": "API接口或函数签名",
      "from": "来源模块/任务"
    }
  ]
}
```

#### tests 结构

```json
[
  {
    "target": "测试目标描述",
    "api": "测试函数或API"
  }
]
```

#### 输出格式设计

**全部成功情况**：
```yaml
success: true
summary: "全部 2 个任务创建成功"
results:
  - pathName: "project-a/module-auth/task-login"
    name: "登录功能"
    status: "created"
  - pathName: "project-a/module-auth/task-logout"
    name: "登出功能"
    status: "created"
```

**部分成功情况**：
```yaml
success: partial
summary: "2 个操作中 1 个成功，1 个失败"
results:
  - pathName: "project-a/module-auth/task-login"
    name: "登录功能"
    status: "created"
  - pathName: "project-a/module-auth/task-logout"
    name: "登出功能"
    status: "failed"
    error: "pathName 已存在: project-a/module-auth/task-logout"
```

### 3.3 create_project - 批量创建项目

#### 参数设计

```json
{
  "operations": [
    {
      "name": "电商系统",
      "pathName": "ecommerce-system",
      "data": {
        "description": "完整的电商平台后端服务",
        "repository": "https://github.com/org/ecommerce"
      }
    },
    {
      "name": "支付网关",
      "pathName": "payment-gateway",
      "data": {
        "description": "第三方支付集成服务"
      }
    }
  ]
}
```

| 参数名 | 类型 | 必填 | 说明 |
|--------|------|------|------|
| operations | object[] | 是 | 批量操作数组 |
| operations[].name | string | 是 | 项目名称 |
| operations[].pathName | string | 是 | 项目唯一标识路径（不含斜杠） |
| operations[].data | object | 否 | 项目属性字段 |
| operations[].data.description | string | 否 | 项目描述 |
| operations[].data.repository | string | 否 | 代码仓库地址 |

#### 输出格式设计

**全部成功情况**：
```yaml
success: true
summary: "全部 2 个项目创建成功"
results:
  - pathName: "ecommerce-system"
    name: "电商系统"
    status: "created"
  - pathName: "payment-gateway"
    name: "支付网关"
    status: "created"
```

**部分成功情况**：
```yaml
success: partial
summary: "2 个操作中 1 个成功，1 个失败"
results:
  - pathName: "ecommerce-system"
    name: "电商系统"
    status: "created"
  - pathName: "payment-gateway"
    name: "支付网关"
    status: "failed"
    error: "pathName 已存在: payment-gateway"
```

---

## 4. MCP 注册对比

### 4.1 create_module MCP 注册对比

**原注册（create_node，现在）**：
```go
// 位置: server.go:162-169
s.server.AddTool(mcp.NewTool("create_node",
    mcp.WithDescription("统一创建节点接口。根据 type 创建模块或任务。type='module' 时创建模块，type='task' 时创建任务。"),
    mcp.WithString("type", mcp.Description("节点类型：module 或 task"), mcp.Required()),
    mcp.WithString("parentPath", mcp.Description("父节点 pathName。创建模块时为项目或父模块 pathName，创建任务时为所属模块 pathName"), mcp.Required()),
    mcp.WithString("name", mcp.Description("节点名称"), mcp.Required()),
    mcp.WithString("pathName", mcp.Description("节点 pathName（可选，不传则根据 parentPath 和 name 自动生成）")),
    mcp.WithObject("data", mcp.Description("类型相关的具体字段。module: description, prompt, upstreamContractSummary, downstreamContractSummary；task: description, prompt, upstreamContractDetail, downstreamContractDetail, tests, codePaths")),
), s.handleCreateNode)
```

**新注册（create_module，改后）**：
```go
// 位置: server.go registerNodeTools() 中
s.server.AddTool(mcp.NewTool("create_module",
    mcp.WithDescription(`批量创建模块。使用 operations 数组，每项包含 {parentPath, name, pathName?, data?}。

支持部分成功：每个操作独立执行，返回各自的创建结果。
pathName 不传则自动生成（parentPath + "/" + name 的 slug 形式）。

data 可选字段:
- description(string): 模块描述
- prompt(string): 模块提示词
- upstreamContractSummary(string): 上游契约摘要
- downstreamContractSummary(string): 下游契约摘要
- status(string): 初始状态，默认 designing`),
    mcp.WithArray("operations", mcp.Description("批量操作数组，每项: {parentPath: string, name: string, pathName?: string, data?: object}"), mcp.Required()),
), s.handleCreateModule)
```

**主要变化**：
- ✅ 独立的 `create_module` 函数，移除 `type` 参数
- ✅ 使用 `operations` 数组支持批量创建
- ✅ `operations` 标记为 Required
- ✅ 描述中详细说明 data 字段

### 4.2 create_task MCP 注册对比

**原注册（create_node，现在）**：
```go
// 同上，使用 create_node
```

**新注册（create_task，改后）**：
```go
// 位置: server.go registerNodeTools() 中
s.server.AddTool(mcp.NewTool("create_task",
    mcp.WithDescription(`批量创建任务。使用 operations 数组，每项包含 {parentPath, name, pathName?, data?}。

支持部分成功：每个操作独立执行，返回各自的创建结果。
parentPath 必须是模块路径（包含一个斜杠）。

data 可选字段:
- description(string): 任务描述
- prompt(string): 任务提示词
- upstreamContractDetail(object): {title, list:[{label,contract_api,from}]}
- downstreamContractDetail(object): {title, list:[{label,contract_api,from}]}
- tests(array): [{target:string, api:string}]
- codePaths(string[]): 代码文件路径列表
- status(string): 初始状态，默认 ready`),
    mcp.WithArray("operations", mcp.Description("批量操作数组，每项: {parentPath: string, name: string, pathName?: string, data?: object}"), mcp.Required()),
), s.handleCreateTask)
```

**主要变化**：
- ✅ 独立的 `create_task` 函数
- ✅ 使用 `operations` 数组支持批量创建
- ✅ 描述中详细说明复杂字段结构

### 4.3 create_project MCP 注册（新增）

**新注册**：
```go
// 位置: server.go registerNodeTools() 中
s.server.AddTool(mcp.NewTool("create_project",
    mcp.WithDescription(`批量创建项目。使用 operations 数组，每项包含 {name, pathName, data?}。

支持部分成功：每个操作独立执行，返回各自的创建结果。
pathName 是项目的唯一标识，不含斜杠。

data 可选字段:
- description(string): 项目描述
- repository(string): 代码仓库地址`),
    mcp.WithArray("operations", mcp.Description("批量操作数组，每项: {name: string, pathName: string, data?: object}"), mcp.Required()),
), s.handleCreateProject)
```

---

## 5. 实现要点

### 5.1 创建操作的非原子性设计

与 `modify_module`/`modify_task` 不同，创建操作**不需要原子性保证**：

| 特性 | 修改操作 | 创建操作 |
|------|---------|---------|
| 原子性 | 需要（版本冲突检查） | 不需要（新资源无冲突） |
| 部分成功 | 不支持（全成功或全失败） | 支持（独立执行） |
| 错误处理 | 返回冲突详情 | 返回每个操作的错误 |
| 事务 | 使用数据库事务 | 逐个执行，无事务 |

### 5.2 实现流程

```
1. 参数解析
   ├── 解析 operations 数组
   └── 验证每项的必填字段

2. 逐个执行创建操作
   ├── 遍历 operations
   ├── 调用对应的 API 创建资源
   ├── 记录每个操作的结果（成功/失败）
   └── 继续执行下一个操作（即使前一个失败）

3. 汇总结果
   ├── 统计成功/失败数量
   ├── 生成 summary
   └── 返回 YAML 格式结果
```

### 5.3 pathName 自动生成规则

当 `pathName` 未提供时，按以下规则自动生成：

```
pathName = parentPath + "/" + slug(name)
```

其中 `slug(name)` 的转换规则：
- 转换为小写
- 空格替换为下划线
- 中文保持原样（或转换为拼音，根据实际需求）

示例：
- `name: "认证模块"` → `slug: "认证模块"` → `pathName: "project-a/认证模块"`
- `name: "User Auth"` → `slug: "user_auth"` → `pathName: "project-a/user_auth"`

### 5.4 错误处理

**常见错误类型**：

| 错误类型 | 错误信息示例 |
|---------|-------------|
| 父节点不存在 | `父模块不存在: project-a/module-auth` |
| pathName 冲突 | `pathName 已存在: project-a/module-auth` |
| 参数验证失败 | `缺少必填字段: name` |
| API 调用失败 | `创建失败: 数据库连接错误` |

### 5.5 代码结构建议

```go
// tools_modify.go

// handleCreateModuleImpl 批量创建模块
func (s *MCPServer) handleCreateModuleImpl(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
    // 1. 解析 operations 数组
    operations, err := parseCreateOperations(request, "module")
    if err != nil {
        return mcp.NewToolResultText(err.Error()), nil
    }
    
    // 2. 逐个执行创建操作
    results := make([]CreateResult, len(operations))
    for i, op := range operations {
        results[i] = s.createSingleModule(op)
    }
    
    // 3. 汇总结果
    return formatCreateResults(results, "模块"), nil
}

// CreateResult 创建操作结果
type CreateResult struct {
    PathName string `json:"pathName"`
    Name     string `json:"name"`
    Status   string `json:"status"` // "created" 或 "failed"
    Error    string `json:"error,omitempty"`
}

// formatCreateResults 格式化创建结果
func formatCreateResults(results []CreateResult, resourceType string) *mcp.CallToolResult {
    successCount := 0
    failedCount := 0
    for _, r := range results {
        if r.Status == "created" {
            successCount++
        } else {
            failedCount++
        }
    }
    
    var output strings.Builder
    total := len(results)
    
    if failedCount == 0 {
        output.WriteString(fmt.Sprintf("success: true\n"))
        output.WriteString(fmt.Sprintf("summary: \"全部 %d 个%s创建成功\"\n", total, resourceType))
    } else if successCount == 0 {
        output.WriteString(fmt.Sprintf("success: false\n"))
        output.WriteString(fmt.Sprintf("summary: \"全部 %d 个操作失败\"\n", total))
    } else {
        output.WriteString(fmt.Sprintf("success: partial\n"))
        output.WriteString(fmt.Sprintf("summary: \"%d 个操作中 %d 个成功，%d 个失败\"\n", total, successCount, failedCount))
    }
    
    output.WriteString("results:\n")
    for _, r := range results {
        output.WriteString(fmt.Sprintf("  - pathName: \"%s\"\n", r.PathName))
        output.WriteString(fmt.Sprintf("    name: \"%s\"\n", r.Name))
        output.WriteString(fmt.Sprintf("    status: \"%s\"\n", r.Status))
        if r.Error != "" {
            output.WriteString(fmt.Sprintf("    error: \"%s\"\n", r.Error))
        }
    }
    
    return mcp.NewToolResultText(output.String())
}
```

---

## 6. 测试要点

### 6.1 create_module 测试场景

| 场景 | 输入 | 预期输出 |
|------|------|---------|
| 单个创建成功 | operations: [{parentPath, name}] | success: true, 1 个 created |
| 批量创建成功 | operations: [{...}, {...}] | success: true, 2 个 created |
| 部分成功 | 其中一个 parentPath 不存在 | success: partial, 1 created, 1 failed |
| 全部失败 | 所有 parentPath 不存在 | success: false, 2 failed |
| pathName 冲突 | pathName 已存在 | success: false, error: "pathName 已存在" |
| 自动生成 pathName | 不传 pathName | 自动生成 parentPath/name_slug |
| 空数组 | operations: [] | 返回错误提示 |

### 6.2 create_task 测试场景

| 场景 | 输入 | 预期输出 |
|------|------|---------|
| 单个创建成功 | operations: [{parentPath, name}] | success: true, 1 个 created |
| 批量创建成功 | operations: [{...}, {...}] | success: true, 2 个 created |
| 复杂字段 | 包含 upstreamContractDetail, tests | 正确解析并创建 |
| parentPath 不是模块 | parentPath 是项目路径 | 返回错误 |
| 空数组 | operations: [] | 返回错误提示 |

### 6.3 create_project 测试场景

| 场景 | 输入 | 预期输出 |
|------|------|---------|
| 单个创建成功 | operations: [{name, pathName}] | success: true, 1 个 created |
| 批量创建成功 | operations: [{...}, {...}] | success: true, 2 个 created |
| pathName 冲突 | pathName 已存在 | success: false, error |
| pathName 包含斜杠 | pathName: "a/b" | 返回错误 |
| 空数组 | operations: [] | 返回错误提示 |

### 6.4 关键测试用例

#### 部分成功测试示例
```
场景：批量创建3个模块，其中1个父模块不存在
输入：operations=[
  {parentPath: "project-a", name: "模块1"},
  {parentPath: "project-not-exist", name: "模块2"},
  {parentPath: "project-a", name: "模块3"}
]
预期：
- success: partial
- summary: "3 个操作中 2 个成功，1 个失败"
- results 中模块1和模块3 status="created"
- results 中模块2 status="failed"，包含错误信息
```

#### 复杂字段测试示例
```
场景：创建包含完整契约信息的任务
输入：operations=[{
  parentPath: "project-a/module-auth",
  name: "登录功能",
  data: {
    upstreamContractDetail: {
      title: "上游契约",
      list: [{label: "用户数据", contract_api: "UserService.getUser()", from: "module-user"}]
    },
    tests: [{target: "验证登录", api: "test_login()"}]
  }
}]
预期：
- success: true
- 任务正确创建
- 契约和测试信息正确保存
```

---

## 7. 迁移计划

### 7.1 兼容性考虑

**保留 create_node（可选）**：
- 短期内保留 `create_node` 作为兼容接口
- 在描述中添加废弃提示
- 长期计划中移除

**或直接替换**：
- 移除 `create_node`
- 只提供 `create_module`、`create_task`、`create_project`
- 更新文档和示例

### 7.2 实施步骤

1. **阶段1：添加新函数**
   - 在 `server.go` 中注册 `create_module`、`create_task`、`create_project`
   - 在 `tools_modify.go` 中实现三个函数

2. **阶段2：测试验证**
   - 编写单元测试
   - 进行集成测试
   - 验证各种边界情况

3. **阶段3：文档更新**
   - 更新 `docs/mcp-guide.md`
   - 更新 `docs/mcp-tools-reference.md`

4. **阶段4：废弃旧接口（可选）**
   - 标记 `create_node` 为废弃
   - 设置移除时间表

---

## 8. 文件修改清单

| 文件 | 修改内容 |
|------|---------|
| `backend/internal/mcp/server.go` | 添加3个新函数的 MCP 注册，可选废弃 create_node |
| `backend/internal/mcp/tools_modify.go` | 实现 handleCreateModuleImpl、handleCreateTaskImpl、handleCreateProjectImpl |
| `backend/internal/mcp/tools_modify_test.go` | 新增批量创建测试用例 |
| `docs/mcp-guide.md` | 更新创建操作使用说明 |
| `docs/mcp-tools-reference.md` | 添加新接口文档 |

---

## 9. 总结

本设计方案将现有的 `create_node` 重构为三个独立的批量创建函数：

| 函数 | 用途 | 关键参数 |
|------|------|---------|
| `create_module` | 批量创建模块 | parentPath, name, pathName?, data? |
| `create_task` | 批量创建任务 | parentPath, name, pathName?, data? |
| `create_project` | 批量创建项目 | name, pathName, data? |

**核心设计原则**：
1. **批量操作**：使用 `operations` 数组支持批量创建
2. **部分成功**：每个操作独立执行，支持部分成功
3. **一致性**：与现有 `modify_*` 接口格式保持一致
4. **详细反馈**：返回每个操作的详细结果和错误信息
