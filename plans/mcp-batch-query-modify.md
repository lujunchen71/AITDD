# MCP 函数批量查询/修改重构计划

## 1. 概述

本文档说明如何将以下 4 个 MCP 函数重构为支持批量操作：

- `query_module` - 批量查询模块，每个路径可指定不同字段
- `query_task` - 批量查询任务，每个路径可指定不同字段
- `modify_module` - 批量修改模块，原子性保证（全成功或全失败）
- `modify_task` - 批量修改任务，原子性保证（全成功或全失败）

### 核心改进点

#### 查询函数改进
- 使用 `queries` 数组参数，每项包含 `{pathName, fields?}`
- **每个路径可以指定不同的查询字段**（适配 AI 智能查询能力）
- 按路径层级组织输出（树形结构）

#### 修改函数改进  
- 使用 `operations` 数组参数，每项包含 `{pathName, version, data}`
- **原子性保证：预检所有版本号，任一失败则全部拒绝**
- 详细的版本冲突错误信息（指明哪个路径冲突、当前版本、提供版本）
- 使用数据库事务保证数据一致性

## 2. 当前接口分析

### 2.1 query_module

**当前定义** ([`server.go:99-107`](backend/internal/mcp/server.go:99)):
```go
s.server.AddTool(mcp.NewTool("query_module",
    mcp.WithDescription(`查询模块信息...`),
    mcp.WithString("pathName", mcp.Description("模块路径，格式: 项目名/模块名"), mcp.Required()),
    mcp.WithArray("fields", mcp.Description("指定返回字段，如 [\"name\",\"status\"]")),
), s.handleQueryModule)
```

**当前实现** ([`tools_get.go:1157-1219`](backend/internal/mcp/tools_get.go:1157)):
- 接收单个 `pathName` 字符串
- 调用 API `/modules/by-path/{pathName}`
- 返回单个模块的 YAML 格式信息

### 2.2 query_task

**当前定义** ([`server.go:109-118`](backend/internal/mcp/server.go:109)):
```go
s.server.AddTool(mcp.NewTool("query_task",
    mcp.WithDescription(`查询任务信息...`),
    mcp.WithString("pathName", mcp.Description("任务路径，格式: 项目名/模块名/任务名"), mcp.Required()),
    mcp.WithArray("fields", mcp.Description("指定返回字段，如 [\"name\",\"tests\"]")),
), s.handleQueryTask)
```

**当前实现** ([`tools_get.go:1243-1305`](backend/internal/mcp/tools_get.go:1243)):
- 接收单个 `pathName` 字符串
- 调用 API `/tasks/by-path/{pathName}`
- 返回单个任务的 YAML 格式信息

### 2.3 modify_module

**当前定义** ([`server.go:168-176`](backend/internal/mcp/server.go:168)):
```go
s.server.AddTool(mcp.NewTool("modify_module",
    mcp.WithDescription(`修改模块。只更新 data 中传入的字段...`),
    mcp.WithString("pathName", mcp.Description("模块路径，格式: 项目名/模块名"), mcp.Required()),
    mcp.WithNumber("version", mcp.Description("当前版本号，用于乐观锁"), mcp.Required()),
    mcp.WithObject("data", mcp.Description("要修改的字段..."), mcp.Required()),
), s.handleModifyModule)
```

**当前实现** ([`tools_modify.go:441-510`](backend/internal/mcp/tools_modify.go:441)):
- 接收单个 `pathName`、`version`、`data`
- 使用乐观锁机制（version）
- 调用 PUT API `/modules/by-path/{pathName}`
- 返回成功/失败及新版本号

### 2.4 modify_task

**当前定义** ([`server.go:178-192`](backend/internal/mcp/server.go:178)):
```go
s.server.AddTool(mcp.NewTool("modify_task",
    mcp.WithDescription(`修改任务。只更新 data 中传入的字段...`),
    mcp.WithString("pathName", mcp.Description("任务路径，格式: 项目名/模块名/任务名"), mcp.Required()),
    mcp.WithNumber("version", mcp.Description("当前版本号，用于乐观锁"), mcp.Required()),
    mcp.WithObject("data", mcp.Description("要修改的字段..."), mcp.Required()),
), s.handleModifyTask)
```

