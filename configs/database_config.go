package configs

import (
	"fmt"
	"log"
	"os"
	"strconv"
	"time"

	"zatrano/utils"

	"github.com/joho/godotenv"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

// DB represents the database connection
var DB *gorm.DB

// DatabaseConfig holds database configuration parameters
type DatabaseConfig struct {
	Host     string
	Port     int
	User     string
	Password string
	Name     string
	SSLMode  string // PostgreSQL için yeni alan
	TimeZone string // PostgreSQL için yeni alan
}

// loadEnv loads environment variables from .env file
func loadEnv() {
	// Load .env file from different locations (with fallback)
	envFiles := []string{".env", "../.env", "../../.env"}
	var loaded bool

	for _, file := range envFiles {
		if _, err := os.Stat(file); err == nil {
			if err := godotenv.Load(file); err == nil {
				log.Printf("Loaded environment variables from %s", file)
				loaded = true
				break
			}
		}
	}

	if !loaded {
		log.Println("No .env file found, using system environment variables")
	}
}

// InitDB initializes the database connection
func InitDB() {
	// First load environment variables
	loadEnv()

	portStr := utils.GetEnvWithDefault("DB_PORT", "5432") // Varsayılan port 5432 olarak değişti
	port, err := strconv.Atoi(portStr)
	if err != nil {
		log.Fatalf("Invalid DB_PORT: %v", err)
	}

	dbConfig := DatabaseConfig{
		Host:     utils.GetEnvWithDefault("DB_HOST", "localhost"),
		Port:     port,
		User:     utils.GetEnvWithDefault("DB_USERNAME", "postgres"), // Varsayılan kullanıcı postgres
		Password: utils.GetEnvWithDefault("DB_PASSWORD", ""),
		Name:     utils.GetEnvWithDefault("DB_DATABASE", "myapp"),
		SSLMode:  utils.GetEnvWithDefault("DB_SSL_MODE", "disable"), // SSL modu eklendi
		TimeZone: utils.GetEnvWithDefault("DB_TIMEZONE", "UTC"),     // Zaman dilimi eklendi
	}

	log.Printf("Database Config: Host=%s, Port=%d, User=%s, DB=%s, SSLMode=%s, TimeZone=%s",
		dbConfig.Host, dbConfig.Port, dbConfig.User, dbConfig.Name, dbConfig.SSLMode, dbConfig.TimeZone)

	// PostgreSQL DSN (Data Source Name) oluşturma
	dsn := fmt.Sprintf("host=%s user=%s password=%s dbname=%s port=%d sslmode=%s TimeZone=%s",
		dbConfig.Host,
		dbConfig.User,
		dbConfig.Password,
		dbConfig.Name,
		dbConfig.Port,
		dbConfig.SSLMode,
		dbConfig.TimeZone)

	var gormerr error
	DB, gormerr = gorm.Open(postgres.Open(dsn), &gorm.Config{
		Logger: logger.Default.LogMode(getGormLogLevel()),
		NowFunc: func() time.Time {
			return time.Now().UTC()
		},
	})

	if gormerr != nil {
		log.Fatalf("Failed to connect to database: %v\nDSN: %s", gormerr, maskPasswordInDSN(dsn))
	}

	// Configure connection pool
	sqlDB, err := DB.DB()
	if err != nil {
		log.Fatalf("Failed to get database instance: %v", err)
	}

	// Connection pool settings
	sqlDB.SetMaxIdleConns(utils.GetEnvAsInt("DB_MAX_IDLE_CONNS", 10))
	sqlDB.SetMaxOpenConns(utils.GetEnvAsInt("DB_MAX_OPEN_CONNS", 100))
	sqlDB.SetConnMaxLifetime(time.Duration(utils.GetEnvAsInt("DB_CONN_MAX_LIFETIME_MINUTES", 60)) * time.Minute)

	log.Println("Database connection established successfully")
}

// Helper function to determine GORM log level
func getGormLogLevel() logger.LogLevel {
	switch os.Getenv("DB_LOG_LEVEL") {
	case "silent":
		return logger.Silent
	case "error":
		return logger.Error
	case "warn":
		return logger.Warn
	default:
		return logger.Info
	}
}

// Helper function to mask password in DSN for logging
func maskPasswordInDSN(dsn string) string {
	// Simple masking - replace password with ****
	// Note: This is a basic implementation, adjust as needed
	return dsn[:len(dsn)-len("your_password")] + "****"
}

// GetDB returns the database instance
func GetDB() *gorm.DB {
	if DB == nil {
		log.Fatal("Database connection not initialized. Call InitDB() first.")
	}
	return DB
}

// CloseDB closes the database connection
func CloseDB() error {
	if DB == nil {
		return nil
	}

	db, err := DB.DB()
	if err != nil {
		return fmt.Errorf("failed to get database instance: %v", err)
	}
	return db.Close()
}
