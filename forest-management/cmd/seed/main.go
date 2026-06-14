package main

import (
	"fmt"
	"log"
	"strconv"

	"forest-management/config"
	"forest-management/database"
	"forest-management/internal/models"
	"forest-management/pkg/security"
	"forest-management/pkg/utils"
)

func main() {
	// Load config
	config.InitConfig()
	cfg := config.AppConfig
	normalizedAdminPhone, err := security.NormalizeNepalMobile(cfg.SeedAdminPhone)
	if err != nil {
		log.Fatalf("invalid SEED_ADMIN_PHONE: %v", err)
	}
	cfg.SeedAdminPhone = normalizedAdminPhone

	// Connect and run the same migrations used by the API. This makes the seed
	// command safe to run against a newly created, empty database.
	database.InitDB()
	db := database.DB

	if cfg.SeedAdminPhone == "" || cfg.SeedAdminPassword == "" {
		log.Fatal("❌ SEED_ADMIN_PHONE and SEED_ADMIN_PASSWORD must be set in the environment")
	}
	if err := security.ValidatePassword(cfg.SeedAdminPassword); err != nil {
		log.Fatalf("❌ SEED_ADMIN_PASSWORD is not strong enough: %v", err)
	}

	// The seeded administrator is a one-time bootstrap account. Once a real
	// committee administrator exists, re-running the seed command must not
	// reactivate or overwrite that bootstrap account.
	var realAdminCount int64
	db.Model(&models.User{}).
		Where("role = ? AND status = ? AND phone <> ? AND is_bootstrap_admin = ?", "admin", "active", cfg.SeedAdminPhone, false).
		Count(&realAdminCount)

	hashedPass, err := utils.HashPassword(cfg.SeedAdminPassword)
	if err != nil {
		log.Fatalf("❌ Failed to hash password: %v", err)
	}

	var existing models.User
	result := db.Where("phone = ?", cfg.SeedAdminPhone).First(&existing)
	if result.Error != nil {
		if realAdminCount > 0 {
			fmt.Println("ℹ️ A real administrator already exists; bootstrap admin was not created")
		} else {
			admin := models.User{
				Name:               cfg.SeedAdminName,
				Phone:              cfg.SeedAdminPhone,
				Password:           hashedPass,
				Role:               "admin",
				Status:             "active",
				IsBootstrapAdmin:   true,
				MustChangePassword: true,
			}
			if err := db.Create(&admin).Error; err != nil {
				log.Fatalf("❌ Failed to create bootstrap admin: %v", err)
			}
			fmt.Println("✅ Initial bootstrap admin user created")
		}
	} else {
		updates := map[string]interface{}{
			"name":                 cfg.SeedAdminName,
			"role":                 "admin",
			"is_bootstrap_admin":   true,
			"must_change_password": true,
		}
		if realAdminCount > 0 {
			updates["status"] = "inactive"
			fmt.Println("✅ Bootstrap admin remains inactive because a committee admin exists")
		} else {
			updates["password"] = hashedPass
			updates["status"] = "active"
			fmt.Println("✅ Bootstrap admin account updated")
		}
		if err := db.Model(&existing).Updates(updates).Error; err != nil {
			log.Fatalf("❌ Failed to update bootstrap admin: %v", err)
		}
	}

	fmt.Printf("Admin phone: %s\n", cfg.SeedAdminPhone)

	if cfg.SeedSamitiName != "" {
		wardNo, err := strconv.Atoi(cfg.SeedSamitiWardNo)
		if err != nil || wardNo <= 0 {
			wardNo = 1
		}

		registrationNo := cfg.SeedSamitiRegistrationNo
		var registrationNoPtr *string
		if registrationNo != "" {
			registrationNoPtr = &registrationNo
		}

		settings := models.SamitiSetting{
			Name:           cfg.SeedSamitiName,
			RegistrationNo: registrationNoPtr,
			Address:        cfg.SeedSamitiAddress,
			WardNo:         wardNo,
			Municipality:   cfg.SeedSamitiMunicipality,
			District:       cfg.SeedSamitiDistrict,
			Province:       cfg.SeedSamitiProvince,
		}

		var existingSettings models.SamitiSetting
		if err := db.First(&existingSettings).Error; err != nil {
			if err := db.Create(&settings).Error; err != nil {
				log.Fatalf("❌ Failed to create organization settings: %v", err)
			}
			fmt.Println("✅ Organization settings created")
		} else {
			updates := map[string]interface{}{
				"name":            settings.Name,
				"registration_no": settings.RegistrationNo,
				"address":         settings.Address,
				"ward_no":         settings.WardNo,
				"municipality":    settings.Municipality,
				"district":        settings.District,
				"province":        settings.Province,
			}
			if err := db.Model(&existingSettings).Updates(updates).Error; err != nil {
				log.Fatalf("❌ Failed to update organization settings: %v", err)
			}
			fmt.Println("✅ Organization settings updated")
		}
	}
}