**当前实现** ([`tools_modify.go:532-634`](backend/internal/mcp/tools_modify.go:532)):
- 接收单个 `pathName`、`version`、`data`
- 使用乐观锁机制（version）
- 特殊处理 JSON 字段（tests, codePaths 等）
- 调用 PUT API `/tasks/by-path/{pathName}`
- 返回成功/失败及新版本号

---

## 3. 新接口设计

### 3.1 query_module 批量查询

#### 参数设计

```json
{
  "queries": [
    {"pathName": "项目/模块A", "fields": ["name", "status"]},
    {"pathName": "项目/模块B", "fields": ["description", "prompt", "version"]}
  ]
}
```

| 参数名 | 类型 | 必填 | 说明 |
|--------|------|------|------|
| queries | object[] | 是 | 查询数组，每个查询独立指定路径和字段 |
| queries[].pathName | string | 是 | 模块路径 |
| queries[].fields | string[] | 否 | 该路径查询的字段，不填返回所有字段 |

#### 输出格式设计

按路径层级组织，使用树形缩进：

```yaml
project-a:
  module-auth:
    pathName: "project-a/module-auth"
    name: "认证模块"
    status: "developing"
    version: 3
  module-api:
    pathName: "project-a/module-api"
    description: "提供RESTful API接口"
    prompt: "创建RESTful API服务..."
    version: 5
project-b:
  module-core:
    pathName: "project-b/module-core"
    name: "核心模块"
    status: "designing"
    version: 1
```

**说明**：不同模块可以返回不同的字段（如 module-auth 返回 name/status，module-api 返回 description/prompt）

### 3.2 query_task 批量查询

#### 参数设计

```json
{
  "queries": [
    {"pathName": "项目/模块/任务A", "fields": ["name", "status", "tests"]},
    {"pathName": "项目/模块/任务B", "fields": ["prompt", "codePaths", "testResult"]}
  ]
}
```

| 参数名 | 类型 | 必填 | 说明 |
|--------|------|------|------|
| queries | object[] | 是 | 查询数组，每个查询独立指定路径和字段 |
| queries[].pathName | string | 是 | 任务路径（必须包含两个斜杠） |
| queries[].fields | string[] | 否 | 该路径查询的字段，不填返回所有字段 |

**限制**: 只能传任务路径（包含两个斜杠），不支持传模块路径查询所有子任务

#### 输出格式设计

```yaml
project-a:
  module-auth:
    task-login:
      pathName: "project-a/module-auth/task-login"
      name: "登录功能"
      status: "in_progress"
      tests: 
        - {target: "验证登录表单", api: "test_login_form()"}
      version: 2
    task-logout:
      pathName: "project-a/module-auth/task-logout"
      prompt: "创建登出功能..."
      codePaths: ["src/auth/logout.ts"]
      testResult: ["登出测试通过"]
      version: 1
```

**说明**：不同任务可以返回不同的字段（如 task-login 返回 name/status/tests，task-logout 返回 prompt/codePaths/testResult）

### 3.3 modify_module 批量修改

#### 参数设计

```json
{
  "operations": [
    {"pathName": "项目/模块A", "version": 3, "data": {"status": "completed"}},
    {"pathName": "项目/模块B", "version": 5, "data": {"description": "新描述"}}
  ]
}
```

| 参数名 | 类型 | 必填 | 说明 |
|--------|------|------|------|
| operations | object[] | 是 | 批量操作数组 |
| operations[].pathName | string | 是 | 模块路径 |
| operations[].version | number | 是 | 当前版本号（乐观锁） |
| operations[].data | object | 是 | 要修改的字段 |

#### 原子性保证流程

1. **预检阶段**：先检查所有操作的版本号
   - 逐个查询每个路径的当前版本
   - 与提供的版本号对比
   - 收集所有版本冲突信息
   
2. **决策阶段**：
   - 如果**任何一个**版本冲突，拒绝整个批量操作
   - 如果全部版本正常，进入写入阶段
   
