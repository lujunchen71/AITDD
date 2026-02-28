# Issue 问答系统数据模型设计

> 基于 `mcp-compiler-design.md` 设计文档，设计问答系统的数据模型

---

## 1. 需求分析

### 1.1 Issue 功能需求

根据设计文档，Issue 系统需要支持：

| 操作 | 说明 |
|------|------|
| `create_issue` | 下游任务发现上游有问题时发起 |
| `reply_issue` | 对问题进行回复 |
| `resolve_issue` | 标记问题已解决 |
| `query_issues` | 按状态、类型、方向过滤查询问题列表 |

### 1.2 Issue 属性

| 属性 | 类型 | 说明 |
|------|------|------|
| fromTaskPathName | string | 发起方任务 pathName |
| toTaskPathName | string | 接收方任务 pathName |
| type | string | contract / test / other |
| title | string | 问题标题 |
| status | string | pending / replied / resolved |
| messages | []Message | 消息列表（包含初始内容和后续回复） |
| createdAt | timestamp | 创建时间 |
| updatedAt | timestamp | 更新时间 |

### 1.3 Message 消息结构

| 属性 | 类型 | 说明 |
|------|------|------|
| sender | string | 发送方标识（taskPathName 或 "system"） |
| timestamp | int64 | 时间戳（毫秒） |
| content | string | 消息内容 |
| read | bool | 是否已读 |

---

## 2. 数据模型设计

### 2.1 设计决策

**方案选择**：创建独立的 `issues` 表

**理由**：
1. 设计文档明确要求"新增 `issues` 表，独立的问题追踪，替代 `tasks.issue_details` 字段"
2. 独立表便于高效查询和索引
3. 支持复杂的状态管理和消息链
4. 与现有 Notification 模型保持一致的架构风格

### 2.2 Message 结构体（嵌入式）

Message 作为 Issue 中的嵌入式结构，存储为 JSON 字符串：

```go
// Message 消息结构 - 用于 Issue 中的消息列表
type Message struct {
    Sender    string `json:"sender"`    // 发送方标识（taskPathName 或 "system"）
    Timestamp int64  `json:"timestamp"` // 时间戳（毫秒）
    Content   string `json:"content"`   // 消息内容
    Read      bool   `json:"read"`      // 是否已读
}
```

### 2.3 Issue 模型结构体

```go
package models

import (
    "time"

    "gorm.io/gorm"
)

// Issue 问题/问答
type Issue struct {
    ID               string    `json:"id" gorm:"primaryKey;type:text"`
    FromTaskPathName string    `json:"fromTaskPathName" gorm:"not null;type:text;index"` // 发起方任务 pathName
    ToTaskPathName   string    `json:"toTaskPathName" gorm:"not null;type:text;index"`   // 接收方任务 pathName
    Type             string    `json:"type" gorm:"not null;type:text;index"`             // contract / test / other
    Title            string    `json:"title" gorm:"not null;type:text"`                  // 问题标题
    Status           string    `json:"status" gorm:"not null;default:'pending';type:text;index"` // pending / replied / resolved
    Messages         string    `json:"messages" gorm:"type:text"`                        // JSON array of Message objects
    CreatedAt        int64     `json:"createdAt" gorm:"not null"`
    UpdatedAt        int64     `json:"updatedAt" gorm:"not null"`
    Version          int       `json:"version" gorm:"not null;default:1"`
    SyncStatus       string    `json:"syncStatus" gorm:"not null;default:'SYNCED';type:text"`
}

// IssueType 问题类型枚举
const (
    IssueTypeContract = "contract" // 契约问题
    IssueTypeTest     = "test"     // 测试问题
    IssueTypeOther    = "other"    // 其他问题
)

// IssueStatus 问题状态枚举
const (
    IssueStatusPending  = "pending"  // 待处理
    IssueStatusReplied  = "replied"  // 已回复
    IssueStatusResolved = "resolved" // 已解决
)

// BeforeCreate 创建前钩子
func (i *Issue) BeforeCreate(_ *gorm.DB) error {
    if i.CreatedAt == 0 {
        i.CreatedAt = time.Now().UnixMilli()
    }
    if i.UpdatedAt == 0 {
        i.UpdatedAt = time.Now().UnixMilli()
    }
    return nil
}

// BeforeUpdate 更新前钩子
func (i *Issue) BeforeUpdate(_ *gorm.DB) error {
    i.UpdatedAt = time.Now().UnixMilli()
    return nil
}

// TableName 指定表名
func (Issue) TableName() string {
    return "issues"
}
```

### 2.4 完整文件：`backend/internal/models/issue.go`

