# AITDD MCP 编译器架构设计文档

> **核心原则**：所有接口通过 pathName 索引，不暴露内部 ID。返回内容简洁直观，面向 AI 消费。

---

## 1. 项目上下文

### `init_project` - 初始化/设置当前项目

**说明**：不传 pathName 则列出所有项目供选择；传入 pathName 则设置为当前项目。

**参数**：

| 参数 | 类型 | 必填 | 说明 |
|------|------|------|------|
| pathName | string | 否 | 项目路径名称 |

**输出（列表模式，未传 pathName）**：
```
projects:
  - name: "AITDD"
    pathName: "aitdd"
    description: "AI驱动的设计与开发工具"
  - name: "PVZ"
    pathName: "pvz"
    description: "植物大战僵尸游戏"
```

**输出（设置模式，传入 pathName）**：
```
success: true
projectName: "AITDD"
pathName: "aitdd"
```

---

### `get_context` - 获取当前上下文

**参数**：无

**输出**：
```
projectName: "AITDD"
pathName: "aitdd"
apiBaseUrl: "http://localhost:3000"
```

---

## 2. 信息查询

### `query_node` - 查询节点信息

**说明**：通用查询接口，根据 path 自动识别类型（project/module/task），返回对应信息。fields 不填则返回该节点所有字段，填了只返回指定字段。

**参数**：

| 参数 | 类型 | 必填 | 说明 |
|------|------|------|------|
| path | string | 是 | 节点 pathName |
| fields | string[] | 否 | 指定返回字段，不填返回所有 |

**输出（查询 project，path="aitdd"）**：
```
pathName: "aitdd"
name: "AITDD"
constitution: "项目公约内容..."
totalModules: 5
totalTasks: 23
completedTasks: 12
```

**输出（查询 module，path="aitdd/auth"）**：
```
pathName: "aitdd/auth"
name: "认证模块"
description: "负责用户认证相关功能"
prompt: "模块提示词..."
status: "developing"
upstreamContractSummary: "..."
downstreamContractSummary: "..."
version: 3
```

**输出（查询 task，path="aitdd/auth/login"）**：
```
pathName: "aitdd/auth/login"
modulePathName: "aitdd/auth"
name: "用户登录"
description: "实现用户登录功能"
status: "completed"
prompt: "任务提示词..."
upstreamContractDetail: "上游契约JSON..."
downstreamContractDetail: "下游契约JSON..."
tests: "[测试用例JSON数组]"
testResult: "[测试结果JSON数组]"
codePaths: "[代码路径JSON数组]"
bugLog: "[Bug日志JSON数组]"
humanAssistance: "{人工协助JSON}"
locked: false
lockedBy: ""
version: 3
```

**输出（指定 fields=["name","status"]，path="aitdd/auth/login"）**：
```
name: "用户登录"
status: "completed"
```

---

### `query_project_index_tree` - 查询项目计划索引树

**说明**：以树状文件列表形式展示项目→模块→任务的层级结构。

**参数**：

| 参数 | 类型 | 必填 | 说明 |
|------|------|------|------|
| pathName | string | 否 | 起始 pathName，不传则返回整个项目树 |

**输出**：
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

每行格式：`pathName片段 # 名称, 类型, 状态`

---

### `query_file_code_path_tree` - 查询代码文件路径树

**说明**：以树状文件列表形式展示任务或模块关联的实际代码文件路径。

**参数**：

| 参数 | 类型 | 必填 | 说明 |
|------|------|------|------|
| pathName | string | 是 | 任务或模块的 pathName |

**输出（查询模块 path="aitdd/auth"）**：
```
aitdd/auth                         # 认证模块, module
├── src/auth/login.ts              # 用户登录 (aitdd/auth/login)
├── src/auth/login.test.ts         # 用户登录 (aitdd/auth/login)
├── src/auth/register.ts           # 用户注册 (aitdd/auth/register)
├── src/auth/register.test.ts      # 用户注册 (aitdd/auth/register)
└── src/auth/logout.ts             # 用户登出 (aitdd/auth/logout)
totalFiles: 5
```

**输出（查询任务 path="aitdd/auth/login"）**：
```
aitdd/auth/login                   # 用户登录, task
├── src/auth/login.ts
└── src/auth/login.test.ts
totalFiles: 2
```

每行格式：`文件路径 # 所属任务名称 (任务pathName)`

---

## 3. 验证检查

### `check_contract_alignment` - 契约对齐检查

**参数**：

