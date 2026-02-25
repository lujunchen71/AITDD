-- 006_add_module_position.sql
-- 添加模块节点位置字段

-- 添加位置字段到 modules 表
ALTER TABLE modules ADD COLUMN position_x REAL DEFAULT NULL;
ALTER TABLE modules ADD COLUMN position_y REAL DEFAULT NULL;

-- 添加位置更新时间字段（用于调试和审计）
ALTER TABLE modules ADD COLUMN position_updated_at INTEGER DEFAULT NULL;

-- 创建位置索引（可选，用于按位置查询）
CREATE INDEX IF NOT EXISTS idx_modules_position ON modules(position_x, position_y);