```go
package models

import (
    "time"

    "gorm.io/gorm"
)

// Message 消息结构 - 用于 Issue 中的消息列表
// 注意：这是一个值类型，不对应独立的数据库表
// 存储时会序列化为 JSON 字符串
type Message struct {
    Sender    string `json:"sender"`    // 发送方标识（taskPathName 或 "system"）
    Timestamp int64  `json:"timestamp"` // 时间戳（毫秒）
    Content   string `json:"content"`   // 消息内容
    Read      bool   `json:"read"`      // 是否已读
}

// Issue 问题/问答
// 用于任务之间的问答追踪
type Issue struct {
    ID               string `json:"id" gorm:"primaryKey;type:text"`
    FromTaskPathName string `json:"fromTaskPathName" gorm:"not null;type:text;index"` // 发起方任务 pathName
    ToTaskPathName   string `json:"toTaskPathName" gorm:"not null;type:text;index"`   // 接收方任务 pathName
    Type             string `json:"type" gorm:"not null;type:text;index"`             // contract / test / other
    Title            string `json:"title" gorm:"not null;type:text"`                  // 问题标题
    Status           string `json:"status" gorm:"not null;default:'pending';type:text;index"` // pending / replied / resolved
    Messages         string `json:"messages" gorm:"type:text"`                        // JSON array of Message objects
    CreatedAt        int64  `json:"createdAt" gorm:"not null"`
    UpdatedAt        int64  `json:"updatedAt" gorm:"not null"`
    Version          int    `json:"version" gorm:"not null;default:1"`
    SyncStatus       string `json:"syncStatus" gorm:"not null;default:'SYNCED';type:text"`
}

// IssueType 问题类型枚举
const (
    IssueTypeContract = "contract" // 契约问题
    IssueTypeTest     = "test"     // 测试问题
    IssueTypeOther    = "other"    // 其他问题
)

// IssueStatus 问题状态枚举
const (
    IssueStatusPending  = "pending"  // 待处理
    IssueStatusReplied  = "replied"  // 已回复
    IssueStatusResolved = "resolved" // 已解决
)

// BeforeCreate 创建前钩子
func (i *Issue) BeforeCreate(_ *gorm.DB) error {
    if i.CreatedAt == 0 {
        i.CreatedAt = time.Now().UnixMilli()
    }
    if i.UpdatedAt == 0 {
        i.UpdatedAt = time.Now().UnixMilli()
    }
    return nil
}

// BeforeUpdate 更新前钩子
func (i *Issue) BeforeUpdate(_ *gorm.DB) error {
    i.UpdatedAt = time.Now().UnixMilli()
    return nil
}

// TableName 指定表名
func (Issue) TableName() string {
    return "issues"
}
```

---

## 3. 数据库迁移

### 3.1 新建迁移文件：`backend/migrations/008_create_issues.sql`

```sql
-- 创建 issues 表
CREATE TABLE IF NOT EXISTS issues (
    id TEXT PRIMARY KEY,
    from_task_path_name TEXT NOT NULL,
    to_task_path_name TEXT NOT NULL,
    type TEXT NOT NULL,
    title TEXT NOT NULL,
    status TEXT NOT NULL DEFAULT 'pending',
    messages TEXT,
    created_at INTEGER NOT NULL,
    updated_at INTEGER NOT NULL,
    version INTEGER NOT NULL DEFAULT 1,
    sync_status TEXT NOT NULL DEFAULT 'SYNCED'
);

-- 创建索引
CREATE INDEX IF NOT EXISTS idx_issues_from_task ON issues(from_task_path_name);
CREATE INDEX IF NOT EXISTS idx_issues_to_task ON issues(to_task_path_name);
CREATE INDEX IF NOT EXISTS idx_issues_status ON issues(status);
CREATE INDEX IF NOT EXISTS idx_issues_type ON issues(type);

-- 可选：迁移现有 tasks.issue_details 数据到 issues 表
-- 由于 issue_details 是非结构化文本，需要根据实际数据格式进行迁移
-- 这里提供一个基本的迁移模板，需要根据实际情况调整
-- INSERT INTO issues (id, from_task_path_name, to_task_path_name, type, title, status, messages, created_at, updated_at)
-- SELECT 
--     lower(hex(randomblob(16))),  -- 生成随机 UUID
--     '',  -- 需要从 issue_details 解析
--     t.path_name,  -- 当前任务的 pathName
--     'other',  -- 默认类型
--     'Migrated Issue',  -- 默认标题
--     'pending',  -- 默认状态
--     t.issue_details,  -- 原始内容作为第一条消息
--     strftime('%s', 'now') * 1000,
--     strftime('%s', 'now') * 1000
-- FROM tasks t
-- WHERE t.issue_details IS NOT NULL AND t.issue_details != '';
```

