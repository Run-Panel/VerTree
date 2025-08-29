-- 004_fix_application_channels_table.sql
-- Migration to fix application_channels table structure
-- GORM's many2many doesn't handle custom fields properly, so we need manual migration

-- Drop the existing simple table
DROP TABLE IF EXISTS `application_channels`;

-- Recreate with proper structure matching ApplicationChannel model
CREATE TABLE `application_channels` (
    `id` integer PRIMARY KEY,
    `app_id` text NOT NULL,
    `channel_name` text NOT NULL,
    `is_enabled` boolean DEFAULT true,
    `auto_publish` boolean DEFAULT false,
    `rollout_percentage` integer DEFAULT 100,
    `created_at` datetime,
    `updated_at` datetime,
    `deleted_at` datetime,
    CONSTRAINT `fk_application_channels_application` FOREIGN KEY (`app_id`) REFERENCES `applications`(`app_id`),
    CONSTRAINT `fk_application_channels_channel` FOREIGN KEY (`channel_name`) REFERENCES `channels`(`name`)
);

-- Create indexes
CREATE UNIQUE INDEX `idx_app_channel` ON `application_channels`(`app_id`,`channel_name`);
CREATE INDEX `idx_application_channels_deleted_at` ON `application_channels`(`deleted_at`);

-- Insert default channel associations for existing applications
-- This ensures all existing apps have access to the default channels
INSERT INTO `application_channels` (`app_id`, `channel_name`, `is_enabled`, `auto_publish`, `rollout_percentage`, `created_at`, `updated_at`)
SELECT 
    a.app_id,
    c.name,
    CASE WHEN c.name = 'stable' THEN true ELSE false END as is_enabled,
    false as auto_publish,
    100 as rollout_percentage,
    datetime('now') as created_at,
    datetime('now') as updated_at
FROM applications a
CROSS JOIN channels c
WHERE a.deleted_at IS NULL AND c.deleted_at IS NULL
AND NOT EXISTS (
    SELECT 1 FROM application_channels ac 
    WHERE ac.app_id = a.app_id AND ac.channel_name = c.name
);

