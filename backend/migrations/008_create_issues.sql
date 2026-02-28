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
