-- 为已存在的管理员用户设置首次登录标志
UPDATE admins SET first_login = true WHERE first_login IS NULL;