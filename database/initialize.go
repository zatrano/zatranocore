// database/initialize.go
package database

import (
	"errors"
	"fmt" // fmt.Errorf için eklendi
	"log"
	"time"

	// Migrations paketini import et
	"zatrano/database/migrations"
	"zatrano/database/seeders"
	"zatrano/models"

	"gorm.io/gorm"
)

func Initialize(db *gorm.DB) {
	sqlDB, err := db.DB()
	if err != nil {
		log.Fatalf("Failed to get database instance: %v", err)
	}
	sqlDB.SetMaxIdleConns(10)
	sqlDB.SetMaxOpenConns(100)
	sqlDB.SetConnMaxLifetime(time.Hour)

	tx := db.Begin()
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
			log.Fatal("Database initialization failed, rolled back:", r)
		}
		// Eğer tx commit edilmemişse (hata nedeniyle), rollback yapıldığından emin ol
		// Commit başarılıysa tx.Error nil olur.
		if tx.Error != nil && !errors.Is(tx.Error, gorm.ErrInvalidTransaction) { // gorm.ErrInvalidTransaction commit veya rollback yapıldığını gösterir
			log.Println("Rolling back transaction due to error during initialization.")
			tx.Rollback()
		}
	}()

	log.Println("Starting database initialization...")

	log.Println("Running migrations using GORM AutoMigrate functions...")
	// Yeniden düzenlenmiş RunMigrationsInOrder çağrısı
	if err := RunMigrationsInOrder(tx); err != nil {
		// Rollback defer içinde zaten ele alınacak, burada sadece loglama ve çıkış yeterli
		log.Fatal("Migration failed:", err) // Fatal otomatik olarak programı sonlandırır
	}
	log.Println("Migrations completed.")

	log.Println("Running seeders...")
	if err := CheckAndRunSeeders(tx); err != nil {
		// Rollback defer içinde zaten ele alınacak, burada sadece loglama ve çıkış yeterli
		log.Fatal("Seeding failed:", err)
	}
	log.Println("Seeders completed.")

	log.Println("Committing transaction...")
	if err := tx.Commit().Error; err != nil {
		// Rollback defer içinde zaten ele alınacak, burada sadece loglama ve çıkış yeterli
		log.Fatal("Commit failed:", err)
	}

	log.Println("✅ Database initialization completed successfully")
}

// executeSQL ve addConstraint fonksiyonları artık burada gerekli değil, kaldırılabilir.

// RunMigrationsInOrder fonksiyonunu GORM migration fonksiyonlarını çağıracak şekilde güncelle
func RunMigrationsInOrder(db *gorm.DB) error {
	// Foreign key bağımlılığı nedeniyle önce Teams tablosunu migrate etmeliyiz.
	// Users tablosu Teams tablosuna referans veriyor (team_id).
	log.Println(" -> Running Team migrations (GORM)...")
	if err := migrations.MigrateTeamsTable(db); err != nil {
		log.Printf("    Failed to migrate Teams table: %v", err)
		// Hata mesajına daha fazla bağlam ekle
		return fmt.Errorf("failed during team migration: %w", err)
	}
	log.Println("    Team migrations completed.")

	// Sonra Users tablosunu migrate et.
	// MigrateUsersTable fonksiyonu ENUM oluşturmayı ve foreign key'i de ele alıyor.
	log.Println(" -> Running User migrations (GORM)...")
	if err := migrations.MigrateUsersTable(db); err != nil {
		log.Printf("    Failed to migrate Users table: %v", err)
		// Hata mesajına daha fazla bağlam ekle
		return fmt.Errorf("failed during user migration: %w", err)
	}
	log.Println("    User migrations completed.")

	log.Println("All GORM-based migrations executed successfully.")
	return nil
}

// CheckAndRunSeeders fonksiyonu önceki haliyle aynı kalabilir
func CheckAndRunSeeders(db *gorm.DB) error {
	systemUser := seeders.GetSystemUserConfig()
	var existingUser models.User
	// FirstOrInit yerine First kullanmak daha uygun, yoksa seed yerine boş kayıt oluşturabilir.
	result := db.Where("account = ? AND type = ?", systemUser.Account, models.System).First(&existingUser)

	if result.Error != nil {
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			log.Printf("Creating system user: %s (%s)...\n", systemUser.Name, systemUser.Account)
			if err := seeders.SeedSystemUser(db); err != nil {
				// Hata mesajına daha fazla bağlam ekle
				return fmt.Errorf("failed to seed system user: %w", err)
			}
			log.Println(" -> System user created.")
		} else {
			// Veritabanı hatası oluştu
			log.Printf("Error checking for system user: %v\n", result.Error)
			return fmt.Errorf("database error while checking for system user: %w", result.Error)
		}
	} else {
		// Kullanıcı zaten var
		log.Printf("System user '%s' (%s) already exists, skipping creation.\n",
			existingUser.Name, existingUser.Account)
	}
	return nil
}