3. **写入阶段**：
   - 开启数据库事务
   - 逐个执行修改操作
   - 任一失败则回滚所有操作
   - 全部成功则提交事务

#### 输出格式设计

**成功情况**：
```yaml
success: true
summary: "全部 2 个模块修改成功"
results:
  - pathName: "project-a/module-auth"
    newVersion: 4
  - pathName: "project-a/module-api"
    newVersion: 6
```

**版本冲突情况**（原子性保证，全部拒绝）：
```yaml
success: false
error: VERSION_CONFLICT
message: "批量修改已取消：发现 1 个版本冲突"
conflicts:
  - pathName: "project-a/module-api"
    providedVersion: 5
    currentVersion: 6
    message: "版本落后，请获取新版本后重试"
checks:
  - pathName: "project-a/module-auth"
    providedVersion: 3
    currentVersion: 3
    status: "版本正常"
  - pathName: "project-a/module-api"
    providedVersion: 5
    currentVersion: 6
    status: "版本冲突"
```

**说明**：即使只有一个版本冲突，整个批量操作也会被拒绝，保证数据一致性

### 3.4 modify_task 批量修改

#### 参数设计

```json
{
  "operations": [
    {"pathName": "项目/模块/任务A", "version": 2, "data": {"status": "completed"}},
    {"pathName": "项目/模块/任务B", "version": 1, "data": {"tests": [{"target": "验证XX", "api": "test_xx()"}]}}
  ]
}
```

| 参数名 | 类型 | 必填 | 说明 |
|--------|------|------|------|
| operations | object[] | 是 | 批量操作数组 |
| operations[].pathName | string | 是 | 任务路径 |
| operations[].version | number | 是 | 当前版本号（乐观锁） |
| operations[].data | object | 是 | 要修改的字段 |

#### 原子性保证流程

与 `modify_module` 相同的三阶段流程：
1. **预检阶段**：检查所有版本号
2. **决策阶段**：任一冲突则全部拒绝
3. **写入阶段**：事务保证原子性

#### 输出格式设计

**成功情况**：
```yaml
success: true
summary: "全部 2 个任务修改成功"
results:
  - pathName: "project-a/module-auth/task-login"
    newVersion: 3
  - pathName: "project-a/module-auth/task-logout"
    newVersion: 2
```

**版本冲突情况**（原子性保证，全部拒绝）：
```yaml
success: false
error: VERSION_CONFLICT
message: "批量修改已取消：发现 1 个版本冲突"
conflicts:
  - pathName: "project-a/module-auth/task-login"
    providedVersion: 2
    currentVersion: 3
    message: "版本落后，请获取新版本后重试"
checks:
  - pathName: "project-a/module-auth/task-login"
    providedVersion: 2
    currentVersion: 3
    status: "版本冲突"
  - pathName: "project-a/module-auth/task-logout"
    providedVersion: 1
    currentVersion: 1
    status: "版本正常"
```

---

## 4. 实现要点（概要说明）

### 4.1 MCP 注册修改 ([`server.go`](backend/internal/mcp/server.go))

#### query_module / query_task 注册要点

- 使用 `queries` 数组参数，每项包含 `{pathName, fields?}`
- **每个路径可以指定不同字段**
- `queries` 标记为 Required
- 删除原有的 `pathName` 和 `fields` 参数
- 更新 MCP 描述，说明只支持批量查询模式

#### modify_module / modify_task 注册要点

- 使用 `operations` 数组参数，每项包含 `{pathName, version, data}`
- `operations` 标记为 Required
- 删除原有的 `pathName`, `version`, `data` 参数
- 更新 MCP 描述，强调原子性保证和版本预检机制

### 4.2 查询函数实现要点 ([`tools_get.go`](backend/internal/mcp/tools_get.go))

#### handleQueryModuleImpl / handleQueryTaskImpl 重构

**主要逻辑**：
1. **参数解析**：
   - 解析 `queries` 数组参数
   - 每项提取 `pathName` 和 `fields`
   
2. **批量查询**：
   - 遍历 queries 数组
   - 根据每个查询的 pathName 和 fields 调用 API
   - 收集所有结果和错误信息
   
