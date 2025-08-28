package main

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"log"

	"github.com/Run-Panel/VerTree/internal/config"
	"github.com/Run-Panel/VerTree/internal/database"
	"github.com/Run-Panel/VerTree/internal/models"
)

// migrateToApps migrates existing data to the new application-based structure
func main() {
	log.Println("Starting migration to application-based structure...")

	// Load configuration
	cfg, err := config.LoadConfig()
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	// Initialize database
	if err := database.Initialize(cfg); err != nil {
		log.Fatalf("Failed to initialize database: %v", err)
	}

	// Run migration
	if err := migrateData(); err != nil {
		log.Fatalf("Migration failed: %v", err)
	}

	log.Println("Migration completed successfully!")
}

func migrateData() error {
	db := database.DB

	// Step 1: Check if default application already exists
	var defaultApp models.Application
	err := db.Where("app_id = ?", "app_default_legacy").First(&defaultApp).Error

	if err != nil {
		// Create default application
		log.Println("Creating default application...")

		// Find first admin to assign as creator
		var admin models.Admin
		if err := db.First(&admin).Error; err != nil {
			return fmt.Errorf("no admin found to assign as creator: %w", err)
		}

		defaultApp = models.Application{
			AppID:       "app_default_legacy",
			Name:        "Default Application",
			Description: "Default application for migrating existing versions and channels",
			IsActive:    true,
			CreatedBy:   admin.ID,
		}

		if err := db.Create(&defaultApp).Error; err != nil {
			return fmt.Errorf("failed to create default application: %w", err)
		}

		log.Printf("Created default application with ID: %s", defaultApp.AppID)
	} else {
		log.Printf("Default application already exists with ID: %s", defaultApp.AppID)
	}

	// Step 2: Migrate versions without app_id
	log.Println("Migrating versions to default application...")

	result := db.Model(&models.Version{}).
		Where("app_id IS NULL OR app_id = ''").
		Update("app_id", defaultApp.AppID)

	if result.Error != nil {
		return fmt.Errorf("failed to migrate versions: %w", result.Error)
	}

	log.Printf("Migrated %d versions to default application", result.RowsAffected)

	// Step 3: Migrate channels without app_id
	log.Println("Migrating channels to default application...")

	result = db.Model(&models.Channel{}).
		Where("app_id IS NULL OR app_id = ''").
		Update("app_id", defaultApp.AppID)

	if result.Error != nil {
		return fmt.Errorf("failed to migrate channels: %w", result.Error)
	}

	log.Printf("Migrated %d channels to default application", result.RowsAffected)

	// Step 4: Create default API key for the application
	log.Println("Creating default API key...")

	var existingKey models.ApplicationKey
	err = db.Where("app_id = ? AND name = ?", defaultApp.AppID, "Default Key").First(&existingKey).Error

	if err != nil {
		// Generate API key secret
		keySecret, err := models.GenerateKeySecret()
		if err != nil {
			return fmt.Errorf("failed to generate key secret: %w", err)
		}

		// Create hash
		hasher := sha256.New()
		hasher.Write([]byte(keySecret))
		keyHash := hex.EncodeToString(hasher.Sum(nil))

		defaultKey := models.ApplicationKey{
			AppID:       defaultApp.AppID,
			Name:        "Default Key",
			KeySecret:   keySecret,
			KeyHash:     keyHash,
			Permissions: models.PermissionsList([]string{"check_update", "download", "install"}),
			IsActive:    true,
			CreatedBy:   defaultApp.CreatedBy,
		}

		if err := db.Create(&defaultKey).Error; err != nil {
			return fmt.Errorf("failed to create default API key: %w", err)
		}

		log.Printf("Created default API key: %s", defaultKey.KeyID)
		log.Printf("API Key Secret (save this!): %s", keySecret)
		log.Printf("Authorization header: Bearer %s:%s", defaultApp.AppID, keySecret)
	} else {
		log.Printf("Default API key already exists: %s", existingKey.KeyID)
	}

	// Step 5: Display summary
	log.Println("\n=== Migration Summary ===")

	var versionCount, channelCount, keyCount int64
	db.Model(&models.Version{}).Where("app_id = ?", defaultApp.AppID).Count(&versionCount)
	db.Model(&models.Channel{}).Where("app_id = ?", defaultApp.AppID).Count(&channelCount)
	db.Model(&models.ApplicationKey{}).Where("app_id = ?", defaultApp.AppID).Count(&keyCount)

	log.Printf("Application: %s (%s)", defaultApp.Name, defaultApp.AppID)
	log.Printf("Versions: %d", versionCount)
	log.Printf("Channels: %d", channelCount)
	log.Printf("API Keys: %d", keyCount)
	log.Println("========================")

	return nil
}
