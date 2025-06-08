package config

import (
	"log"
	"os"

	"github.com/joho/godotenv"
)

// Config holds the application configuration
type Config struct {
	JWTSecret string
	DBHost    string
	DBUser    string
	DBPassword string
	DBName    string
	DBPort    string
}

// AppConfig is the global configuration variable
var AppConfig Config

// LoadConfig loads configuration from environment variables
func LoadConfig() {
	appEnv := os.Getenv("APP_ENV")
	envFile := ".env"
	if appEnv == "test" {
		envFile = ".env.test"
		log.Println("Running in test environment, attempting to load .env.test")
	}

	// Load .env file if it exists
	if err := godotenv.Load(envFile); err != nil {
		log.Printf("No %s file found, relying on environment variables or defaults. Error: %v\n", envFile, err)
	}

	AppConfig.JWTSecret = os.Getenv("JWT_SECRET")
	if AppConfig.JWTSecret == "" {
		log.Fatal("JWT_SECRET environment variable is required")
	}

	AppConfig.DBHost = os.Getenv("DB_HOST")
	AppConfig.DBUser = os.Getenv("DB_USER")
	AppConfig.DBPassword = os.Getenv("DB_PASSWORD")
	AppConfig.DBName = os.Getenv("DB_NAME")
	AppConfig.DBPort = os.Getenv("DB_PORT")

	// Basic check for essential DB config
	if AppConfig.DBHost == "" || AppConfig.DBUser == "" || AppConfig.DBName == "" || AppConfig.DBPort == "" {
		log.Println("Warning: One or more database connection environment variables (DB_HOST, DB_USER, DB_NAME, DB_PORT) are not set.")
		// Depending on the application's needs, you might want to Fatal here if DB is essential at startup
		// For now, we allow it to proceed as ConnectDB will handle the fatal error if connection fails.
	}
}
