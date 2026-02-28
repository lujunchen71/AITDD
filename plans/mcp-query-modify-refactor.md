# MCP query/modify 函数重构方案

## 问题分析

当前 `query_node` 和 `modify_node` 是通用接口，描述过于简单，导致 AI 无法知道：
- 不同节点类型（module/task）有哪些字段
- 每个字段的数据格式是什么
- 哪些字段可以修改，哪些是只读的

## 重构方案：拆分为类型特定函数，删除旧函数

### 1. query_module - 查询模块

```
query_module(pathName, fields?)
```

| 参数 | 类型 | 必填 | 说明 |
|------|------|------|------|
| pathName | string | 是 | 模块路径，格式: `项目名/模块名` |
| fields | string[] | 否 | 指定返回字段，不填返回所有 |

**可查询字段**:
- `name` (string) - 模块名称
- `description` (string) - 模块描述
- `status` (string) - 状态: designing/developing/completed/deprecated
- `prompt` (string) - 模块级别的开发提示
- `upstreamContractSummary` (string) - 上游契约摘要
- `downstreamContractSummary` (string) - 下游契约摘要
- `testCoverage` (number) - 测试覆盖率 0-100
- `locked` (boolean) - 是否被锁定
- `version` (number) - 版本号

---

### 2. query_task - 查询任务

```
query_task(pathName, fields?)
```

| 参数 | 类型 | 必填 | 说明 |
|------|------|------|------|
| pathName | string | 是 | 任务路径，格式: `项目名/模块名/任务名` |
| fields | string[] | 否 | 指定返回字段，不填返回所有 |

**可查询字段**:
- `name` (string) - 任务名称
- `description` (string) - 任务描述
- `status` (string) - 状态: ready/claimed/in_progress/pending_review/completed/failed/blocked
- `prompt` (string) - 任务开发提示
- `upstreamContractDetail` (object) - 上游契约详情
- `downstreamContractDetail` (object) - 下游契约详情
- `tests` (array) - 测试用例列表
- `testResult` (array) - 测试结果列表
- `codePaths` (array) - 代码文件路径列表
- `bugLog` (array) - Bug 日志
- `humanAssistance` (object) - 人工协助信息
- `issueDetails` (string) - 问题详情
- `locked` (boolean) - 是否被锁定
- `version` (number) - 版本号

---

### 3. modify_module - 修改模块

```
modify_module(pathName, version, data)
```

| 参数 | 类型 | 必填 | 说明 |
|------|------|------|------|
| pathName | string | 是 | 模块路径，格式: `项目名/模块名` |
| version | number | 是 | 当前版本号，用于乐观锁 |
| data | object | 是 | 要修改的字段键值对 |

**可修改字段**:
| 字段 | 类型 | 说明 |
|------|------|------|
| name | string | 模块名称 |
| description | string | 模块描述 |
| status | string | designing/developing/completed/deprecated |
| prompt | string | 开发提示 |
| upstreamContractSummary | string | 上游契约摘要 |
| downstreamContractSummary | string | 下游契约摘要 |
| testCoverage | number | 测试覆盖率 0-100 |

**示例**:
```json
{
  "pathName": "Calculator/UI模块",
  "version": 1,
  "data": {
    "description": "新的描述",
    "status": "developing"
  }
}
```

---

### 4. modify_task - 修改任务

```
modify_task(pathName, version, data)
```

| 参数 | 类型 | 必填 | 说明 |
|------|------|------|------|
| pathName | string | 是 | 任务路径，格式: `项目名/模块名/任务名` |
| version | number | 是 | 当前版本号，用于乐观锁 |
| data | object | 是 | 要修改的字段键值对 |

