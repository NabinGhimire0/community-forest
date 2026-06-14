package database

import (
	"fmt"
	"log"
	"time"

	"forest-management/config"
	"forest-management/internal/models"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

var DB *gorm.DB

// InitDB connects to PostgreSQL and optionally runs migrations. Production
// deployments should keep RUN_AUTO_MIGRATE=false and execute cmd/migrate as a
// controlled deployment step.
func InitDB() {
	Connect()
	if config.AppConfig.RunAutoMigrate {
		if err := Migrate(DB); err != nil {
			log.Fatalf("❌ Migration failed: %v", err)
		}
		if err := RunPostMigrationMaintenance(DB); err != nil {
			log.Fatalf("❌ Post-migration maintenance failed: %v", err)
		}
	} else if err := CleanupExpiredSessions(DB); err != nil {
		log.Printf("⚠️ Session cleanup warning: %v", err)
	}
}

func Connect() {
	cfg := config.AppConfig
	dsn := fmt.Sprintf(
		"host=%s port=%s user=%s password=%s dbname=%s sslmode=%s application_name=bansamiti",
		cfg.DBHost, cfg.DBPort, cfg.DBUser, cfg.DBPassword, cfg.DBName, cfg.DBSSLMode,
	)

	gormConfig := &gorm.Config{PrepareStmt: true}
	if cfg.AppEnv == "development" {
		gormConfig.Logger = logger.Default.LogMode(logger.Info)
	} else {
		gormConfig.Logger = logger.Default.LogMode(logger.Error)
	}

	var err error
	DB, err = gorm.Open(postgres.Open(dsn), gormConfig)
	if err != nil {
		log.Fatalf("❌ Failed to connect to database: %v", err)
	}

	sqlDB, err := DB.DB()
	if err != nil {
		log.Fatalf("❌ Failed to configure database pool: %v", err)
	}
	sqlDB.SetMaxOpenConns(cfg.DBMaxOpenConns)
	sqlDB.SetMaxIdleConns(cfg.DBMaxIdleConns)
	sqlDB.SetConnMaxLifetime(time.Duration(cfg.DBConnMaxMinutes) * time.Minute)
	if err := sqlDB.Ping(); err != nil {
		log.Fatalf("❌ Database ping failed: %v", err)
	}
	fmt.Println("✅ Connected to database")
}

func Migrate(db *gorm.DB) error {
	if err := db.AutoMigrate(
		&models.SamitiSetting{},
		&models.User{},
		&models.UserSession{},
		&models.SamitiHead{},
		&models.Member{},
		&models.FamilyMember{},
		&models.FiscalYear{},
		&models.FeeSetting{},
		&models.ResourceType{},
		&models.ResourceItem{},
		&models.ResourceRate{},
		&models.Stock{},
		&models.Request{},
		&models.Payment{},
		&models.Transaction{},
		&models.ExpenseCategory{},
		&models.Expense{},
		&models.Fine{},
		&models.Letter{},
		&models.AuditLog{},
		&models.Notification{},
		&models.NotificationReceipt{},
		&models.FileUpload{},
	); err != nil {
		return err
	}
	fmt.Println("✅ Database migration completed")
	return nil
}

func RunPostMigrationMaintenance(db *gorm.DB) error {
	cfg := config.AppConfig
	if cfg.SeedAdminPhone != "" {
		if err := db.Model(&models.User{}).
			Where("phone = ? AND role = ?", cfg.SeedAdminPhone, "admin").
			Update("is_bootstrap_admin", true).Error; err != nil {
			return fmt.Errorf("mark bootstrap administrator: %w", err)
		}
	}

	statements := []string{
		`CREATE UNIQUE INDEX IF NOT EXISTS idx_single_active_fiscal_year
		 ON fiscal_years ((is_active)) WHERE is_active = true`,
		`CREATE UNIQUE INDEX IF NOT EXISTS idx_unique_member_fiscal_membership_fee
		 ON transactions (member_id, fiscal_year_id)
		 WHERE type = 'membership_fee' AND record_status <> 'reversed'`,
		`CREATE INDEX IF NOT EXISTS idx_sessions_active_lookup
		 ON user_sessions (token_hash, expires_at)
		 WHERE revoked_at IS NULL`,
		`CREATE INDEX IF NOT EXISTS idx_audit_created_action
		 ON audit_logs (created_at DESC, action)`,
		`UPDATE stocks AS s
		 SET reserved_quantity = approved.total_reserved
		 FROM (
			 SELECT resource_item_id, fiscal_year_id, COALESCE(SUM(quantity_approved), 0) AS total_reserved
			 FROM requests
			 WHERE status = 'approved' AND quantity_approved IS NOT NULL
			 GROUP BY resource_item_id, fiscal_year_id
		 ) AS approved
		 WHERE s.resource_item_id = approved.resource_item_id
		   AND s.fiscal_year_id = approved.fiscal_year_id
		   AND s.remaining_quantity >= approved.total_reserved
		   AND s.reserved_quantity IS DISTINCT FROM approved.total_reserved`,
		`UPDATE stocks AS s
		 SET reserved_quantity = 0
		 WHERE reserved_quantity <> 0
		   AND NOT EXISTS (
			 SELECT 1 FROM requests r
			 WHERE r.status = 'approved'
			   AND r.quantity_approved IS NOT NULL
			   AND r.resource_item_id = s.resource_item_id
			   AND r.fiscal_year_id = s.fiscal_year_id
		   )`,
	}
	for _, statement := range statements {
		if err := db.Exec(statement).Error; err != nil {
			return err
		}
	}
	return nil
}

func CleanupExpiredSessions(db *gorm.DB) error {
	return db.Where("expires_at < ?", time.Now().UTC().Add(-7*24*time.Hour)).Delete(&models.UserSession{}).Error
}
