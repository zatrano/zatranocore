package migrations

import (
	"log"

	"zatrano/models"

	"gorm.io/gorm"
)

func MigrateTeamsTable(db *gorm.DB) error {
	err := db.AutoMigrate(&models.Team{})
	if err != nil {
		log.Printf("Failed to migrate teams table: %v", err)
		return err
	}

	log.Println("✅ Teams table migrated successfully with GORM")

	return nil
}
