package main

import (
	"log"

	"forest-management/config"
	"forest-management/database"
)

func main() {
	config.InitConfig()
	database.Connect()
	if err := database.Migrate(database.DB); err != nil {
		log.Fatalf("migration failed: %v", err)
	}
	if err := database.RunPostMigrationMaintenance(database.DB); err != nil {
		log.Fatalf("post-migration maintenance failed: %v", err)
	}
}