| 参数 | 类型 | 必填 | 说明 |
|------|------|------|------|
| upstreamPathName | string | 是 | 上游任务 pathName |
| downstreamPathName | string | 是 | 下游任务 pathName |

**输出**：
```
aligned: false
differences:
  - field: "responseFormat"
    upstream: "{type: 'json', schema: {...}}"
    downstream: "{type: 'xml'}"
```

---

### `check_task_readiness` - 任务准备度检查

**参数**：

| 参数 | 类型 | 必填 | 说明 |
|------|------|------|------|
| pathName | string | 是 | 任务 pathName |

**输出**：
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

### `check_dependencies` - 依赖关系检查

**参数**：

| 参数 | 类型 | 必填 | 说明 |
|------|------|------|------|
| pathName | string | 是 | 任务/模块 pathName |
| type | string | 是 | "task" 或 "module" |

**输出**：
```
hasCycle: false
blocked: true
blockedBy:
  - pathName: "aitdd/auth/login"
    name: "用户登录"
    status: "in_progress"
upstream:
  - pathName: "aitdd/auth/login"
    name: "用户登录"
downstream:
  - pathName: "aitdd/auth/logout"
    name: "用户登出"
```

---

## 4. 节点操作（CRUD）

### `create_node` - 创建节点

**说明**：通用创建接口，根据 type 创建模块或任务。

**参数**：

| 参数 | 类型 | 必填 | 说明 |
|------|------|------|------|
| type | string | 是 | "module" 或 "task" |
| parentPath | string | 是 | 父节点 pathName（module 的父是 project 或 module，task 的父是 module） |
| name | string | 是 | 名称 |
| pathName | string | 否 | 不传则自动生成 |
| data | object | 否 | 类型相关的具体字段 |

**data 字段（type="module"）**：

| 字段 | 说明 |
|------|------|
| description | 模块描述 |
| prompt | 模块提示词 |
| upstreamContractSummary | 上游契约摘要 |
| downstreamContractSummary | 下游契约摘要 |

**data 字段（type="task"）**：

| 字段 | 说明 |
|------|------|
| description | 任务描述 |
| prompt | 任务提示词 |
| upstreamContractDetail | 上游契约详情 JSON |
| downstreamContractDetail | 下游契约详情 JSON |
| tests | 测试用例 JSON 数组 |
| codePaths | 代码路径 JSON 数组 |

**输出**：
```
success: true
pathName: "aitdd/auth/login"
```

---

### `modify_node` - 修改节点

**说明**：通用修改接口，根据 path 自动识别类型，只更新传入的字段。也用于修改依赖关系等。

**参数**：

| 参数 | 类型 | 必填 | 说明 |
|------|------|------|------|
| path | string | 是 | 节点 pathName |
| version | number | 是 | 当前版本号（乐观锁） |
| data | object | 是 | 要修改的字段键值对 |

**data 可用字段（project）**：

| 字段 | 说明 |
|------|------|
| name | 项目名称 |
| constitution | 项目公约 |

**data 可用字段（module）**：

| 字段 | 说明 |
|------|------|
| name | 模块名称 |
| description | 模块描述 |
| prompt | 模块提示词 |
| status | 模块状态 |
| upstreamContractSummary | 上游契约摘要 |
| downstreamContractSummary | 下游契约摘要 |

**data 可用字段（task）**：

| 字段 | 说明 |
|------|------|
| name | 任务名称 |
| description | 任务描述 |
| status | 任务状态 |
| prompt | 任务提示词 |
| upstreamContractDetail | 上游契约详情 |
| downstreamContractDetail | 下游契约详情 |
| tests | 测试用例 |
| codePaths | 代码路径 |
| testResult | 测试结果 |

**输出**：
```
success: true
version: 4
```

---

### `delete_node` - 删除节点

**说明**：通用删除接口，根据 path 自动识别类型。删除模块时级联删除其下所有任务。

**参数**：

| 参数 | 类型 | 必填 | 说明 |
|------|------|------|------|
| path | string | 是 | 节点 pathName |
| force | boolean | 否 | 强制删除（忽略依赖） |

**输出**：
```
success: true
deletedChildren: 3
```

---

## 5. 依赖管理

### `create_dependency` - 创建依赖

**参数**：

| 参数 | 类型 | 必填 | 说明 |
|------|------|------|------|
| type | string | 是 | "module" 或 "task" |
| upstreamPath | string | 是 | 上游 pathName |
| downstreamPath | string | 是 | 下游 pathName |
| dependencyType | string | 否 | required / optional / conditional |
| contractSummary | string | 否 | 契约摘要 |

**输出**：
```
success: true
```

---

