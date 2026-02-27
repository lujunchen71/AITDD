# AITDD 数据库模型定义

本文档列出所有数据库模型的完整字段定义，供 MCP 重构参考。

---

## 1. Project（项目）

**表名**: `projects`

| 字段名 | JSON标签 | Go类型 | GORM约束 | 描述 |
|-------|---------|--------|---------|------|
| ID | `id` | string | primaryKey;type:text | 项目ID |
| Name | `name` | string | not null;type:text | 项目名称 |
| PathName | `pathName` | string | unique;type:text | 路径标识 |
| Constitution | `constitution` | string | type:text | 项目公约/架构信息 |
| CreatedAt | `createdAt` | int64 | not null | 创建时间戳(毫秒) |
| UpdatedAt | `updatedAt` | int64 | not null | 更新时间戳(毫秒) |
| Version | `version` | int | not null;default:1 | 版本号(乐观锁) |
| SyncStatus | `syncStatus` | string | not null;default:'SYNCED';type:text | 同步状态 |

**SyncStatus 枚举值**:
- `SYNCED` - 已同步
- `PENDING_UPLOAD` - 待上传
- `PENDING_DOWNLOAD` - 待下载
- `CONFLICT` - 冲突

---

## 2. Module（模块）

**表名**: `modules`

| 字段名 | JSON标签 | Go类型 | GORM约束 | 描述 |
|-------|---------|--------|---------|------|
| ID | `id` | string | primaryKey;type:text | 模块ID |
| ParentID | `parentId` | *string | type:text;index | 父模块ID |
| ProjectID | `projectId` | string | not null;type:text;index | 所属项目ID |
| Name | `name` | string | not null;type:text | 模块名称 |
| PathName | `pathName` | string | unique;type:text | 路径标识 |
| Description | `description` | string | type:text | 模块描述 |
| Prompt | `prompt` | string | type:text | 模块提示词 |
| Status | `status` | string | not null;default:'designing';type:text;index | 模块状态 |
| TestCoverage | `testCoverage` | float64 | default:0 | 测试覆盖率 |
| UpstreamContractSummary | `upstreamContractSummary` | string | type:text | 上游契约摘要 |
| DownstreamContractSummary | `downstreamContractSummary` | string | type:text | 下游契约摘要 |
| Locked | `locked` | bool | not null;default:false | 是否锁定 |
| LockedBy | `lockedBy` | *string | type:text | 锁定者 |
| LockedAt | `lockedAt` | *int64 | type:integer | 锁定时间 |
| LockExpiresAt | `lockExpiresAt` | *int64 | type:integer | 锁过期时间 |
| PositionX | `positionX` | *float64 | type:real | X坐标(图形视图) |
| PositionY | `positionY` | *float64 | type:real | Y坐标(图形视图) |
| PositionUpdatedAt | `positionUpdatedAt` | *int64 | type:integer | 位置更新时间 |
| CreatedAt | `createdAt` | int64 | not null | 创建时间戳 |
| UpdatedAt | `updatedAt` | int64 | not null | 更新时间戳 |
| Version | `version` | int | not null;default:1 | 版本号 |
| SyncStatus | `syncStatus` | string | not null;default:'SYNCED';type:text | 同步状态 |

**Status 枚举值**:
- `designing` - 设计中
- `developing` - 开发中
- `completed` - 已完成
- `deprecated` - 已废弃

---

## 3. Task（任务）

**表名**: `tasks`