3. **构建路径树**：
   - 将查询结果按路径层级组织成树形结构
   - 每个路径保存其对应的字段数据
   
4. **输出渲染**：
   - 按树形结构输出 YAML 格式
   - 每个节点只输出其对应的字段
   - 附加错误信息列表

### 4.3 修改函数实现要点 ([`tools_modify.go`](backend/internal/mcp/tools_modify.go))

#### handleModifyModuleImpl / handleModifyTaskImpl 重构

**三阶段原子性保证**：

**阶段1：预检所有版本号**
- 逐个查询每个路径的当前版本
- 与提供的版本号进行对比
- 收集所有版本冲突信息
- 任一冲突则拒绝整个批量操作，返回详细冲突信息

**阶段2：决策**
- 如果有任何版本冲突：
  - 返回 `success: false`
  - 返回 `error: VERSION_CONFLICT`
  - 列出所有冲突详情（pathName, providedVersion, currentVersion）
  - 列出所有检查结果（哪些正常，哪些冲突）
- 如果全部版本正常，进入写入阶段

**阶段3：事务写入**
- 开启数据库事务
- 逐个执行修改操作
- 如任一操作失败，回滚整个事务
- 全部成功则提交事务
- 返回每个操作的新版本号

**返回格式**：
- 成功：`{success: true, summary, results: [{pathName, newVersion}]}`
- 冲突：`{success: false, error, message, conflicts: [...], checks: [...]}`

### 4.4 MCP 注册对比说明 ([`server.go`](backend/internal/mcp/server.go))

#### query_module MCP 注册对比

**原注册（现在）**：
```go
// 位置: server.go:99-107
s.server.AddTool(mcp.NewTool("query_module",
    mcp.WithDescription(`查询模块信息。

可查询字段: name, description, status, prompt, upstreamContractSummary, downstreamContractSummary, testCoverage, locked, version

fields 不填返回所有字段。`),
    mcp.WithString("pathName", mcp.Description("模块路径，格式: 项目名/模块名"), mcp.Required()),
    mcp.WithArray("fields", mcp.Description("指定返回字段，如 [\"name\",\"status\"]")),
), s.handleQueryModule)
```

**新注册（改后）**：
```go
// 位置: server.go:99 左右
s.server.AddTool(mcp.NewTool("query_module",
    mcp.WithDescription(`查询模块信息。使用 queries 数组批量查询，每个路径可指定不同字段。

可查询字段: name, description, status, prompt, upstreamContractSummary, 
downstreamContractSummary, testCoverage, locked, version

输出按路径层级组织。`),
    mcp.WithArray("queries", mcp.Description("查询数组，每项: {pathName: string, fields?: string[]}"), mcp.Required()),
), s.handleQueryModule)
```

**主要变化**：
- ✅ 使用 `queries` 数组参数（每个路径可指定不同字段）
- ✅ 删除 `pathName` 和 `fields` 参数
- ✅ `queries` 标记为 Required

#### query_task MCP 注册对比

**原注册（现在）**：
```go
// 位置: server.go:109-118
s.server.AddTool(mcp.NewTool("query_task",
    mcp.WithDescription(`查询任务信息。

可查询字段: name, description, status, prompt, upstreamContractDetail, downstreamContractDetail, tests, testResult, codePaths, bugLog, humanAssistance, issueDetails, locked, version

fields 不填返回所有字段。`),
    mcp.WithString("pathName", mcp.Description("任务路径，格式: 项目名/模块名/任务名"), mcp.Required()),
    mcp.WithArray("fields", mcp.Description("指定返回字段，如 [\"name\",\"tests\"]")),
), s.handleQueryTask)
```

**新注册（改后）**：
```go
// 位置: server.go:109 左右
s.server.AddTool(mcp.NewTool("query_task",
    mcp.WithDescription(`查询任务信息。使用 queries 数组批量查询，每个路径可指定不同字段。

注意：只能传任务路径（包含两个斜杠），不支持传模块路径查询所有子任务。

可查询字段: name, description, status, prompt, upstreamContractDetail, 
downstreamContractDetail, tests, testResult, codePaths, bugLog, 
humanAssistance, issueDetails, locked, version