### `delete_dependency` - 删除依赖

**参数**：

| 参数 | 类型 | 必填 | 说明 |
|------|------|------|------|
| type | string | 是 | "module" 或 "task" |
| upstreamPath | string | 是 | 上游 pathName |
| downstreamPath | string | 是 | 下游 pathName |

**输出**：
```
success: true
```

---

### `query_dependencies` - 查询依赖

**参数**：

| 参数 | 类型 | 必填 | 说明 |
|------|------|------|------|
| type | string | 是 | "module" 或 "task" |
| pathName | string | 是 | 资源 pathName |
| direction | string | 否 | upstream / downstream / both |

**输出**：
```
upstream:
  - pathName: "aitdd/auth/login"
    name: "用户登录"
    dependencyType: "required"
downstream:
  - pathName: "aitdd/auth/logout"
    name: "用户登出"
    dependencyType: "optional"
```

---

## 6. 问答系统（Issue）

### `create_issue` - 创建问题

**说明**：下游任务发现上游有问题时发起。

**参数**：

| 参数 | 类型 | 必填 | 说明 |
|------|------|------|------|
| fromTaskPathName | string | 是 | 发起方任务 pathName |
| toTaskPathName | string | 是 | 接收方任务 pathName |
| type | string | 是 | contract / test / other |
| title | string | 是 | 问题标题 |
| content | string | 是 | 问题内容 |

**输出**：
```
success: true
status: "pending"
```

---

### `reply_issue` - 回复问题

**参数**：

| 参数 | 类型 | 必填 | 说明 |
|------|------|------|------|
| fromTaskPathName | string | 是 | 发起方任务 pathName |
| toTaskPathName | string | 是 | 接收方任务 pathName |
| title | string | 是 | 问题标题（定位用） |
| replyContent | string | 是 | 回复内容 |

**输出**：
```
success: true
status: "replied"
```

---

### `resolve_issue` - 解决问题

**参数**：

| 参数 | 类型 | 必填 | 说明 |
|------|------|------|------|
| fromTaskPathName | string | 是 | 发起方任务 pathName |
| toTaskPathName | string | 是 | 接收方任务 pathName |
| title | string | 是 | 问题标题（定位用） |

**输出**：
```
success: true
status: "resolved"
```

---

### `query_issues` - 查询问题列表

**参数**：

| 参数 | 类型 | 必填 | 说明 |
|------|------|------|------|
| taskPathName | string | 否 | 任务 pathName 过滤 |
| direction | string | 否 | from / to / both（默认 both） |
| status | string | 否 | pending / replied / resolved |
| type | string | 否 | contract / test / other |

**输出**：
```
issues:
  - fromTaskPathName: "aitdd/auth/register"
    fromTaskName: "用户注册"
    toTaskPathName: "aitdd/auth/login"
    toTaskName: "用户登录"
    type: "contract"
    title: "登录接口返回格式不一致"
    status: "pending"
    createdAt: 1709100000000
total: 1
```

---

## 7. 静态编译

### `compile_static` - 静态编译

**说明**：硬编码规则检查，不需要 AI 推理。

**参数**：

| 参数 | 类型 | 必填 | 说明 |
|------|------|------|------|
| scope | string | 是 | project / module / task |
| pathName | string | 否 | scope 为 module/task 时必填 |
| rules | string[] | 否 | 指定规则 ID，不传则全部检查 |

**输出**：
```
success: false
errorCount: 2
warningCount: 1
errors:
  - ruleId: "E-S-04"
    ruleName: "提示词为空"
    resourceType: "task"
    resourceName: "用户登录"
    resourcePathName: "aitdd/auth/login"
    message: "任务 [用户登录] 提示词为空"
    suggestion: "请为任务添加提示词"
  - ruleId: "E-S-07"
    ruleName: "契约不一致"
    resourceType: "task"
    resourceName: "用户注册"
    resourcePathName: "aitdd/auth/register"
    message: "任务 [用户登录] 与 [用户注册] 契约不一致"
    suggestion: "请检查上下游契约定义"
warnings:
  - ruleId: "W-S-01"
    ruleName: "测试用例为空"
    resourceType: "task"
    resourceName: "用户注册"
    resourcePathName: "aitdd/auth/register"
    message: "任务 [用户注册] 未定义测试用例"
    suggestion: "建议添加测试用例"
```

**静态规则（Error - 阻断性）**：

