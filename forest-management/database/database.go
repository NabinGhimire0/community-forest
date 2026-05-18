package database

import (
	"fmt"
	"log"

	"forest-management/config"
	"forest-management/internal/models"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

var DB *gorm.DB

func InitDB() {
	cfg := config.AppConfig

	dsn := fmt.Sprintf(
		"host=%s port=%s user=%s password=%s dbname=%s sslmode=%s",
		cfg.DBHost, cfg.DBPort, cfg.DBUser, cfg.DBPassword, cfg.DBName, cfg.DBSSLMode,
	)

	var err error

	// In development, use detailed logging; in production, silent
	gormConfig := &gorm.Config{}
	if cfg.AppEnv == "development" {
		gormConfig.Logger = logger.Default.LogMode(logger.Info)
	}

	DB, err = gorm.Open(postgres.Open(dsn), gormConfig)
	if err != nil {
		log.Fatalf("❌ Failed to connect to database: %v", err)
	}

	fmt.Println("✅ Connected to database")

	// Auto-migrate all models — creates/updates tables
	err = DB.AutoMigrate(
		// Organization
		&models.SamitiSetting{},
		&models.SamitiHead{},

		// Auth & Users
		&models.User{},
		&models.Member{},
		&models.FamilyMember{},

		// Fiscal & Finance
		&models.FiscalYear{},
		&models.FeeSetting{},

		// Resources
		&models.ResourceType{},
		&models.ResourceItem{},
		&models.ResourceRate{},
		&models.Stock{},

		// Workflow
		&models.Request{},
		&models.Payment{},
		&models.Transaction{},

		// Expenses
		&models.ExpenseCategory{},
		&models.Expense{},

		// Fines
		&models.Fine{},

		// Letters
		&models.Letter{},
		&models.AuditLog{},
		&models.Notification{},
		&models.FileUpload{},
	)
	if err != nil {
		log.Fatalf("❌ Migration failed: %v", err)
	}

	fmt.Println("✅ Database migration completed")
}
