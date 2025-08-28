-- Migration script to add application support and migrate existing data
-- This script creates a default application and assigns existing versions and channels to it

-- Step 1: Create a default application (if not exists)
INSERT OR IGNORE INTO applications (
    app_id, 
    name, 
    description, 
    is_active, 
    created_by, 
    created_at, 
    updated_at
) VALUES (
    'app_default_legacy',
    'Default Application', 
    'Default application for migrating existing versions and channels',
    1,
    1,
    datetime('now'),
    datetime('now')
);

-- Step 2: Update existing versions to use the default application
UPDATE versions 
SET app_id = 'app_default_legacy' 
WHERE app_id IS NULL OR app_id = '';

-- Step 3: Update existing channels to use the default application  
UPDATE channels 
SET app_id = 'app_default_legacy'
WHERE app_id IS NULL OR app_id = '';

-- Step 4: Update existing update_stats to reference the default app (if needed)
-- This is optional since update_stats doesn't have app_id field yet
-- UPDATE update_stats SET app_id = 'app_default_legacy' WHERE app_id IS NULL;
