package configs

import (
	"encoding/gob"
	"log"
	"strconv"
	"time"

	"zatrano/models"
	"zatrano/utils"

	"github.com/gofiber/fiber/v2/middleware/session"
	"github.com/gofiber/storage/postgres/v3"
)

// DatabaseConfig holds database configuration parameters
type sessionDatabaseConfig struct {
	Host     string
	Port     int
	User     string
	Password string
	Name     string
	SSLMode  string
}

var Session *session.Store

// InitSession initializes the session store (original function)
func InitSession() {
	Session = createSessionStore()
	utils.InitializeSessionStore(Session)
}

// SetupSession is an alias for InitSession for better naming consistency
func SetupSession() *session.Store {
	if Session == nil {
		InitSession()
	}
	return Session
}

func createSessionStore() *session.Store {
	portStr := utils.GetEnvWithDefault("DB_PORT", "5432")
	port, err := strconv.Atoi(portStr)
	if err != nil {
		log.Fatalf("Invalid DB_PORT: %v", err)
	}

	dbConfig := sessionDatabaseConfig{
		Host:     utils.GetEnvWithDefault("DB_HOST", "localhost"),
		Port:     port,
		User:     utils.GetEnvWithDefault("DB_USERNAME", "postgres"),
		Password: utils.GetEnvWithDefault("DB_PASSWORD", ""),
		Name:     utils.GetEnvWithDefault("DB_DATABASE", "myapp"),
		SSLMode:  utils.GetEnvWithDefault("DB_SSL_MODE", "disable"),
	}

	storage := postgres.New(postgres.Config{
		Host:       dbConfig.Host,
		Port:       dbConfig.Port,
		Username:   dbConfig.User,
		Password:   dbConfig.Password,
		Database:   dbConfig.Name,
		SSLMode:    dbConfig.SSLMode,
		Reset:      false,
		Table:      "sessions",
		GCInterval: 10 * time.Second,
	})

	store := session.New(session.Config{
		Storage:    storage,
		Expiration: 24 * time.Hour,
	})

	// Gob tiplerini kaydet
	registerGobTypes()

	return store
}

func registerGobTypes() {
	gob.Register(models.UserType("")) // UserType için gob kaydı
	gob.Register(&models.User{})      // User struct'ı için gob kaydı
}
