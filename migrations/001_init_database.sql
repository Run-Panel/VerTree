-- 001_init_database.sql
-- Complete database initialization with all required tables

-- Enable foreign key constraints
PRAGMA foreign_keys = ON;

-- 1. 应用程序表 (applications)
CREATE TABLE IF NOT EXISTS applications (
    app_id TEXT PRIMARY KEY,
    name TEXT NOT NULL,
    display_name TEXT NOT NULL,
    description TEXT,
    icon_url TEXT,
    is_active BOOLEAN DEFAULT true,
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    deleted_at DATETIME NULL
);

-- 2. 全局渠道表 (channels)
CREATE TABLE IF NOT EXISTS channels (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    name TEXT NOT NULL UNIQUE,
    display_name TEXT NOT NULL,
    description TEXT,
    is_active BOOLEAN DEFAULT true,
    sort_order INTEGER DEFAULT 0,
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    deleted_at DATETIME NULL
);

-- 3. 应用-渠道关联表 (application_channels) - 完整版
CREATE TABLE IF NOT EXISTS application_channels (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    app_id TEXT NOT NULL,
    channel_name TEXT NOT NULL,
    is_enabled BOOLEAN DEFAULT true,
    auto_publish BOOLEAN DEFAULT false,
    rollout_percentage INTEGER DEFAULT 100 CHECK (rollout_percentage >= 0 AND rollout_percentage <= 100),
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    deleted_at DATETIME NULL,
    UNIQUE(app_id, channel_name),
    FOREIGN KEY (app_id) REFERENCES applications(app_id) ON DELETE CASCADE,
    FOREIGN KEY (channel_name) REFERENCES channels(name) ON UPDATE CASCADE
);

-- 4. 版本表 (versions)
CREATE TABLE IF NOT EXISTS versions (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    app_id TEXT NOT NULL,
    version TEXT NOT NULL,
    channel TEXT NOT NULL DEFAULT 'stable',
    title TEXT NOT NULL,
    description TEXT,
    release_notes TEXT,
    breaking_changes TEXT,
    min_upgrade_version TEXT,
    file_url TEXT NOT NULL,
    file_size INTEGER NOT NULL,
    file_checksum TEXT NOT NULL,
    is_published BOOLEAN DEFAULT false,
    is_forced BOOLEAN DEFAULT false,
    publish_time DATETIME,
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    deleted_at DATETIME NULL,
    UNIQUE(app_id, version),
    FOREIGN KEY (app_id) REFERENCES applications(app_id) ON DELETE CASCADE
);

-- 5. 应用密钥表 (application_keys)
CREATE TABLE IF NOT EXISTS application_keys (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    app_id TEXT NOT NULL,
    name TEXT NOT NULL,
    key_value TEXT NOT NULL UNIQUE,
    is_active BOOLEAN DEFAULT true,
    permissions TEXT, -- JSON string of permissions
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    deleted_at DATETIME NULL,
    UNIQUE(app_id, name),
    FOREIGN KEY (app_id) REFERENCES applications(app_id) ON DELETE CASCADE
);

-- 6. 管理员表 (admins)
CREATE TABLE IF NOT EXISTS admins (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    username TEXT NOT NULL UNIQUE,
    email TEXT NOT NULL UNIQUE,
    password TEXT NOT NULL,
    role TEXT DEFAULT 'admin', -- admin, superadmin
    is_active BOOLEAN DEFAULT true,
    first_login BOOLEAN DEFAULT true,
    last_login DATETIME NULL,
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    deleted_at DATETIME NULL
);

-- 7. 更新统计表 (update_stats)
CREATE TABLE IF NOT EXISTS update_stats (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    version TEXT NOT NULL,
    client_id TEXT,
    client_version TEXT,
    region TEXT,
    ip_address TEXT,
    user_agent TEXT,
    action TEXT NOT NULL, -- check, download, install, success, failed
    error_message TEXT,
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    deleted_at DATETIME NULL
);

-- 创建索引
CREATE INDEX IF NOT EXISTS idx_applications_deleted_at ON applications(deleted_at);
CREATE INDEX IF NOT EXISTS idx_channels_deleted_at ON channels(deleted_at);
CREATE INDEX IF NOT EXISTS idx_application_channels_deleted_at ON application_channels(deleted_at);
CREATE INDEX IF NOT EXISTS idx_application_channels_app_id ON application_channels(app_id);
CREATE INDEX IF NOT EXISTS idx_application_channels_channel_name ON application_channels(channel_name);
CREATE INDEX IF NOT EXISTS idx_versions_deleted_at ON versions(deleted_at);
CREATE INDEX IF NOT EXISTS idx_versions_app_id ON versions(app_id);
CREATE INDEX IF NOT EXISTS idx_versions_channel ON versions(channel);
CREATE INDEX IF NOT EXISTS idx_versions_published ON versions(is_published);
CREATE INDEX IF NOT EXISTS idx_application_keys_deleted_at ON application_keys(deleted_at);
CREATE INDEX IF NOT EXISTS idx_application_keys_app_id ON application_keys(app_id);
CREATE INDEX IF NOT EXISTS idx_admins_deleted_at ON admins(deleted_at);
CREATE INDEX IF NOT EXISTS idx_update_stats_version ON update_stats(version);
CREATE INDEX IF NOT EXISTS idx_update_stats_action ON update_stats(action);
CREATE INDEX IF NOT EXISTS idx_update_stats_client_id ON update_stats(client_id);
CREATE INDEX IF NOT EXISTS idx_update_stats_created_at ON update_stats(created_at);
CREATE INDEX IF NOT EXISTS idx_update_stats_deleted_at ON update_stats(deleted_at);

-- 插入默认渠道数据
INSERT OR IGNORE INTO channels (name, display_name, description, sort_order) VALUES
('stable', '稳定版', '经过充分测试的稳定版本', 1),
('beta', '测试版', '功能完整的测试版本', 2),
('alpha', '预览版', '最新功能预览版本', 3);
