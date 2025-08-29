-- 005_fix_foreign_key_constraints.sql
-- Migration to fix incorrect foreign key constraints in applications table
-- The applications table should NOT reference versions or application_keys
-- Instead, versions and application_keys should reference applications

-- Disable foreign key checks temporarily
PRAGMA foreign_keys = OFF;

-- Step 1: Recreate applications table without the incorrect foreign key constraints
CREATE TABLE `applications_new` (
    `id` integer PRIMARY KEY,
    `app_id` text NOT NULL,
    `name` text NOT NULL,
    `description` text,
    `icon` text,
    `is_active` numeric DEFAULT true,
    `created_by` integer NOT NULL,
    `created_at` datetime,
    `updated_at` datetime,
    `deleted_at` datetime,
    CONSTRAINT `fk_applications_created_by_admin` FOREIGN KEY (`created_by`) REFERENCES `admins`(`id`)
);

-- Copy data from old table
INSERT INTO `applications_new` 
SELECT `id`, `app_id`, `name`, `description`, `icon`, `is_active`, `created_by`, `created_at`, `updated_at`, `deleted_at`
FROM `applications`;

-- Drop old table and rename new one
DROP TABLE `applications`;
ALTER TABLE `applications_new` RENAME TO `applications`;

-- Recreate indexes for applications table
CREATE INDEX `idx_applications_deleted_at` ON `applications`(`deleted_at`);
CREATE UNIQUE INDEX `idx_applications_name` ON `applications`(`name`);
CREATE UNIQUE INDEX `idx_applications_app_id` ON `applications`(`app_id`);

-- Step 2: Recreate versions table with correct foreign key
CREATE TABLE `versions_new` (
    `id` integer PRIMARY KEY,
    `app_id` text,
    `version` text NOT NULL,
    `channel` text NOT NULL DEFAULT "stable",
    `title` text NOT NULL,
    `description` text,
    `release_notes` text,
    `breaking_changes` text,
    `min_upgrade_version` text,
    `file_url` text NOT NULL,
    `file_size` integer NOT NULL,
    `file_checksum` text NOT NULL,
    `is_published` numeric DEFAULT false,
    `is_forced` numeric DEFAULT false,
    `publish_time` datetime,
    `created_at` datetime,
    `updated_at` datetime,
    `deleted_at` datetime,
    CONSTRAINT `fk_versions_application` FOREIGN KEY (`app_id`) REFERENCES `applications`(`app_id`)
);

-- Copy data from old table
INSERT INTO `versions_new` 
SELECT `id`, `app_id`, `version`, `channel`, `title`, `description`, `release_notes`, `breaking_changes`, 
       `min_upgrade_version`, `file_url`, `file_size`, `file_checksum`, `is_published`, `is_forced`, 
       `publish_time`, `created_at`, `updated_at`, `deleted_at`
FROM `versions`;

-- Drop old table and rename new one
DROP TABLE `versions`;
ALTER TABLE `versions_new` RENAME TO `versions`;

-- Recreate indexes for versions table
CREATE INDEX `idx_versions_deleted_at` ON `versions`(`deleted_at`);
CREATE UNIQUE INDEX `idx_app_version` ON `versions`(`app_id`,`version`);

-- Step 3: Recreate application_keys table with correct foreign key
CREATE TABLE `application_keys_new` (
    `id` integer PRIMARY KEY,
    `key_id` text NOT NULL,
    `app_id` text NOT NULL,
    `name` text NOT NULL,
    `key_secret` text NOT NULL,
    `key_hash` text NOT NULL,
    `permissions` json,
    `is_active` numeric DEFAULT true,
    `last_used` datetime,
    `created_by` integer NOT NULL,
    `created_at` datetime,
    `updated_at` datetime,
    `deleted_at` datetime,
    CONSTRAINT `fk_application_keys_created_by_admin` FOREIGN KEY (`created_by`) REFERENCES `admins`(`id`),
    CONSTRAINT `fk_application_keys_application` FOREIGN KEY (`app_id`) REFERENCES `applications`(`app_id`)
);

-- Copy data from old table
INSERT INTO `application_keys_new` 
SELECT `id`, `key_id`, `app_id`, `name`, `key_secret`, `key_hash`, `permissions`, `is_active`, 
       `last_used`, `created_by`, `created_at`, `updated_at`, `deleted_at`
FROM `application_keys`;

-- Drop old table and rename new one
DROP TABLE `application_keys`;
ALTER TABLE `application_keys_new` RENAME TO `application_keys`;

-- Recreate indexes for application_keys table
CREATE INDEX `idx_application_keys_app_id` ON `application_keys`(`app_id`);
CREATE UNIQUE INDEX `idx_application_keys_key_id` ON `application_keys`(`key_id`);
CREATE INDEX `idx_application_keys_deleted_at` ON `application_keys`(`deleted_at`);
CREATE INDEX `idx_application_keys_key_hash` ON `application_keys`(`key_hash`);

-- Re-enable foreign key checks
PRAGMA foreign_keys = ON;