| 字段名 | JSON标签 | GORM约束 | Go类型 | 描述 |
|-------|---------|---------|--------|------|
| ID | `id` | primaryKey;type:text | string | 任务ID |
| ModuleID | `moduleId` | not null;type:text;index | string | 所属模块ID |
| Name | `name` | not null;type:text | string | 任务名称 |
| PathName | `pathName` | unique;type:text | string | 路径标识 |
| Description | `description` | type:text | string | 任务描述 |
| Status | `status` | not null;default:'ready';type:text;index | string | 任务状态 |
| Assignee | `assignee` | type:text | *string | 负责人 |
| UpstreamContractDetail | `upstreamContractDetail` | type:text | string | 上游契约详情(JSON) |
| DownstreamContractDetail | `downstreamContractDetail` | type:text | string | 下游契约详情(JSON) |
| Prompt | `prompt` | type:text | string | 任务提示词 |
| Tests | `tests` | type:text | string | 测试用例(JSON数组) |
| TestResult | `testResult` | type:text | string | 测试结果(JSON数组) |
| BugLog | `bugLog` | column:bug_log;type:text | string | Bug日志(JSON数组) |
| CodePaths | `codePaths` | type:text | string | 代码路径(JSON数组) |
| HumanAssistance | `humanAssistance` | type:text | string | 人工协助(JSON对象) |
| IssueDetails | `issueDetails` | column:issue_details;type:text;default:'' | string | 问题详情 |
| Locked | `locked` | not null;default:false | bool | 是否锁定 |
| LockedBy | `lockedBy` | type:text | *string | 锁定者 |
| LockedAt | `lockedAt` | type:integer | *int64 | 锁定时间 |
| LockExpiresAt | `lockExpiresAt` | type:integer | *int64 | 锁过期时间 |
| CreatedAt | `createdAt` | not null | int64 | 创建时间戳 |
| UpdatedAt | `updatedAt` | not null | int64 | 更新时间戳 |
| Version | `version` | not null;default:1 | int | 版本号 |
| SyncStatus | `syncStatus` | not null;default:'SYNCED';type:text | string | 同步状态 |

**Status 枚举值**:
- `ready` - 就绪
- `claimed` - 已认领
- `in_progress` - 进行中
- `pending_review` - 待审核
- `completed` - 已完成
- `failed` - 失败
- `blocked` - 阻塞

---

## 4. Dependency（任务依赖）

**表名**: `dependencies`

| 字段名 | JSON标签 | Go类型 | GORM约束 | 描述 |
|-------|---------|--------|---------|------|
| ID | `id` | string | primaryKey;type:text | 依赖ID |
| UpstreamTaskID | `upstreamTaskId` | string | not null;type:text;index | 上游任务ID |
| DownstreamTaskID | `downstreamTaskId` | string | not null;type:text;index | 下游任务ID |
| ContractSummary | `contractSummary` | string | type:text | 契约摘要 |
| CreatedAt | `createdAt` | int64 | not null | 创建时间戳 |
| Version | `version` | int | not null;default:1 | 版本号 |
| SyncStatus | `syncStatus` | string | not null;default:'SYNCED';type:text | 同步状态 |

---

## 5. ModuleDependency（模块依赖）

**表名**: `module_dependencies`

| 字段名 | JSON标签 | Go类型 | GORM约束 | 描述 |
|-------|---------|--------|---------|------|
| ID | `id` | string | primaryKey;type:text | 依赖ID |
| ModuleID | `moduleId` | string | not null;type:text;index | 依赖方模块ID |
| DependsOnModuleID | `dependsOnModuleId` | string | not null;type:text;index | 被依赖模块ID |
| DependencyType | `dependencyType` | string | type:text;default:'required' | 依赖类型 |
| ContractSummary | `contractSummary` | string | type:text | 契约摘要 |
| CreatedAt | `createdAt` | int64 | not null | 创建时间戳 |
| UpdatedAt | `updatedAt` | int64 | not null | 更新时间戳 |
| Version | `version` | int | not null;default:1 | 版本号 |
| SyncStatus | `syncStatus` | string | not null;default:'SYNCED';type:text | 同步状态 |

**DependencyType 枚举值**:
- `required` - 必需依赖
- `optional` - 可选依赖
- `conditional` - 条件依赖

### 扩展类型

**ModuleDependencyDetail** - 模块依赖详情（包含被依赖模块信息）:
```go
type ModuleDependencyDetail struct {
    ModuleDependency
    DependsOnModule *Module `json:"dependsOnModule"`
}
```

**ModuleDependentDetail** - 模块被依赖详情（包含依赖方模块信息）:
```go
type ModuleDependentDetail struct {
    ModuleDependency
    Module *Module `json:"module"`
}
```

---

## 6. Notification（通知）

**表名**: `notifications`

