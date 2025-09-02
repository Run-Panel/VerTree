-- Migration: Add GitHub Integration Tables (SQLite)
-- Description: Add tables for GitHub repository binding and file caching (SQLite compatible)
-- Created: 2024-01-20

-- Create github_repositories table
CREATE TABLE IF NOT EXISTS github_repositories (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    app_id VARCHAR(32) NOT NULL,
    repository_url VARCHAR(500) NOT NULL,
    owner_name VARCHAR(100) NOT NULL,
    repo_name VARCHAR(100) NOT NULL,
    branch_name VARCHAR(100) NOT NULL DEFAULT 'main',
    access_token VARCHAR(500), -- Encrypted GitHub token
    webhook_secret VARCHAR(100), -- Secret for webhook validation
    webhook_id INTEGER DEFAULT 0,
    is_active BOOLEAN DEFAULT true,
    auto_sync BOOLEAN DEFAULT true,
    auto_publish BOOLEAN DEFAULT false,
    default_channel VARCHAR(20) DEFAULT 'stable',
    last_sync_at DATETIME,
    last_sync_status VARCHAR(50) DEFAULT 'pending',
    last_sync_error TEXT,
    sync_count INTEGER DEFAULT 0,
    created_by INTEGER NOT NULL,
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    deleted_at DATETIME,
    
    -- Foreign key constraints
    FOREIGN KEY (app_id) REFERENCES applications(app_id) ON DELETE CASCADE,
    FOREIGN KEY (created_by) REFERENCES admins(id) ON DELETE RESTRICT
);

-- Create indexes for github_repositories
CREATE INDEX IF NOT EXISTS idx_github_repositories_app_id ON github_repositories(app_id);
CREATE INDEX IF NOT EXISTS idx_github_repositories_owner_repo ON github_repositories(owner_name, repo_name);
CREATE INDEX IF NOT EXISTS idx_github_repositories_deleted_at ON github_repositories(deleted_at);
CREATE INDEX IF NOT EXISTS idx_github_repositories_sync_status ON github_repositories(last_sync_status);

-- Create github_releases table
CREATE TABLE IF NOT EXISTS github_releases (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    repository_id INTEGER NOT NULL,
    release_id INTEGER NOT NULL,
    tag_name VARCHAR(100) NOT NULL,
    release_name VARCHAR(200) NOT NULL,
    body TEXT,
    is_prerelease BOOLEAN DEFAULT false,
    is_draft BOOLEAN DEFAULT false,
    published_at DATETIME,
    download_url VARCHAR(500),
    file_size INTEGER DEFAULT 0,
    file_checksum VARCHAR(128),
    local_file_path VARCHAR(500),
    sync_status VARCHAR(50) DEFAULT 'pending',
    version_id INTEGER, -- Link to versions table
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    deleted_at DATETIME,
    
    -- Constraints
    FOREIGN KEY (repository_id) REFERENCES github_repositories(id) ON DELETE CASCADE,
    FOREIGN KEY (version_id) REFERENCES versions(id) ON DELETE SET NULL,
    UNIQUE(repository_id, release_id)
);

-- Create indexes for github_releases
CREATE INDEX IF NOT EXISTS idx_github_releases_repository_id ON github_releases(repository_id);
CREATE INDEX IF NOT EXISTS idx_github_releases_version_id ON github_releases(version_id);
CREATE INDEX IF NOT EXISTS idx_github_releases_sync_status ON github_releases(sync_status);
CREATE INDEX IF NOT EXISTS idx_github_releases_published_at ON github_releases(published_at);
CREATE INDEX IF NOT EXISTS idx_github_releases_deleted_at ON github_releases(deleted_at);

-- Create file_cache table
CREATE TABLE IF NOT EXISTS file_cache (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    app_id VARCHAR(32) NOT NULL,
    version VARCHAR(100) NOT NULL,
    original_url VARCHAR(500) NOT NULL,
    local_path VARCHAR(500) NOT NULL,
    file_size INTEGER NOT NULL,
    file_checksum VARCHAR(128) NOT NULL,
    content_type VARCHAR(100),
    downloaded_at DATETIME NOT NULL,
    last_accessed DATETIME NOT NULL,
    access_count INTEGER DEFAULT 0,
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    
    -- Constraints
    FOREIGN KEY (app_id) REFERENCES applications(app_id) ON DELETE CASCADE
);

-- Create indexes for file_cache
CREATE INDEX IF NOT EXISTS idx_file_cache_app_id ON file_cache(app_id);
CREATE INDEX IF NOT EXISTS idx_file_cache_version ON file_cache(version);
CREATE INDEX IF NOT EXISTS idx_file_cache_last_accessed ON file_cache(last_accessed);
CREATE INDEX IF NOT EXISTS idx_file_cache_checksum ON file_cache(file_checksum);