| 规则ID | 名称 | 检查条件 |
|--------|------|----------|
| E-S-01 | 代码路径为空 | codePaths 为空 |
| E-S-02 | 存在 Bug 日志 | bugLog 不为空 |
| E-S-03 | 需要人工协助 | humanAssistance 不为空 |
| E-S-04 | 提示词为空 | prompt 为空 |
| E-S-05 | 上游契约为空 | upstreamContractDetail 为空（非首个任务） |
| E-S-06 | 下游契约为空 | downstreamContractDetail 为空（非末尾任务） |
| E-S-07 | 契约不一致 | 上游的下游契约 ≠ 下游的上游契约（字符对比） |
| E-S-08 | 状态异常 | 状态为 completed 但存在错误信息 |
| E-S-09 | 锁定修改冲突 | 任务被锁时尝试修改 |

**静态规则（Warning - 非阻断性）**：

| 规则ID | 名称 | 检查条件 |
|--------|------|----------|
| W-S-01 | 测试用例为空 | tests 为空 |
| W-S-02 | 模块描述为空 | 模块 description 为空 |
| W-S-03 | 存在未回复问题 | 有 status=pending 的 Issue |

---

## 8. 动态编译

### `compile_dynamic` - 动态编译

**说明**：需要 AI 推理的检查，通用接口。

**参数**：

| 参数 | 类型 | 必填 | 说明 |
|------|------|------|------|
| scope | string | 是 | project / module / task |
| pathName | string | 否 | 资源 pathName |
| checks | string[] | 否 | 指定检查项 |

**输出**：
```
success: true
totalChecks: 2
passedChecks: 1
results:
  - checkId: "D-01"
    checkName: "模块设计合理性"
    passed: true
    score: 90
    analysis: "模块划分合理，职责单一..."
    suggestions:
      - priority: "medium"
        content: "建议将用户认证相关功能独立为一个模块"
  - checkId: "D-05"
    checkName: "任务准备度评估"
    passed: false
    score: 60
    analysis: "任务准备度不足..."
    suggestions:
      - priority: "high"
        content: "需要补充上游契约的输入参数说明"
```

**动态检查项**：

| 检查ID | 名称 | 说明 |
|--------|------|------|
| D-01 | 模块设计合理性 | 模块职责是否单一、边界是否清晰 |
| D-02 | 契约摘要对齐 | 上下游契约摘要一致性 |
| D-03 | 契约充分性 | 契约是否覆盖所有必要接口 |
| D-04 | 架构调整建议 | 整体架构优化建议 |
| D-05 | 任务准备度评估 | 任务完整性、清晰度、可执行性 |

---

## 9. 锁定管理

### `lock_resource` - 锁定资源

**参数**：

| 参数 | 类型 | 必填 | 说明 |
|------|------|------|------|
| resourceType | string | 是 | "module" 或 "task" |
| pathName | string | 是 | 资源 pathName |

**输出**：
```
success: true
```

---

### `unlock_resource` - 解锁资源

**参数**：

| 参数 | 类型 | 必填 | 说明 |
|------|------|------|------|
| resourceType | string | 是 | "module" 或 "task" |
| pathName | string | 是 | 资源 pathName |

**输出**：
```
success: true
```

---

### `query_lock_status` - 查询锁定状态

**参数**：

| 参数 | 类型 | 必填 | 说明 |
|------|------|------|------|
| resourceType | string | 是 | "module" 或 "task" |
| pathName | string | 是 | 资源 pathName |

**输出**：
```
locked: true
lockedBy: "agent-task-owner-01"
lockedAt: 1709100000000
```

---

## 10. 通知管理

### `send_notification` - 发送通知

**参数**：

| 参数 | 类型 | 必填 | 说明 |
|------|------|------|------|
| pathName | string | 是 | 目标任务 pathName |
| type | string | 是 | 通知类型 |
| title | string | 是 | 通知标题 |
| message | string | 是 | 通知内容 |

**输出**：
```
success: true
```

---

### `query_notifications` - 查询通知

**参数**：

| 参数 | 类型 | 必填 | 说明 |
|------|------|------|------|
| pathName | string | 否 | 任务 pathName 过滤 |
| unreadOnly | boolean | 否 | 仅未读 |

**输出**：
```
notifications:
  - type: "task_completed"
    title: "上游任务已完成"
    message: "用户登录任务已完成，可以开始开发"
    read: false
    createdAt: 1709100000000
total: 1
unreadCount: 1
```

---

### `mark_notification_read` - 标记通知已读

**参数**：

| 参数 | 类型 | 必填 | 说明 |
|------|------|------|------|
| pathName | string | 是 | 任务 pathName |

**输出**：
```
success: true
```

---

## 11. 辅助工具

