# 数据库设计文档

## 概述

数据库采用SQLite，以本地文件形式存储在`.aitdd/aitdd.db`。所有表使用UUID作为主键，并包含版本和同步状态字段，以便后续分布式同步。

## 核心表结构

### 1. projects（项目信息）

| 字段名 | 类型 | 约束 | 说明 |
|--------|------|------|------|
| id | TEXT | PRIMARY KEY | UUID |
| name | TEXT | NOT NULL | 项目名称 |
| constitution | TEXT | | 项目公约（Markdown格式） |
| created_at | INTEGER | NOT NULL | 创建时间（Unix时间戳，毫秒） |
| updated_at | INTEGER | NOT NULL | 更新时间 |
| version | INTEGER | NOT NULL DEFAULT 1 | 乐观锁版本号 |
| sync_status | TEXT | NOT NULL DEFAULT 'SYNCED' | 同步状态 |

**同步状态枚举**:
- `SYNCED`: 已同步
- `PENDING_UPLOAD`: 待上传到远程
- `PENDING_DOWNLOAD`: 待从远程下载
- `CONFLICT`: 冲突，需人工解决

### 2. modules（模块）

| 字段名 | 类型 | 约束 | 说明 |
|--------|------|------|------|
| id | TEXT | PRIMARY KEY | UUID |
| parent_id | TEXT | FOREIGN KEY → modules.id | 父模块ID，可为NULL（根模块） |
| project_id | TEXT | NOT NULL, FOREIGN KEY → projects.id | 所属项目ID |
| name | TEXT | NOT NULL | 模块名称 |
| description | TEXT | | 详细描述 |
| prompt | TEXT | | 模块提示词（供AI参考） |
| status | TEXT | NOT NULL DEFAULT 'designing' | 状态枚举 |
| test_coverage | REAL | DEFAULT 0 | 测试覆盖率（0-100） |
| upstream_contract_summary | TEXT | | 上游契约概览（简短描述） |
| downstream_contract_summary | TEXT | | 下游契约概览 |
| locked | INTEGER | NOT NULL DEFAULT 0 | 0=未锁定, 1=锁定 |
| locked_by | TEXT | | 锁定者标识（AI ID或用户名） |
| locked_at | INTEGER | | 锁定时间 |
| lock_expires_at | INTEGER | | 锁过期时间 |
| created_at | INTEGER | NOT NULL | |
| updated_at | INTEGER | NOT NULL | |
| version | INTEGER | NOT NULL DEFAULT 1 | |
| sync_status | TEXT | NOT NULL DEFAULT 'SYNCED' | |

**状态枚举**:
- `designing`: 设计中
- `developing`: 开发中
- `completed`: 已完成
- `deprecated`: 已废弃

### 3. tasks（任务）

| 字段名 | 类型 | 约束 | 说明 |
|--------|------|------|------|
| id | TEXT | PRIMARY KEY | UUID |
| module_id | TEXT | NOT NULL, FOREIGN KEY → modules.id | 所属模块ID |
| name | TEXT | NOT NULL | 任务名称 |
| description | TEXT | | 详细描述 |
| status | TEXT | NOT NULL DEFAULT 'ready' | 状态枚举 |
| assignee | TEXT | | 分配给谁（AI ID或用户名） |
| upstream_contract_detail | TEXT | | JSON：上游详细契约接口 |
| downstream_contract_detail | TEXT | | JSON：下游详细契约接口 |
| prompt | TEXT | | 任务提示词（供AI实现时参考） |
| tests | TEXT | | JSON数组：测试用例列表 |
| logs | TEXT | | JSON数组：执行日志 |
| code_paths | TEXT | | JSON数组：代码文件路径（相对于项目根目录） |
| human_assistance | TEXT | | JSON对象：需要人类协助的事项 |
| locked | INTEGER | NOT NULL DEFAULT 0 | |
| locked_by | TEXT | | |
| locked_at | INTEGER | | |
| lock_expires_at | INTEGER | | |
| created_at | INTEGER | NOT NULL | |
| updated_at | INTEGER | NOT NULL | |
| version | INTEGER | NOT NULL DEFAULT 1 | |
| sync_status | TEXT | NOT NULL DEFAULT 'SYNCED' | |

