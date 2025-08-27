package main

import (
	"fmt"
	"log"

	"github.com/Run-Panel/VerTree/internal/config"
	"github.com/Run-Panel/VerTree/internal/database"
	"github.com/Run-Panel/VerTree/internal/models"
	"github.com/Run-Panel/VerTree/internal/utils"
)

func main() {
	// Load configuration
	cfg := config.Load()

	// Initialize database
	if err := database.Initialize(cfg); err != nil {
		log.Fatalf("Failed to initialize database: %v", err)
	}
	defer database.Close()

	// Run database migrations to ensure tables exist
	if err := database.AutoMigrate(); err != nil {
		log.Fatalf("Failed to run database migrations: %v", err)
	}

	// Create default admin
	password := "admin123"
	hashedPassword, err := utils.HashPassword(password)
	if err != nil {
		log.Fatalf("Failed to hash password: %v", err)
	}

	admin := models.Admin{
		Username: "admin",
		Email:    "admin@runpanel.dev",
		Password: hashedPassword,
		Role:     "superadmin",
		IsActive: true,
	}

	db := database.GetDB()

	// Check if admin already exists
	var count int64
	db.Model(&models.Admin{}).Where("username = ?", admin.Username).Count(&count)

	if count > 0 {
		fmt.Printf("✅ Admin user '%s' already exists\n", admin.Username)
		return
	}

	// Create admin
	if err := db.Create(&admin).Error; err != nil {
		log.Fatalf("Failed to create admin: %v", err)
	}

	fmt.Printf("🔐 Default admin created successfully:\n")
	fmt.Printf("   Username: %s\n", admin.Username)
	fmt.Printf("   Email: %s\n", admin.Email)
	fmt.Printf("   Password: %s\n", password)
	fmt.Printf("   Role: %s\n", admin.Role)
	fmt.Printf("   ⚠️  PLEASE CHANGE THE DEFAULT PASSWORD AFTER FIRST LOGIN!\n")
}
