package main

import (
	"fmt"
	"log"

	"forest-management/config"
	"forest-management/database"
	"forest-management/internal/models"
	"forest-management/pkg/utils"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

func main() {
	// Load config
	config.InitConfig()
	cfg := config.AppConfig

	// Connect to DB
	dsn := fmt.Sprintf(
		"host=%s port=%s user=%s password=%s dbname=%s sslmode=%s",
		cfg.DBHost, cfg.DBPort, cfg.DBUser, cfg.DBPassword, cfg.DBName, cfg.DBSSLMode,
	)

	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Info),
	})
	if err != nil {
		log.Fatalf("❌ DB connection failed: %v", err)
	}
	database.DB = db

	// Hash the admin password
	hashedPass, err := utils.HashPassword("admin123")
	if err != nil {
		log.Fatalf("❌ Failed to hash password: %v", err)
	}

	// Create admin user
	admin := models.User{
		Name:     "System Admin",
		Phone:    "9800000000",
		Password: hashedPass,
		Role:     "admin",
		Status:   "active",
	}

	result := db.Where("phone = ?", "9800000000").First(&models.User{})
	if result.Error != nil {
		// User doesn't exist — create it
		if err := db.Create(&admin).Error; err != nil {
			log.Fatalf("❌ Failed to create admin: %v", err)
		}
		fmt.Println("✅ Default admin user created!")
	} else {
		// User exists — update password
		db.Model(&models.User{}).Where("phone = ?", "9800000000").Update("password", hashedPass)
		fmt.Println("✅ Admin password reset!")
	}

	fmt.Println("")
	fmt.Println("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━")
	fmt.Println("  ADMIN CREDENTIALS:")
	fmt.Println("  Phone:    9800000000")
	fmt.Println("  Password: admin123")
	fmt.Println("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━")
	fmt.Println("")
}