**状态枚举**:
- `ready`: 就绪，可被认领
- `claimed`: 已被认领
- `in_progress`: 进行中
- `pending_review`: 等待审核
- `completed`: 已完成
- `failed`: 失败
- `blocked`: 被阻塞

**tests JSON结构**:
```json
[
  {
    "id": "test-uuid",
    "name": "测试用例名称",
    "command": "npm test login.test.js",
    "expected": "预期结果描述",
    "actual": "实际结果",
    "status": "passed|failed|pending",
    "ran_at": 1645564800000
  }
]
```

**logs JSON结构**:
```json
[
  {
    "id": "log-uuid",
    "level": "info|warn|error|debug",
    "message": "日志消息",
    "timestamp": 1645564800000,
    "metadata": {}
  }
]
```

**human_assistance JSON结构**:
```json
{
  "items": [
    {
      "id": "item-uuid",
      "description": "需要人类确认的事项",
      "status": "pending|approved|rejected",
      "reviewed_by": "username",
      "reviewed_at": 1645564800000
    }
  ],
  "all_approved": false
}
```

**contract_detail JSON结构**:
```json
{
  "interfaces": [
    {
      "name": "接口名称",
      "type": "function|class|api|cli",
      "signature": "函数签名或API路径",
      "description": "接口描述",
      "inputs": [...],
      "outputs": [...]
    }
  ],
  "data_structures": [...],
  "version": "1.0.0"
}
```

### 4. dependencies（任务依赖）

| 字段名 | 类型 | 约束 | 说明 |
|--------|------|------|------|
| id | TEXT | PRIMARY KEY | UUID |
| upstream_task_id | TEXT | NOT NULL, FOREIGN KEY → tasks.id | 上游任务ID |
| downstream_task_id | TEXT | NOT NULL, FOREIGN KEY → tasks.id | 下游任务ID |
| contract | TEXT | | JSON：依赖契约细节 |
| status | TEXT | NOT NULL DEFAULT 'active' | 依赖状态 |
| created_at | INTEGER | NOT NULL | |
| updated_at | INTEGER | NOT NULL | |
| sync_status | TEXT | NOT NULL DEFAULT 'SYNCED' | |

**UNIQUE约束**: `(upstream_task_id, downstream_task_id)`

**contract JSON结构**:
```json
{
  "interface_version": "1.0.0",
  "data_format": "json",
  "description": "依赖描述"
}
```

**status枚举**:
- `active`: 活跃
- `suspended`: 暂停
- `removed`: 已移除

### 5. notifications（任务间通知）

| 字段名 | 类型 | 约束 | 说明 |
|--------|------|------|------|
| id | TEXT | PRIMARY KEY | UUID |
| from_task_id | TEXT | FOREIGN KEY → tasks.id | 发送方任务ID |
| to_task_id | TEXT | FOREIGN KEY → tasks.id | 接收方任务ID |
| type | TEXT | NOT NULL | 通知类型 |
| title | TEXT | NOT NULL | 通知标题 |
| message | TEXT | NOT NULL | 通知内容 |
| read | INTEGER | NOT NULL DEFAULT 0 | 0=未读, 1=已读 |
| read_at | INTEGER | | 读取时间 |
| created_at | INTEGER | NOT NULL | |
| sync_status | TEXT | NOT NULL DEFAULT 'SYNCED' | |

**type枚举**:
- `task_completed`: 任务完成
- `contract_changed`: 契约变更
- `block_resolved`: 阻塞解决
- `assistance_required`: 需要协助
- `custom`: 自定义消息

### 6. change_history（修改历史）

| 字段名 | 类型 | 约束 | 说明 |
|--------|------|------|------|
| id | TEXT | PRIMARY KEY | UUID |
| table_name | TEXT | NOT NULL | 变更的表名 |
| record_id | TEXT | NOT NULL | 变更记录的ID |
| change_type | TEXT | NOT NULL | INSERT/UPDATE/DELETE |
| old_data | TEXT | | JSON：变更前数据 |
| new_data | TEXT | | JSON：变更后数据 |
| changed_by | TEXT | NOT NULL | 操作者标识 |
| changed_at | INTEGER | NOT NULL | |
| sync_version | INTEGER | NOT NULL | 用于同步的版本号 |

### 7. config（配置）