输出按路径层级组织。`),
    mcp.WithArray("queries", mcp.Description("查询数组，每项: {pathName: string, fields?: string[]}，pathName必须包含两个斜杠"), mcp.Required()),
), s.handleQueryTask)
```

**主要变化**：
- ✅ 使用 `queries` 数组参数（每个路径可指定不同字段）
- ✅ 删除 `pathName` 和 `fields` 参数
- ✅ `queries` 标记为 Required
- ✅ 路径验证要求在描述中说明

#### modify_module MCP 注册对比

**原注册（现在）**：
```go
// 位置: server.go:168-176
s.server.AddTool(mcp.NewTool("modify_module",
    mcp.WithDescription(`修改模块。只更新 data 中传入的字段。

可修改字段: name(string), description(string), status(designing|developing|completed|deprecated), prompt(string), upstreamContractSummary(string), downstreamContractSummary(string), testCoverage(0-100)`),
    mcp.WithString("pathName", mcp.Description("模块路径，格式: 项目名/模块名"), mcp.Required()),
    mcp.WithNumber("version", mcp.Description("当前版本号，用于乐观锁"), mcp.Required()),
    mcp.WithObject("data", mcp.Description("要修改的字段，如 {\"description\":\"新描述\",\"status\":\"developing\"}"), mcp.Required()),
), s.handleModifyModule)
```

**新注册（改后）**：
```go
// 位置: server.go:168 左右
s.server.AddTool(mcp.NewTool("modify_module",
    mcp.WithDescription(`批量修改模块。使用 operations 数组，每项包含 {pathName, version, data}。

原子性保证：预检所有版本号，任一冲突则全部拒绝。使用数据库事务保证原子性。

可修改字段: name(string), description(string), status(designing|developing|completed|deprecated), prompt(string), upstreamContractSummary(string), downstreamContractSummary(string), testCoverage(0-100)`),
    mcp.WithArray("operations", mcp.Description("批量操作数组，每项: {pathName: string, version: number, data: object}"), mcp.Required()),
), s.handleModifyModule)
```

**主要变化**：
- ✅ 使用 `operations` 数组参数
- ✅ 删除 `pathName`, `version`, `data` 参数
- ✅ `operations` 标记为 Required
- ✅ 描述中强调原子性保证和预检机制

#### modify_task MCP 注册对比

**原注册（现在）**：
```go
// 位置: server.go:178-192
s.server.AddTool(mcp.NewTool("modify_task",
    mcp.WithDescription(`修改任务。只更新 data 中传入的字段。

可修改字段:
- name, description, status, prompt: string
- upstreamContractDetail/downstreamContractDetail: {title, list:[{label,contract_api,from}]}
- tests: [{target:string, api:string}]
- testResult/codePaths/bugLog: string[]
- humanAssistance: object
- issueDetails: string`),
    mcp.WithString("pathName", mcp.Description("任务路径，格式: 项目名/模块名/任务名"), mcp.Required()),
    mcp.WithNumber("version", mcp.Description("当前版本号，用于乐观锁"), mcp.Required()),
    mcp.WithObject("data", mcp.Description("要修改的字段，如 {\"status\":\"in_progress\",\"tests\":[{\"target\":\"验证XX\",\"api\":\"test_xx()\"}]}"), mcp.Required()),
), s.handleModifyTask)
```

**新注册（改后）**：
```go
// 位置: server.go:178 左右
s.server.AddTool(mcp.NewTool("modify_task",
    mcp.WithDescription(`批量修改任务。使用 operations 数组，每项包含 {pathName, version, data}。

原子性保证：预检所有版本号，任一冲突则全部拒绝。使用数据库事务保证原子性。

可修改字段:
- name, description, status, prompt: string
- upstreamContractDetail/downstreamContractDetail: {title, list:[{label,contract_api,from}]}
- tests: [{target:string, api:string}]
- testResult/codePaths/bugLog: string[]
- humanAssistance: object
- issueDetails: string`),
    mcp.WithArray("operations", mcp.Description("批量操作数组，每项: {pathName: string, version: number, data: object}"), mcp.Required()),
), s.handleModifyTask)
```