### 3.2 清理旧字段（可选）：`backend/migrations/009_remove_issue_details.sql`

```sql
-- 删除 tasks 表中的 issue_details 字段
-- 注意：SQLite 不支持 DROP COLUMN，需要重建表

-- 首先确认 issues 表已创建并迁移完成
-- 然后执行以下操作：

-- 创建新表（不含 issue_details）
CREATE TABLE tasks_new (
    id TEXT PRIMARY KEY,
    module_id TEXT NOT NULL,
    name TEXT NOT NULL,
    path_name TEXT UNIQUE,
    description TEXT,
    status TEXT DEFAULT 'ready',
    assignee TEXT,
    upstream_contract_detail TEXT,
    downstream_contract_detail TEXT,
    prompt TEXT,
    tests TEXT,
    test_result TEXT,
    bug_log TEXT,
    code_paths TEXT,
    human_assistance TEXT,
    locked INTEGER DEFAULT 0,
    locked_by TEXT,
    locked_at INTEGER,
    lock_expires_at INTEGER,
    created_at INTEGER,
    updated_at INTEGER,
    version INTEGER DEFAULT 1,
    sync_status TEXT DEFAULT 'SYNCED',
    FOREIGN KEY (module_id) REFERENCES modules(id)
);

-- 复制数据（排除 issue_details）
INSERT INTO tasks_new 
SELECT id, module_id, name, path_name, description, status, assignee, 
       upstream_contract_detail, downstream_contract_detail, prompt, 
       tests, test_result, bug_log, code_paths, human_assistance,
       locked, locked_by, locked_at, lock_expires_at, 
       created_at, updated_at, version, sync_status
FROM tasks;

-- 删除旧表
DROP TABLE tasks;

-- 重命名新表
ALTER TABLE tasks_new RENAME TO tasks;

-- 重建索引
CREATE INDEX IF NOT EXISTS idx_tasks_module_id ON tasks(module_id);
CREATE INDEX IF NOT EXISTS idx_tasks_status ON tasks(status);
CREATE INDEX IF NOT EXISTS idx_tasks_path_name ON tasks(path_name);
```

---

## 4. 模型关系图

```mermaid
erDiagram
    PROJECT ||--o{ MODULE : contains
    MODULE ||--o{ TASK : contains
    MODULE }o--o{ MODULE : depends
    TASK }o--o{ TASK : depends
    TASK ||--o{ NOTIFICATION : receives
    TASK ||--o{ ISSUE : from
    TASK ||--o{ ISSUE : to

    PROJECT {
        text id PK
        text name
        text pathName UK
        text constitution
        integer createdAt
        integer updatedAt
        integer version
        text syncStatus
    }

    MODULE {
        text id PK
        text parentId FK
        text projectId FK
        text name
        text pathName UK
        text description
        text prompt
        text status
        real testCoverage
        text upstreamContractSummary
        text downstreamContractSummary
        integer locked
        text lockedBy
        integer lockedAt
        integer lockExpiresAt
        real positionX
        real positionY
        integer positionUpdatedAt
        integer createdAt
        integer updatedAt
        integer version
        text syncStatus
    }

    TASK {
        text id PK
        text moduleId FK
        text name
        text pathName UK
        text description
        text status
        text assignee
        text upstreamContractDetail
        text downstreamContractDetail
        text prompt
        text tests
        text testResult
        text bugLog
        text codePaths
        text humanAssistance
        integer locked
        text lockedBy
        integer lockedAt
        integer lockExpiresAt
        integer createdAt
        integer updatedAt
        integer version
        text syncStatus
    }

    NOTIFICATION {
        text id PK
        text fromTaskId FK
        text toTaskId FK
        text type
        text title
        text content
        integer read
        integer createdAt
        text syncStatus
    }

    ISSUE {
        text id PK
        text fromTaskPathName
        text toTaskPathName
        text type
        text title
        text status
        text messages
        integer createdAt
        integer updatedAt
        integer version
        text syncStatus
    }
```

---

## 5. API 使用示例

### 5.1 创建 Issue