| 字段名 | 类型 | 约束 | 说明 |
|--------|------|------|------|
| key | TEXT | PRIMARY KEY | 配置键 |
| value | TEXT | NOT NULL | 配置值（JSON） |
| updated_at | INTEGER | NOT NULL | |

**预定义配置键**:
- `plugin_type`: AI插件类型
- `port`: 服务端口
- `remote_sync`: 远程同步配置
- `lock_timeout`: 锁超时时间（毫秒）

## 索引设计

```sql
-- modules 表索引
CREATE INDEX idx_modules_project_id ON modules(project_id);
CREATE INDEX idx_modules_parent_id ON modules(parent_id);
CREATE INDEX idx_modules_status ON modules(status);

-- tasks 表索引
CREATE INDEX idx_tasks_module_id ON tasks(module_id);
CREATE INDEX idx_tasks_status ON tasks(status);
CREATE INDEX idx_tasks_assignee ON tasks(assignee);

-- dependencies 表索引
CREATE INDEX idx_dependencies_upstream ON dependencies(upstream_task_id);
CREATE INDEX idx_dependencies_downstream ON dependencies(downstream_task_id);

-- notifications 表索引
CREATE INDEX idx_notifications_to_task ON notifications(to_task_id);
CREATE INDEX idx_notifications_read ON notifications(read);
CREATE INDEX idx_notifications_created ON notifications(created_at);

-- change_history 表索引
CREATE INDEX idx_history_table_record ON change_history(table_name, record_id);
CREATE INDEX idx_history_changed_at ON change_history(changed_at);
```

## 触发器

### 自动更新 updated_at

```sql
CREATE TRIGGER update_projects_timestamp 
AFTER UPDATE ON projects
BEGIN
    UPDATE projects SET updated_at = (strftime('%s', 'now') * 1000) WHERE id = NEW.id;
END;

CREATE TRIGGER update_modules_timestamp 
AFTER UPDATE ON modules
BEGIN
    UPDATE modules SET updated_at = (strftime('%s', 'now') * 1000) WHERE id = NEW.id;
END;

CREATE TRIGGER update_tasks_timestamp 
AFTER UPDATE ON tasks
BEGIN
    UPDATE tasks SET updated_at = (strftime('%s', 'now') * 1000) WHERE id = NEW.id;
END;
```

### 自动记录变更历史

```sql
CREATE TRIGGER log_tasks_update
AFTER UPDATE ON tasks
WHEN OLD.version != NEW.version
BEGIN
    INSERT INTO change_history (id, table_name, record_id, change_type, old_data, new_data, changed_by, changed_at, sync_version)
    VALUES (
        lower(hex(randomblob(16))),
        'tasks',
        NEW.id,
        'UPDATE',
        json_object('status', OLD.status, 'version', OLD.version),
        json_object('status', NEW.status, 'version', NEW.version),
        NEW.locked_by,
        strftime('%s', 'now') * 1000,
        NEW.version
    );
END;
```

### 循环依赖检测

```sql
-- 通过应用层实现，在插入dependencies时检查是否存在循环
-- 检测逻辑：使用递归CTE查询是否存在从downstream_task_id回到upstream_task_id的路径
```

## 数据库初始化SQL

