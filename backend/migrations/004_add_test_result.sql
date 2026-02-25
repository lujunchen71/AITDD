-- 004_add_test_result.sql
-- 添加 test_result 字段到 tasks 表，用于存储测试证据字符串数组

ALTER TABLE tasks ADD COLUMN test_result TEXT;

-- 更新注释
COMMENT ON COLUMN tasks.test_result IS 'JSON array of test evidence strings';
