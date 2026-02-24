-- 提示词版本表
-- 用于追踪模块和任务的提示词变更历史

CREATE TABLE IF NOT EXISTS prompt_versions (
    id TEXT PRIMARY KEY,
    entity_type TEXT NOT NULL CHECK (entity_type IN ('module', 'task')),
    entity_id TEXT NOT NULL,
    version INTEGER NOT NULL,
    prompt TEXT NOT NULL,
    change_summary TEXT,
    created_by TEXT,
    created_at INTEGER NOT NULL
);

-- 创建索引
CREATE INDEX IF NOT EXISTS idx_prompt_versions_entity ON prompt_versions(entity_type, entity_id);
CREATE INDEX IF NOT EXISTS idx_prompt_versions_created ON prompt_versions(created_at);

-- 添加说明注释
COMMENT ON TABLE prompt_versions IS '提示词版本表 - 追踪模块和任务的提示词变更历史';
COMMENT ON COLUMN prompt_versions.entity_type IS '实体类型：module 或 task';
COMMENT ON COLUMN prompt_versions.entity_id IS '实体 ID：模块 ID 或任务 ID';
COMMENT ON COLUMN prompt_versions.version IS '版本号，从 1 开始递增';
COMMENT ON COLUMN prompt_versions.prompt IS '提示词内容';
COMMENT ON COLUMN prompt_versions.change_summary IS '变更摘要';
COMMENT ON COLUMN prompt_versions.created_by IS '创建者标识';
COMMENT ON COLUMN prompt_versions.created_at IS '创建时间 (Unix 时间戳，毫秒)';
