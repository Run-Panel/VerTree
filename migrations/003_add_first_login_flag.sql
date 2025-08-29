-- 003_add_first_login_flag.sql
-- 为已存在的管理员用户设置首次登录标志，确保数据一致性
-- 这个迁移在 GORM AutoMigrate 已经创建表结构后执行

-- 为已存在但 first_login 字段为空的管理员用户设置首次登录标志
UPDATE admins SET first_login = true WHERE first_login IS NULL;