**主要变化**：
- ✅ 使用 `operations` 数组参数
- ✅ 删除 `pathName`, `version`, `data` 参数
- ✅ `operations` 标记为 Required
- ✅ 描述中强调原子性保证和预检机制

---

## 5. 测试要点

### 5.1 查询函数测试

**query_module 测试场景**:
- 批量查询多个模块（每个路径指定不同字段）
- 所有路径使用相同字段
- 空数组拒绝
- 部分路径不存在（部分成功）

**query_task 测试场景**:
- 批量查询多个任务（每个路径不同字段）
- 所有路径使用相同字段
- 拒绝模块路径（必须包含两个斜杠）

### 5.2 修改函数测试

**modify_module 测试场景**:
- 批量修改多个模块（operations）
- 原子性测试：任一版本冲突时全部拒绝
- 版本冲突的详细错误信息
- 事务回滚测试
- 空operations数组拒绝

**modify_task 测试场景**:
- 批量修改多个任务（operations）
- 原子性保证测试
- JSON字段处理（tests, codePaths等）
- 版本冲突详细信息
- 空operations数组拒绝

### 5.3 关键测试用例

#### 原子性测试示例
```
场景：批量修改3个模块，其中1个版本落后
预期：全部拒绝，返回详细冲突信息
验证：
- 数据库中3个模块都未被修改
- 返回包含conflicts列表
- 返回包含所有checks结果
```

#### 灵活查询测试示例
```
场景：使用queries参数查询2个模块，指定不同字段
输入：queries=[
  {pathName: "A", fields: ["name","status"]},
  {pathName: "B", fields: ["description","prompt"]}
]
预期：
- 模块A返回name和status
- 模块B返回description和prompt
- 输出按路径层级组织
```

---

## 6. 实施步骤

### 阶段1：MCP注册修改
1. 修改 `server.go`，为4个函数添加新参数
2. 更新 MCP 描述，说明新特性

### 阶段2：查询函数实现
1. 修改 `tools_get.go` 中的 `handleQueryModuleImpl`
   - 支持 `queries` 灵活模式和 `pathNames` 简化模式
   - 实现按路径树形输出
2. 修改 `handleQueryTaskImpl`（同上）

### 阶段3：修改函数实现  
1. 修改 `tools_modify.go` 中的 `handleModifyModuleImpl`
   - 支持 `operations` 批量模式
   - 实现三阶段原子性保证（预检版本 → 决策 → 事务写入）
   - 详细的版本冲突错误信息
2. 修改 `handleModifyTaskImpl`（同上）

### 阶段4：测试和文档
1. 编写单元测试（`tools_get_test.go`, `tools_modify_test.go`）
2. 更新 `docs/mcp-guide.md` 添加批量操作说明
3. 进行集成测试

---

## 7. 注意事项

### 7.1 修改函数的原子性保证关键点

**必须严格遵守三阶段流程**：
1. **预检阶段**：先查询所有路径的当前版本，任一冲突则立即拒绝
2. **决策阶段**：只有全部版本检查通过才进入写入
3. **写入阶段**：使用数据库事务，任一失败则回滚全部

**错误返回必须包含**：
- 哪个路径冲突
- 提供的版本号 vs 当前版本号
- 明确提示"请获取新版本后重试"

### 7.2 查询函数的灵活性关键点

**queries 模式优势**：
- 每个路径可指定不同字段
- 适配大模型的智能查询能力
- 避免不必要的数据传输
- 统一使用数组参数，接口简洁

---

## 8. 文件修改清单

| 文件 | 修改内容 |
|------|---------|
| `backend/internal/mcp/server.go` | 修改4个函数的MCP注册，新增参数 |
| `backend/internal/mcp/tools_get.go` | 重构query_module和query_task实现 |
| `backend/internal/mcp/tools_modify.go` | 重构modify_module和modify_task实现 |
| `backend/internal/mcp/tools_get_test.go` | 新增批量查询测试用例 |
| `backend/internal/mcp/tools_modify_test.go` | 新增批量修改和原子性测试用例 |
| `docs/mcp-guide.md` | 添加批量操作使用说明 |
