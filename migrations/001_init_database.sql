-- 001_init_database.sql
-- Database initialization with default data only
-- Tables are created by GORM AutoMigrate, this file only handles data seeding

-- Enable foreign key constraints
PRAGMA foreign_keys = ON;

-- 插入默认渠道数据
INSERT OR IGNORE INTO channels (name, display_name, description, sort_order) VALUES
('stable', '稳定版', '经过充分测试的稳定版本', 1),
('beta', '测试版', '功能完整的测试版本', 2),
('alpha', '预览版', '最新功能预览版本', 3);
