-- 添加 issue_details 列
ALTER TABLE tasks ADD COLUMN issue_details TEXT DEFAULT '';

-- 重命名 logs 列为 bug_log（SQLite 不支持直接重命名，需要重建表）
-- 首先创建新表
CREATE TABLE tasks_new (
    id TEXT PRIMARY KEY,
    module_id TEXT NOT NULL,
    name TEXT NOT NULL,
    description TEXT,
    status TEXT DEFAULT 'ready',
    assignee TEXT,
    upstream_contract_detail TEXT,
    downstream_contract_detail TEXT,
    prompt TEXT,
    tests TEXT,
    test_result TEXT,
    bug_log TEXT,  -- 原来的 logs
    code_paths TEXT,
    human_assistance TEXT,
    issue_details TEXT DEFAULT '',  -- 新增字段
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

-- 复制数据（logs -> bug_log）
INSERT INTO tasks_new 
SELECT id, module_id, name, description, status, assignee, 
       upstream_contract_detail, downstream_contract_detail, prompt, 
       tests, test_result, logs, -- logs 复制到 bug_log
       code_paths, human_assistance, 
       '', -- issue_details 默认空
       locked, locked_by, locked_at, lock_expires_at, 
       created_at, updated_at, version, sync_status
FROM tasks;

-- 删除旧表
DROP TABLE tasks;

-- 重命名新表
ALTER TABLE tasks_new RENAME TO tasks;

-- 重建索引
CREATE INDEX idx_tasks_module_id ON tasks(module_id);
CREATE INDEX idx_tasks_status ON tasks(status);
