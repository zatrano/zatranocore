package seeders

import (
	"errors"
	"log"

	"zatrano/models"

	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

func GetSystemUserConfig() models.User {
	return models.User{
		Name:     "System",
		Account:  "system@system",
		Type:     models.System,
		Password: "S1st3m@S1st3m",
	}
}

func SeedSystemUser(db *gorm.DB) error {
	systemUserConfig := GetSystemUserConfig()

	hashedPasswordBytes, err := bcrypt.GenerateFromPassword([]byte(systemUserConfig.Password), bcrypt.DefaultCost)
	if err != nil {
		log.Printf("ERROR: Failed to hash password for system user '%s': %v", systemUserConfig.Account, err)
		return err
	}
	hashedPassword := string(hashedPasswordBytes)
	log.Printf("Password for system user '%s' hashed successfully in seeder.", systemUserConfig.Account)

	userToSeed := models.User{
		Name:     systemUserConfig.Name,
		Account:  systemUserConfig.Account,
		Type:     systemUserConfig.Type,
		Password: hashedPassword,
		Status:   true,
	}

	var existingUser models.User
	result := db.Where("account = ? AND type = ?", userToSeed.Account, userToSeed.Type).First(&existingUser)

	if result.Error == nil {
		log.Printf("System user '%s' already exists. Checking if update is needed...", userToSeed.Account)

		pwMatchErr := bcrypt.CompareHashAndPassword([]byte(existingUser.Password), []byte(systemUserConfig.Password))
		needsUpdate := false
		if pwMatchErr != nil {
			log.Printf("Password for system user '%s' needs update.", userToSeed.Account)
			existingUser.Password = hashedPassword
			needsUpdate = true
		}
		if existingUser.Name != userToSeed.Name {
			log.Printf("Name for system user '%s' needs update.", userToSeed.Account)
			existingUser.Name = userToSeed.Name
			needsUpdate = true
		}

		if needsUpdate {
			log.Printf("Updating existing system user '%s'...", userToSeed.Account)
			err = db.Save(&existingUser).Error
			if err != nil {
				log.Printf("ERROR: Failed to update existing system user '%s': %v", userToSeed.Account, err)
				return err
			}
			log.Printf("Existing system user '%s' updated successfully.", userToSeed.Account)
		} else {
			log.Printf("No update needed for existing system user '%s'.", userToSeed.Account)
		}
		return nil

	} else if !errors.Is(result.Error, gorm.ErrRecordNotFound) {
		log.Printf("ERROR: Database error checking for system user '%s': %v", userToSeed.Account, result.Error)
		return result.Error
	}

	log.Printf("System user '%s' not found. Creating...", userToSeed.Account)
	err = db.Create(&userToSeed).Error
	if err != nil {
		log.Printf("ERROR: Failed to create system user '%s': %v", userToSeed.Account, err)
		return err
	}

	log.Printf("System user '%s' created successfully via seeder.", userToSeed.Account)
	return nil
}