```sql
-- 启用外键约束
PRAGMA foreign_keys = ON;

-- 创建表
CREATE TABLE IF NOT EXISTS projects (
    id TEXT PRIMARY KEY,
    name TEXT NOT NULL,
    constitution TEXT,
    created_at INTEGER NOT NULL,
    updated_at INTEGER NOT NULL,
    version INTEGER NOT NULL DEFAULT 1,
    sync_status TEXT NOT NULL DEFAULT 'SYNCED'
);

CREATE TABLE IF NOT EXISTS modules (
    id TEXT PRIMARY KEY,
    parent_id TEXT REFERENCES modules(id) ON DELETE CASCADE,
    project_id TEXT NOT NULL REFERENCES projects(id) ON DELETE CASCADE,
    name TEXT NOT NULL,
    description TEXT,
    prompt TEXT,
    status TEXT NOT NULL DEFAULT 'designing',
    test_coverage REAL DEFAULT 0,
    upstream_contract_summary TEXT,
    downstream_contract_summary TEXT,
    locked INTEGER NOT NULL DEFAULT 0,
    locked_by TEXT,
    locked_at INTEGER,
    lock_expires_at INTEGER,
    created_at INTEGER NOT NULL,
    updated_at INTEGER NOT NULL,
    version INTEGER NOT NULL DEFAULT 1,
    sync_status TEXT NOT NULL DEFAULT 'SYNCED'
);

CREATE TABLE IF NOT EXISTS tasks (
    id TEXT PRIMARY KEY,
    module_id TEXT NOT NULL REFERENCES modules(id) ON DELETE CASCADE,
    name TEXT NOT NULL,
    description TEXT,
    status TEXT NOT NULL DEFAULT 'ready',
    assignee TEXT,
    upstream_contract_detail TEXT,
    downstream_contract_detail TEXT,
    prompt TEXT,
    tests TEXT,
    logs TEXT,
    code_paths TEXT,
    human_assistance TEXT,
    locked INTEGER NOT NULL DEFAULT 0,
    locked_by TEXT,
    locked_at INTEGER,
    lock_expires_at INTEGER,
    created_at INTEGER NOT NULL,
    updated_at INTEGER NOT NULL,
    version INTEGER NOT NULL DEFAULT 1,
    sync_status TEXT NOT NULL DEFAULT 'SYNCED'
);

CREATE TABLE IF NOT EXISTS dependencies (
    id TEXT PRIMARY KEY,
    upstream_task_id TEXT NOT NULL REFERENCES tasks(id) ON DELETE CASCADE,
    downstream_task_id TEXT NOT NULL REFERENCES tasks(id) ON DELETE CASCADE,
    contract TEXT,
    status TEXT NOT NULL DEFAULT 'active',
    created_at INTEGER NOT NULL,
    updated_at INTEGER NOT NULL,
    sync_status TEXT NOT NULL DEFAULT 'SYNCED',
    UNIQUE(upstream_task_id, downstream_task_id)
);

CREATE TABLE IF NOT EXISTS notifications (
    id TEXT PRIMARY KEY,
    from_task_id TEXT REFERENCES tasks(id) ON DELETE SET NULL,
    to_task_id TEXT REFERENCES tasks(id) ON DELETE CASCADE,
    type TEXT NOT NULL,
    title TEXT NOT NULL,
    message TEXT NOT NULL,
    read INTEGER NOT NULL DEFAULT 0,
    read_at INTEGER,
    created_at INTEGER NOT NULL,
    sync_status TEXT NOT NULL DEFAULT 'SYNCED'
);

CREATE TABLE IF NOT EXISTS change_history (
    id TEXT PRIMARY KEY,
    table_name TEXT NOT NULL,
    record_id TEXT NOT NULL,
    change_type TEXT NOT NULL,
    old_data TEXT,
    new_data TEXT,
    changed_by TEXT NOT NULL,
    changed_at INTEGER NOT NULL,
    sync_version INTEGER NOT NULL
);

CREATE TABLE IF NOT EXISTS config (
    key TEXT PRIMARY KEY,
    value TEXT NOT NULL,
    updated_at INTEGER NOT NULL
);

-- 创建索引
CREATE INDEX idx_modules_project_id ON modules(project_id);
CREATE INDEX idx_modules_parent_id ON modules(parent_id);
CREATE INDEX idx_modules_status ON modules(status);
CREATE INDEX idx_tasks_module_id ON tasks(module_id);
CREATE INDEX idx_tasks_status ON tasks(status);
CREATE INDEX idx_tasks_assignee ON tasks(assignee);
CREATE INDEX idx_dependencies_upstream ON dependencies(upstream_task_id);
CREATE INDEX idx_dependencies_downstream ON dependencies(downstream_task_id);
CREATE INDEX idx_notifications_to_task ON notifications(to_task_id);
CREATE INDEX idx_notifications_read ON notifications(read);
CREATE INDEX idx_history_table_record ON change_history(table_name, record_id);
```

## 同步策略

### 本地优先策略

1. 所有写操作首先写入本地SQLite
2. 写入时标记`sync_status = 'PENDING_UPLOAD'`
3. 后台同步服务定期将待上传记录推送到远程

### 冲突检测

1. 每条记录有`version`字段，每次更新递增
2. 同步时比较本地和远程的`version`和`updated_at`
3. 如果远程`updated_at`更新且`version`不同，标记为`CONFLICT`

### 冲突解决

1. 前端显示冲突记录，用户选择保留本地或远程版本
2. 选择后更新本地数据并标记为`SYNCED`
