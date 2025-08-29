-- 002_fix_application_keys.sql
-- Migration to fix ApplicationKey data inconsistencies
-- This migration only handles data fixes, GORM handles table structure

-- Update any NULL or empty permissions to default JSON array
-- This ensures backward compatibility with existing keys
UPDATE application_keys 
SET permissions = '["check_update", "download", "install"]' 
WHERE permissions IS NULL OR permissions = '' OR permissions = 'null';
