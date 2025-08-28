-- Migration to fix ApplicationKey data
-- This migration only handles data fixes, letting GORM handle table structure

-- Update any NULL or empty permissions to default JSON array
UPDATE application_keys 
SET permissions = '["check_update", "download", "install"]' 
WHERE permissions IS NULL OR permissions = '' OR permissions = 'null';
