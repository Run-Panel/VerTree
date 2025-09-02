-- Migration: Add GitHub Integration Tables
-- Description: Add tables for GitHub repository binding and file caching
-- Created: 2024-01-20

-- Create github_repositories table
-- GitHub repository bindings for applications
CREATE TABLE IF NOT EXISTS github_repositories (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    app_id TEXT NOT NULL,
    repository_url TEXT NOT NULL,
    owner_name TEXT NOT NULL,
    repo_name TEXT NOT NULL,
    branch_name TEXT NOT NULL DEFAULT 'main',
    access_token TEXT, -- Encrypted GitHub personal access token or app token
    webhook_secret TEXT, -- Secret used to validate GitHub webhook signatures
    webhook_id INTEGER DEFAULT 0,
    is_active INTEGER DEFAULT 1, -- SQLite boolean: 1=true, 0=false
    auto_sync INTEGER DEFAULT 1, -- Automatically sync new releases from GitHub
    auto_publish INTEGER DEFAULT 0, -- Automatically publish new versions when synced
    default_channel TEXT DEFAULT 'stable',
    last_sync_at DATETIME,
    last_sync_status TEXT DEFAULT 'pending',
    last_sync_error TEXT,
    sync_count INTEGER DEFAULT 0,
    created_by INTEGER NOT NULL,
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    deleted_at DATETIME,
    
    -- Create foreign key constraints
    FOREIGN KEY (app_id) REFERENCES applications(app_id) ON DELETE CASCADE,
    FOREIGN KEY (created_by) REFERENCES admins(id) ON DELETE RESTRICT
);

-- Create indexes for github_repositories
CREATE INDEX IF NOT EXISTS idx_github_repositories_app_id ON github_repositories(app_id);
CREATE INDEX IF NOT EXISTS idx_github_repositories_owner_repo ON github_repositories(owner_name, repo_name);
CREATE INDEX IF NOT EXISTS idx_github_repositories_deleted_at ON github_repositories(deleted_at);
CREATE INDEX IF NOT EXISTS idx_github_repositories_sync_status ON github_repositories(last_sync_status);

-- Create github_releases table
-- Synchronized GitHub releases
CREATE TABLE IF NOT EXISTS github_releases (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    repository_id INTEGER NOT NULL,
    release_id INTEGER NOT NULL,
    tag_name TEXT NOT NULL,
    release_name TEXT NOT NULL,
    body TEXT,
    is_prerelease INTEGER DEFAULT 0, -- SQLite boolean: 1=true, 0=false
    is_draft INTEGER DEFAULT 0,
    published_at DATETIME,
    download_url TEXT,
    file_size INTEGER DEFAULT 0,
    file_checksum TEXT,
    local_file_path TEXT,
    sync_status TEXT DEFAULT 'pending', -- Status: pending, processing, completed, failed
    version_id INTEGER, -- Link to versions table
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    deleted_at DATETIME,
    
    -- Create foreign key constraints
    FOREIGN KEY (repository_id) REFERENCES github_repositories(id) ON DELETE CASCADE,
    FOREIGN KEY (version_id) REFERENCES versions(id) ON DELETE SET NULL,
    UNIQUE (repository_id, release_id)
);

-- Create indexes for github_releases
CREATE INDEX IF NOT EXISTS idx_github_releases_repository_id ON github_releases(repository_id);
CREATE INDEX IF NOT EXISTS idx_github_releases_version_id ON github_releases(version_id);
CREATE INDEX IF NOT EXISTS idx_github_releases_sync_status ON github_releases(sync_status);
CREATE INDEX IF NOT EXISTS idx_github_releases_published_at ON github_releases(published_at);
CREATE INDEX IF NOT EXISTS idx_github_releases_deleted_at ON github_releases(deleted_at);

-- Create file_cache table
-- Cached files from GitHub releases and other sources
CREATE TABLE IF NOT EXISTS file_cache (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    app_id TEXT NOT NULL,
    version TEXT NOT NULL,
    original_url TEXT NOT NULL,
    local_path TEXT NOT NULL, -- Local file system path to cached file
    file_size INTEGER NOT NULL,
    file_checksum TEXT NOT NULL,
    content_type TEXT,
    downloaded_at DATETIME NOT NULL,
    last_accessed DATETIME NOT NULL,
    access_count INTEGER DEFAULT 0, -- Number of times this cached file has been accessed
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    
    -- Create foreign key constraints
    FOREIGN KEY (app_id) REFERENCES applications(app_id) ON DELETE CASCADE
);

-- Create indexes for file_cache
CREATE INDEX IF NOT EXISTS idx_file_cache_app_id ON file_cache(app_id);
CREATE INDEX IF NOT EXISTS idx_file_cache_version ON file_cache(version);
CREATE INDEX IF NOT EXISTS idx_file_cache_last_accessed ON file_cache(last_accessed);
CREATE INDEX IF NOT EXISTS idx_file_cache_checksum ON file_cache(file_checksum);
