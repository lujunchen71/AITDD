-- 002_module_dependencies.sql
-- 添加模块依赖关系表

-- 模块依赖关系表
CREATE TABLE IF NOT EXISTS module_dependencies (
    id TEXT PRIMARY KEY,
    module_id TEXT NOT NULL,              -- 依赖方模块ID
    depends_on_module_id TEXT NOT NULL,   -- 被依赖模块ID
    dependency_type TEXT DEFAULT 'required', -- 依赖类型: required, optional, conditional
    contract_summary TEXT,                -- 合约摘要
    created_at INTEGER NOT NULL,
    updated_at INTEGER NOT NULL,
    version INTEGER NOT NULL DEFAULT 1,
    sync_status TEXT NOT NULL DEFAULT 'SYNCED',
    FOREIGN KEY (module_id) REFERENCES modules(id) ON DELETE CASCADE,
    FOREIGN KEY (depends_on_module_id) REFERENCES modules(id) ON DELETE CASCADE
);

-- 创建索引
CREATE INDEX IF NOT EXISTS idx_module_dependencies_module_id ON module_dependencies(module_id);
CREATE INDEX IF NOT EXISTS idx_module_dependencies_depends_on ON module_dependencies(depends_on_module_id);

-- 添加唯一约束，防止重复依赖
CREATE UNIQUE INDEX IF NOT EXISTS idx_module_dependencies_unique ON module_dependencies(module_id, depends_on_module_id);