**可修改字段**:
| 字段 | 类型 | 说明 |
|------|------|------|
| name | string | 任务名称 |
| description | string | 任务描述 |
| status | string | ready/claimed/in_progress/pending_review/completed/failed/blocked |
| prompt | string | 开发提示 |
| upstreamContractDetail | object | 上游契约详情 |
| downstreamContractDetail | object | 下游契约详情 |
| tests | array | 测试用例列表 |
| testResult | array | 测试结果列表 |
| codePaths | array | 代码文件路径列表 |
| bugLog | array | Bug 日志 |
| humanAssistance | object | 人工协助信息 |
| issueDetails | string | 问题详情 |

**复杂字段格式**:

```json
// upstreamContractDetail / downstreamstreamContractDetail
{
  "title": "契约标题",
  "list": [
    {
      "label": "功能说明",
      "contract_api": "API签名",
      "from": "来源任务"
    }
  ]
}

// tests
[
  {
    "target": "测试目标",
    "api": "test_function_name()"
  }
]

// testResult / codePaths / bugLog
["字符串条目1", "字符串条目2"]

// humanAssistance
{}
```

**示例**:
```json
{
  "pathName": "Calculator/UI模块/主窗口",
  "version": 1,
  "data": {
    "status": "in_progress",
    "tests": [
      {"target": "验证窗口标题", "api": "test_window_title()"},
      {"target": "验证窗口大小", "api": "test_window_size()"}
    ]
  }
}
```

---

## Go 代码实现

### server.go 工具注册

```go
// ==================== 查询工具 ====================

// query_module - 查询模块
s.server.AddTool(mcp.NewTool("query_module",
    mcp.WithDescription(`查询模块信息。

可查询字段: name, description, status, prompt, upstreamContractSummary, downstreamContractSummary, testCoverage, locked, version

fields 不填返回所有字段。`),
    mcp.WithString("pathName", mcp.Description("模块路径，格式: 项目名/模块名"), mcp.Required()),
    mcp.WithArray("fields", mcp.Description("指定返回字段，如 [\"name\",\"status\"]")),
), s.handleQueryModule)

// query_task - 查询任务
s.server.AddTool(mcp.NewTool("query_task",
    mcp.WithDescription(`查询任务信息。

可查询字段: name, description, status, prompt, upstreamContractDetail, downstreamContractDetail, tests, testResult, codePaths, bugLog, humanAssistance, issueDetails, locked, version

fields 不填返回所有字段。`),
    mcp.WithString("pathName", mcp.Description("任务路径，格式: 项目名/模块名/任务名"), mcp.Required()),
    mcp.WithArray("fields", mcp.Description("指定返回字段，如 [\"name\",\"tests\"]")),
), s.handleQueryTask)

// ==================== 修改工具 ====================

// modify_module - 修改模块
s.server.AddTool(mcp.NewTool("modify_module",
    mcp.WithDescription(`修改模块。只更新 data 中传入的字段。

可修改字段: name(string), description(string), status(designing|developing|completed|deprecated), prompt(string), upstreamContractSummary(string), downstreamContractSummary(string), testCoverage(0-100)`),
    mcp.WithString("pathName", mcp.Description("模块路径，格式: 项目名/模块名"), mcp.Required()),
    mcp.WithNumber("version", mcp.Description("当前版本号，用于乐观锁"), mcp.Required()),
    mcp.WithObject("data", mcp.Description("要修改的字段，如 {\"description\":\"新描述\",\"status\":\"developing\"}"), mcp.Required()),
), s.handleModifyModule)

// modify_task - 修改任务
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

---

## 变更清单

### 删除
- `query_node` - 通用查询接口
- `modify_node` - 通用修改接口

### 新增
- `query_module` - 查询模块
- `query_task` - 查询任务
- `modify_module` - 修改模块
- `modify_task` - 修改任务

---

## 优势

1. **类型明确** - 函数名直接表明操作对象类型
2. **字段可见** - 描述中列出所有可用字段
3. **格式清晰** - 复杂字段有明确的结构说明
4. **减少错误** - AI 不会传错字段名或格式