### `open_frontend` - 打开前端网页

**参数**：无

**输出**：
```
success: true
message: "已在浏览器中打开 http://localhost:5173"
```

---

## 12. 完整接口清单

| 分类 | 接口名称 | 说明 |
|------|----------|------|
| 上下文 | `init_project` | 初始化/设置当前项目 |
| 上下文 | `get_context` | 获取当前上下文 |
| 查询 | `query_node` | 查询节点信息（project/module/task 通用） |
| 查询 | `query_project_index_tree` | 查询项目计划索引树 |
| 查询 | `query_file_code_path_tree` | 查询代码文件路径树 |
| 验证 | `check_contract_alignment` | 契约对齐检查 |
| 验证 | `check_task_readiness` | 任务准备度检查 |
| 验证 | `check_dependencies` | 依赖关系检查 |
| CRUD | `create_node` | 创建节点（module/task 通用） |
| CRUD | `modify_node` | 修改节点（project/module/task 通用） |
| CRUD | `delete_node` | 删除节点（module/task 通用，级联删除） |
| 依赖 | `create_dependency` | 创建依赖 |
| 依赖 | `delete_dependency` | 删除依赖 |
| 依赖 | `query_dependencies` | 查询依赖 |
| 锁定 | `lock_resource` | 锁定资源 |
| 锁定 | `unlock_resource` | 解锁资源 |
| 锁定 | `query_lock_status` | 查询锁定状态 |
| 通知 | `send_notification` | 发送通知 |
| 通知 | `query_notifications` | 查询通知 |
| 通知 | `mark_notification_read` | 标记通知已读 |
| 辅助 | `open_frontend` | 打开前端网页 |
| 问题 | `create_issue` | 创建问题 |
| 问题 | `reply_issue` | 回复问题 |
| 问题 | `resolve_issue` | 解决问题 |
| 问题 | `query_issues` | 查询问题列表 |
| 编译 | `compile_static` | 静态编译 |
| 编译 | `compile_dynamic` | 动态编译 |

**接口总数：27 个**

---

## 13. 数据库修改要点

| 变更 | 说明 |
|------|------|
| 新增 `issues` 表 | 独立的问题追踪，替代 tasks.issue_details 字段 |
| 删除 `tasks.issue_details` | 迁移到 issues 表 |
| 新增 `compile_history` 表（可选） | 记录编译历史 |
| 预留 `members` 表 | 后续成员系统 |
| 预留 `audit_logs` 表 | 后续审计日志 |

---

## 14. 与现有系统对照

| 原接口（38个） | 新接口（27个） | 变更 |
|--------|--------|------|
| `init_project` + `set_project` | `init_project` | 合并 |
| `get_config` | `get_context` | 重命名 |
| `get_project_info` + `get_constitution` | `query_node` | 合并为通用查询 |
| `get_all_modules` + `get_module_overview` | `query_node` | 合并为通用查询 |
| `get_module_tasks` + `get_module_task_path_names` + `get_all_task_status` + `get_module_task_status` | `query_node` | 合并为通用查询 |
| `get_task_detail` + `get_task_contracts` + `get_all_task_code_paths` | `query_node` | 合并为通用查询 |
| - | `query_project_index_tree` | **新增** |
| `get_all_task_code_paths` | `query_file_code_path_tree` | 重构为树形 |
| `create_module` + `create_task` | `create_node` | 合并为通用创建 |
| `update_module` + `update_module_full` + `update_task` + `update_task_full` | `modify_node` | 合并为通用修改 |
| `delete_module` + `delete_module_tasks` + `delete_task` | `delete_node` | 合并为通用删除 |
| `create_module_dependency` + `create_task_dependency` | `create_dependency` | 合并 |
| `get_module_dependencies` + `get_task_dependencies` | `query_dependencies` | 合并 |
| `get_project_errors` + `get_module_errors` + `check_module` | `compile_static` | 合并 |
| `read_notifications` | `query_notifications` | 重命名 |
| `get_lock_status` | `query_lock_status` | 重命名 |
| `lock_resource` | `lock_resource` | 保持 |
| `unlock_resource` | `unlock_resource` | 保持 |
| `send_notification` | `send_notification` | 保持 |
| `open_frontend` | `open_frontend` | 保持 |
| - | `compile_dynamic` | **新增** |
| - | `create_issue` | **新增** |
| - | `reply_issue` | **新增** |
| - | `resolve_issue` | **新增** |
| - | `query_issues` | **新增** |
| - | `delete_dependency` | **新增** |
| - | `mark_notification_read` | **新增** |