```json
// MCP Tool: create_issue
{
    "fromTaskPathName": "aitdd/auth/register",
    "toTaskPathName": "aitdd/auth/login",
    "type": "contract",
    "title": "登录接口返回格式不一致",
    "content": "下游任务发现登录接口返回的 JSON 格式与契约定义不符"
}

// 数据库存储
{
    "id": "issue-uuid-001",
    "fromTaskPathName": "aitdd/auth/register",
    "toTaskPathName": "aitdd/auth/login",
    "type": "contract",
    "title": "登录接口返回格式不一致",
    "status": "pending",
    "messages": "[{\"sender\":\"aitdd/auth/register\",\"timestamp\":1709100000000,\"content\":\"下游任务发现登录接口返回的 JSON 格式与契约定义不符\",\"read\":false}]",
    "createdAt": 1709100000000,
    "updatedAt": 1709100000000,
    "version": 1,
    "syncStatus": "SYNCED"
}
```

### 5.2 回复 Issue

```json
// MCP Tool: reply_issue
{
    "fromTaskPathName": "aitdd/auth/register",
    "toTaskPathName": "aitdd/auth/login",
    "title": "登录接口返回格式不一致",
    "replyContent": "收到，我会检查并修复这个问题"
}

// 更新后的 messages 字段
[
    {"sender":"aitdd/auth/register","timestamp":1709100000000,"content":"下游任务发现登录接口返回的 JSON 格式与契约定义不符","read":true},
    {"sender":"aitdd/auth/login","timestamp":1709100100000,"content":"收到，我会检查并修复这个问题","read":false}
]
// status 变为 "replied"
```

### 5.3 解决 Issue

```json
// MCP Tool: resolve_issue
{
    "fromTaskPathName": "aitdd/auth/register",
    "toTaskPathName": "aitdd/auth/login",
    "title": "登录接口返回格式不一致"
}

// 更新后的数据
{
    "status": "resolved",
    "updatedAt": 1709100200000,
    "version": 3
}
```

### 5.4 查询 Issues

```json
// MCP Tool: query_issues
{
    "taskPathName": "aitdd/auth/login",
    "direction": "to",
    "status": "pending"
}

// 响应
{
    "issues": [
        {
            "fromTaskPathName": "aitdd/auth/register",
            "fromTaskName": "用户注册",
            "toTaskPathName": "aitdd/auth/login",
            "toTaskName": "用户登录",
            "type": "contract",
            "title": "登录接口返回格式不一致",
            "status": "pending",
            "createdAt": 1709100000000
        }
    ],
    "total": 1
}
```

---

## 6. 实现注意事项

### 6.1 Messages 字段处理

```go
// 辅助函数：解析 Messages
func ParseMessages(messagesJSON string) ([]Message, error) {
    if messagesJSON == "" {
        return []Message{}, nil
    }
    var messages []Message
    err := json.Unmarshal([]byte(messagesJSON), &messages)
    return messages, err
}

// 辅助函数：序列化 Messages
func SerializeMessages(messages []Message) (string, error) {
    if len(messages) == 0 {
        return "[]", nil
    }
    bytes, err := json.Marshal(messages)
    return string(bytes), err
}

// 辅助函数：添加消息
func AddMessage(issue *Issue, sender, content string) error {
    messages, err := ParseMessages(issue.Messages)
    if err != nil {
        return err
    }
    
    newMessage := Message{
        Sender:    sender,
        Timestamp: time.Now().UnixMilli(),
        Content:   content,
        Read:      false,
    }
    
    messages = append(messages, newMessage)
    
    issue.Messages, err = SerializeMessages(messages)
    if err != nil {
        return err
    }
    
    issue.Version++
    return nil
}
```

### 6.2 Issue 唯一性

同一对任务之间可以有多个 Issue（不同标题或类型），不需要唯一约束。

### 6.3 索引策略

| 索引 | 用途 |
|------|------|
| `idx_issues_from_task` | 查询某任务发起的所有 Issue |
| `idx_issues_to_task` | 查询某任务接收的所有 Issue |
| `idx_issues_status` | 按状态过滤 |
| `idx_issues_type` | 按类型过滤 |

---

## 7. 与现有模型对比

| 特性 | Notification | Issue |
|------|--------------|-------|
| 关系类型 | 单向（from → to） | 双向对话 |
| 消息数量 | 单条 | 多条（消息链） |
| 状态管理 | read/unread | pending/replied/resolved |
| 持久性 | 通知后可删除 | 需要长期保留 |
| 用途 | 事件通知 | 问题追踪与解决 |

---

## 8. 后续工作

1. **Service 层实现**：创建 `backend/internal/services/issue_service.go`
2. **Handler 层实现**：创建 `backend/internal/api/handlers/issue.go`
3. **MCP Tool 实现**：在 `backend/internal/mcp/` 中添加 Issue 相关工具
4. **前端集成**：在任务详情页面展示相关 Issues
5. **数据迁移**：将现有 `tasks.issue_details` 迁移到新表
