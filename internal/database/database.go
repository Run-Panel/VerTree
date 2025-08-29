package database

import (
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"github.com/Run-Panel/VerTree/internal/config"
	"github.com/Run-Panel/VerTree/internal/models"
	"github.com/Run-Panel/VerTree/internal/utils"
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

	// Run database migrations
	if err := AutoMigrate(); err != nil {
		return fmt.Errorf("failed to run auto migration: %w", err)
	}

	// Run SQL migrations
	if err := RunSQLMigrations(); err != nil {
		return fmt.Errorf("failed to run SQL migrations: %w", err)
	}

	log.Printf("Successfully connected to %s database", cfg.Database.Driver)
	return nil
}

// AutoMigrate runs database migrations
func AutoMigrate() error {
	err := DB.AutoMigrate(
		&models.Application{},
		&models.ApplicationKey{},
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
				Name:        "stable",
				DisplayName: "稳定版",
				Description: "经过充分测试的稳定版本",
				IsActive:    true,
				SortOrder:   1,
			},
			{
				Name:        "beta",
				DisplayName: "测试版",
				Description: "功能完整的测试版本",
				IsActive:    true,
				SortOrder:   2,
			},
			{
				Name:        "alpha",
				DisplayName: "预览版",
				Description: "最新功能预览版本",
				IsActive:    true,
				SortOrder:   3,
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
		log.Println("🔐 No admin users found. Creating default admin user...")
		if err := createDefaultAdmin(); err != nil {
			return fmt.Errorf("failed to create default admin: %w", err)
		}
		log.Println("✅ Default admin user created successfully")
	}

	return nil
}

// createDefaultAdmin creates a default admin user
func createDefaultAdmin() error {
	// Import utils package for password hashing
	hashedPassword, err := utils.HashPassword("admin123")
	if err != nil {
		return fmt.Errorf("failed to hash password: %w", err)
	}

	admin := models.Admin{
		Username: "admin",
		Email:    "admin@runpanel.dev",
		Password: hashedPassword,
		Role:     "superadmin",
		IsActive: true,
	}

	if err := DB.Create(&admin).Error; err != nil {
		return fmt.Errorf("failed to create admin: %w", err)
	}

	log.Printf("🔐 Default admin created:")
	log.Printf("   Username: %s", admin.Username)
	log.Printf("   Email: %s", admin.Email)
	log.Printf("   Password: admin123")
	log.Printf("   Role: %s", admin.Role)
	log.Printf("   ⚠️  PLEASE CHANGE THE DEFAULT PASSWORD AFTER FIRST LOGIN!")

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

// Migration represents a database migration
type Migration struct {
	ID        uint   `gorm:"primaryKey"`
	Filename  string `gorm:"uniqueIndex;not null"`
	AppliedAt time.Time
}

// RunSQLMigrations runs SQL migration files from the migrations directory
func RunSQLMigrations() error {
	// Ensure migration tracking table exists
	if err := DB.AutoMigrate(&Migration{}); err != nil {
		return fmt.Errorf("failed to create migrations table: %w", err)
	}

	// Read migration files
	migrationsDir := "./migrations"
	if _, err := os.Stat(migrationsDir); os.IsNotExist(err) {
		log.Println("No migrations directory found, skipping SQL migrations")
		return nil
	}

	files, err := ioutil.ReadDir(migrationsDir)
	if err != nil {
		return fmt.Errorf("failed to read migrations directory: %w", err)
	}

	// Filter and sort SQL files
	var sqlFiles []string
	for _, file := range files {
		if !file.IsDir() && strings.HasSuffix(file.Name(), ".sql") {
			sqlFiles = append(sqlFiles, file.Name())
		}
	}
	sort.Strings(sqlFiles)

	// Apply migrations
	for _, filename := range sqlFiles {
		// Check if migration was already applied
		var migration Migration
		result := DB.Where("filename = ?", filename).First(&migration)
		if result.Error == nil {
			// Migration already applied
			continue
		}

		// Read migration file
		filePath := filepath.Join(migrationsDir, filename)
		content, err := ioutil.ReadFile(filePath)
		if err != nil {
			return fmt.Errorf("failed to read migration file %s: %w", filename, err)
		}

		// Execute migration
		log.Printf("Applying migration: %s", filename)
		if err := DB.Exec(string(content)).Error; err != nil {
			return fmt.Errorf("failed to apply migration %s: %w", filename, err)
		}

		// Record migration as applied
		migration = Migration{
			Filename:  filename,
			AppliedAt: time.Now(),
		}
		if err := DB.Create(&migration).Error; err != nil {
			return fmt.Errorf("failed to record migration %s: %w", filename, err)
		}

		log.Printf("Successfully applied migration: %s", filename)
	}

	return nil
}
