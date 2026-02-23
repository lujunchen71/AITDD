-- 001_init.sql
-- AITDD 数据库初始化迁移

-- 项目表
CREATE TABLE IF NOT EXISTS projects (
    id TEXT PRIMARY KEY,
    name TEXT NOT NULL,
    constitution TEXT,
    created_at INTEGER NOT NULL,
    updated_at INTEGER NOT NULL,
    version INTEGER NOT NULL DEFAULT 1,
    sync_status TEXT NOT NULL DEFAULT 'SYNCED'
);

-- 模块表
CREATE TABLE IF NOT EXISTS modules (
    id TEXT PRIMARY KEY,
    parent_id TEXT,
    project_id TEXT NOT NULL,
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
    sync_status TEXT NOT NULL DEFAULT 'SYNCED',
    FOREIGN KEY (parent_id) REFERENCES modules(id) ON DELETE CASCADE,
    FOREIGN KEY (project_id) REFERENCES projects(id) ON DELETE CASCADE
);

-- 任务表
CREATE TABLE IF NOT EXISTS tasks (
    id TEXT PRIMARY KEY,
    module_id TEXT NOT NULL,
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
    sync_status TEXT NOT NULL DEFAULT 'SYNCED',
    FOREIGN KEY (module_id) REFERENCES modules(id) ON DELETE CASCADE
);

-- 依赖关系表
CREATE TABLE IF NOT EXISTS dependencies (
    id TEXT PRIMARY KEY,
    upstream_task_id TEXT NOT NULL,
    downstream_task_id TEXT NOT NULL,
    contract_summary TEXT,
    created_at INTEGER NOT NULL,
    version INTEGER NOT NULL DEFAULT 1,
    sync_status TEXT NOT NULL DEFAULT 'SYNCED',
    FOREIGN KEY (upstream_task_id) REFERENCES tasks(id) ON DELETE CASCADE,
    FOREIGN KEY (downstream_task_id) REFERENCES tasks(id) ON DELETE CASCADE
);

-- 通知表
CREATE TABLE IF NOT EXISTS notifications (
    id TEXT PRIMARY KEY,
    from_task_id TEXT NOT NULL,
    to_task_id TEXT NOT NULL,
    type TEXT NOT NULL,
    title TEXT NOT NULL,
    content TEXT,
    read INTEGER NOT NULL DEFAULT 0,
    created_at INTEGER NOT NULL,
    sync_status TEXT NOT NULL DEFAULT 'SYNCED',
    FOREIGN KEY (from_task_id) REFERENCES tasks(id) ON DELETE CASCADE,
    FOREIGN KEY (to_task_id) REFERENCES tasks(id) ON DELETE CASCADE
);

-- 变更历史表
CREATE TABLE IF NOT EXISTS change_history (
    id TEXT PRIMARY KEY,
    entity_type TEXT NOT NULL,
    entity_id TEXT NOT NULL,
    action TEXT NOT NULL,
    changes TEXT,
    changed_by TEXT,
    created_at INTEGER NOT NULL
);

-- 配置表
CREATE TABLE IF NOT EXISTS configs (
    id TEXT PRIMARY KEY,
    key TEXT NOT NULL UNIQUE,
    value TEXT,
    created_at INTEGER NOT NULL,
    updated_at INTEGER NOT NULL
);

-- 创建索引
CREATE INDEX IF NOT EXISTS idx_modules_parent_id ON modules(parent_id);
CREATE INDEX IF NOT EXISTS idx_modules_project_id ON modules(project_id);
CREATE INDEX IF NOT EXISTS idx_modules_status ON modules(status);
CREATE INDEX IF NOT EXISTS idx_tasks_module_id ON tasks(module_id);
CREATE INDEX IF NOT EXISTS idx_tasks_status ON tasks(status);
CREATE INDEX IF NOT EXISTS idx_dependencies_upstream ON dependencies(upstream_task_id);
CREATE INDEX IF NOT EXISTS idx_dependencies_downstream ON dependencies(downstream_task_id);
CREATE INDEX IF NOT EXISTS idx_notifications_to_task ON notifications(to_task_id);
CREATE INDEX IF NOT EXISTS idx_notifications_from_task ON notifications(from_task_id);
CREATE INDEX IF NOT EXISTS idx_change_history_entity ON change_history(entity_type, entity_id);
