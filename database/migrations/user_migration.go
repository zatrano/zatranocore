package migrations

import (
	"log"

	"zatrano/models"

	"gorm.io/gorm"
)

func MigrateUsersTable(db *gorm.DB) error {
	err := db.Exec(`DO $$
	BEGIN
		IF NOT EXISTS (SELECT 1 FROM pg_type WHERE typname = 'user_type') THEN
			CREATE TYPE user_type AS ENUM ('system', 'manager', 'agent');
			RAISE NOTICE '✅ user_type ENUM created.';
		ELSE
			RAISE NOTICE '✅ user_type ENUM already exists.';
		END IF;
	END$$`).Error
	if err != nil {
		log.Printf("Failed to create/check user_type enum: %v", err)
		return err
	}

	err = db.AutoMigrate(&models.User{})
	if err != nil {
		log.Printf("Failed to migrate users table: %v", err)
		return err
	}
	log.Println("✅ Users table structure migrated successfully with GORM")

	if !db.Migrator().HasConstraint(&models.User{}, "fk_users_team") {
		err = db.Exec(`
			ALTER TABLE users
			ADD CONSTRAINT fk_users_team
			FOREIGN KEY (team_id) REFERENCES teams(id)
			ON UPDATE CASCADE ON DELETE SET NULL
		`).Error
		if err != nil && !db.Migrator().HasConstraint(&models.User{}, "fk_users_team") {
			log.Printf("Failed to add team foreign key constraint: %v", err)

		} else if err == nil {
			log.Println("✅ Added team foreign key constraint manually")
		} else {
			log.Println("✅ Team foreign key constraint likely already exists (managed by GORM or previous run)")
		}
	} else {
		log.Println("✅ Team foreign key constraint already exists (managed by GORM or previous run)")
	}

	return nil
}
