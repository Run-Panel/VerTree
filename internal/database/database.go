package database

import (
	"fmt"
	"log"
	"os"
	"time"

	"github.com/Run-Panel/VerTree/internal/config"
	"github.com/Run-Panel/VerTree/internal/models"
	"github.com/glebarez/sqlite"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

// DB is the global database instance
var DB *gorm.DB

// Initialize initializes the database connection
func Initialize(cfg *config.Config) error {
	var err error
	var dialector gorm.Dialector

	// Configure GORM logger
	logLevel := logger.Info
	if cfg.App.Environment == "production" {
		logLevel = logger.Error
	}

	gormConfig := &gorm.Config{
		Logger: logger.Default.LogMode(logLevel),
		NowFunc: func() time.Time {
			return time.Now().UTC()
		},
	}

	// Choose database driver based on configuration
	switch cfg.Database.Driver {
	case "postgres":
		dsn := fmt.Sprintf("host=%s user=%s password=%s dbname=%s port=%s sslmode=%s TimeZone=UTC",
			cfg.Database.Host,
			cfg.Database.User,
			cfg.Database.Password,
			cfg.Database.Name,
			cfg.Database.Port,
			cfg.Database.SSLMode,
		)
		dialector = postgres.Open(dsn)
	case "sqlite":
		// For development/testing
		// Ensure data directory exists
		os.MkdirAll("./data", 0755)
		dsn := fmt.Sprintf("./data/%s.db", cfg.Database.Name)
		dialector = sqlite.Open(dsn)
	default:
		return fmt.Errorf("unsupported database driver: %s", cfg.Database.Driver)
	}

	// Connect to database
	DB, err = gorm.Open(dialector, gormConfig)
	if err != nil {
		return fmt.Errorf("failed to connect to database: %w", err)
	}

	// Get underlying sql.DB for additional configuration
	sqlDB, err := DB.DB()
	if err != nil {
		return fmt.Errorf("failed to get sql.DB instance: %w", err)
	}

	// Set connection pool settings
	sqlDB.SetMaxIdleConns(10)
	sqlDB.SetMaxOpenConns(100)
	sqlDB.SetConnMaxLifetime(time.Hour)

	// Test connection
	if err := sqlDB.Ping(); err != nil {
		return fmt.Errorf("failed to ping database: %w", err)
	}

	log.Printf("Successfully connected to %s database", cfg.Database.Driver)
	return nil
}

// AutoMigrate runs database migrations
func AutoMigrate() error {
	err := DB.AutoMigrate(
		&models.Version{},
		&models.Channel{},
		&models.UpdateRule{},
		&models.UpdateStat{},
		&models.Admin{},
		&models.RefreshToken{},
	)
	if err != nil {
		return fmt.Errorf("failed to run auto migration: %w", err)
	}

	log.Println("Database migration completed successfully")
	return nil
}

// SeedDefaultData seeds the database with default data
func SeedDefaultData() error {
	// Check if channels already exist
	var count int64
	if err := DB.Model(&models.Channel{}).Count(&count).Error; err != nil {
		return fmt.Errorf("failed to count channels: %w", err)
	}

	// If channels don't exist, create default ones
	if count == 0 {
		defaultChannels := []models.Channel{
			{
				Name:              "stable",
				DisplayName:       "稳定版",
				Description:       "经过充分测试的稳定版本",
				IsActive:          true,
				AutoPublish:       false,
				RolloutPercentage: 100,
			},
			{
				Name:              "beta",
				DisplayName:       "测试版",
				Description:       "功能完整的测试版本",
				IsActive:          true,
				AutoPublish:       false,
				RolloutPercentage: 100,
			},
			{
				Name:              "alpha",
				DisplayName:       "预览版",
				Description:       "最新功能预览版本",
				IsActive:          true,
				AutoPublish:       false,
				RolloutPercentage: 50,
			},
		}

		for _, channel := range defaultChannels {
			if err := DB.Create(&channel).Error; err != nil {
				return fmt.Errorf("failed to create channel %s: %w", channel.Name, err)
			}
		}

		log.Println("Default channels created successfully")
	}

	// Check if admins already exist
	var adminCount int64
	if err := DB.Model(&models.Admin{}).Count(&adminCount).Error; err != nil {
		return fmt.Errorf("failed to count admins: %w", err)
	}

	// If admins don't exist, create default admin
	if adminCount == 0 {
		log.Println("ℹ️  No admin users found. Please run 'go run scripts/create-admin.go' to create the default admin user.")
	}

	return nil
}

// GetDB returns the database instance
func GetDB() *gorm.DB {
	return DB
}

// Close closes the database connection
func Close() error {
	if DB != nil {
		sqlDB, err := DB.DB()
		if err != nil {
			return fmt.Errorf("failed to get sql.DB instance: %w", err)
		}
		return sqlDB.Close()
	}
	return nil
}
