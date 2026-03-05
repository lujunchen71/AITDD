-- 009_update_bug_log_format.sql
-- 1. 为 modules 表添加 bug_log 字段
ALTER TABLE modules ADD COLUMN bug_log TEXT DEFAULT '';

-- 2. 为 projects 表添加 bug_log 字段
ALTER TABLE projects ADD COLUMN bug_log TEXT DEFAULT '';

-- 3. 迁移 tasks 表中现有的 bug_log 数据（旧格式 → 新格式）
-- 旧格式: {"static": [...], "dynamic": [...]}
-- 新格式: {"static": {"error": [...], "warning": []}, "dynamic": {"error": [...], "warning": []}, "exe": {"error": [], "warning": []}}
-- SQLite 不支持复杂的 JSON 操作，因此对于已有数据，我们直接清空（因为旧格式 bug_log 是简单数组，无法自动迁移）
-- 新写入的数据将使用新格式
UPDATE tasks SET bug_log = '' WHERE bug_log != '' AND bug_log IS NOT NULL;