| 字段名 | JSON标签 | Go类型 | GORM约束 | 描述 |
|-------|---------|--------|---------|------|
| ID | `id` | string | primaryKey;type:text | 通知ID |
| FromTaskID | `fromTaskId` | string | not null;type:text;index | 发送方任务ID |
| ToTaskID | `toTaskId` | string | not null;type:text;index | 接收方任务ID |
| Type | `type` | string | not null;type:text | 通知类型 |
| Title | `title` | string | not null;type:text | 通知标题 |
| Content | `content` | string | type:text | 通知内容 |
| Read | `read` | bool | not null;default:false | 是否已读 |
| CreatedAt | `createdAt` | int64 | not null | 创建时间戳 |
| SyncStatus | `syncStatus` | string | not null;default:'SYNCED';type:text | 同步状态 |

**Type 枚举值**:
- `task_completed` - 任务完成
- `task_failed` - 任务失败
- `task_blocked` - 任务阻塞
- `review_required` - 需要审核
- `contract_change` - 契约变更

---

## 7. ChangeHistory（变更历史）

**表名**: `change_history`

| 字段名 | JSON标签 | Go类型 | GORM约束 | 描述 |
|-------|---------|--------|---------|------|
| ID | `id` | string | primaryKey;type:text | 记录ID |
| EntityType | `entityType` | string | not null;type:text;index | 实体类型 |
| EntityID | `entityId` | string | not null;type:text;index | 实体ID |
| Action | `action` | string | not null;type:text | 操作类型 |
| Changes | `changes` | string | type:text | 变更内容(JSON) |
| ChangedBy | `changedBy` | string | type:text | 变更者 |
| CreatedAt | `createdAt` | int64 | not null | 创建时间戳 |

**EntityType 枚举值**:
- `project` - 项目
- `module` - 模块
- `task` - 任务
- `dependency` - 依赖

**Action 枚举值**:
- `create` - 创建
- `update` - 更新
- `delete` - 删除

---

## 8. Config（配置）

**表名**: `configs`

| 字段名 | JSON标签 | Go类型 | GORM约束 | 描述 |
|-------|---------|--------|---------|------|
| ID | `id` | string | primaryKey;type:text | 配置ID |
| Key | `key` | string | not null;unique;type:text | 配置键 |
| Value | `value` | string | type:text | 配置值 |
| CreatedAt | `createdAt` | int64 | not null | 创建时间戳 |
| UpdatedAt | `updatedAt` | int64 | not null | 更新时间戳 |

**预定义配置键**:
- `plugin_type` - 插件类型
- `remote_sync_enabled` - 远程同步开关
- `remote_url` - 远程服务器URL
- `auth_token` - 认证令牌
- `server_port` - 服务器端口

---

## 9. PromptVersion（提示词版本）

**表名**: `prompt_versions`

| 字段名 | JSON标签 | Go类型 | GORM约束 | 描述 |
|-------|---------|--------|---------|------|
| ID | `id` | string | primaryKey;type:text | 版本ID |
| EntityType | `entityType` | string | not null;type:text;check:entity_type IN ('module', 'task') | 实体类型 |
| EntityID | `entityId` | string | not null;type:text;index | 实体ID |
| Version | `version` | int | not null | 版本号 |
| Prompt | `prompt` | string | not null;type:text | 提示词内容 |
| ChangeSummary | `changeSummary` | string | type:text | 变更摘要 |
| CreatedBy | `createdBy` | string | type:text | 创建者 |
| CreatedAt | `createdAt` | int64 | not null | 创建时间戳 |

---

## 字段类型说明

| Go类型 | JSON类型 | 说明 |
|--------|---------|------|
| string | string | 字符串 |
| int | number | 整数 |
| int64 | number | 64位整数(时间戳) |
| float64 | number | 浮点数 |
| bool | boolean | 布尔值 |
| *string | string/null | 可空字符串 |
| *int64 | number/null | 可空整数 |

## 通用字段

以下字段在多个模型中重复出现：

| 字段名 | 说明 |
|--------|------|
| `createdAt` | 创建时间戳(毫秒级Unix时间戳) |
| `updatedAt` | 更新时间戳(毫秒级Unix时间戳) |
| `version` | 乐观锁版本号 |
| `syncStatus` | 同步状态(用于离线同步功能) |

## 数据库关系

```
Project (1) ──── (N) Module
    │
    └─── (1) ──── (N) Task (通过 Module)

Module (N) ──── (N) Module (通过 ModuleDependency)
Task (N) ──── (N) Task (通过 Dependency)

Module (1) ──── (N) PromptVersion
Task (1) ──── (N) PromptVersion

Task (1) ──── (N) Notification
```
