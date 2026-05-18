package config

import (
	"fmt"
	"os"

	"github.com/joho/godotenv"
)

type Config struct {
	AppPort string
	AppEnv  string

	DBHost     string
	DBPort     string
	DBUser     string
	DBPassword string
	DBName     string
	DBSSLMode  string

	JWTSecret      string
	JWTExpiryHours string

	SMSProvider string
	SMSAPIKey   string
	SMSSenderID string
}

// AppConfig is the global config instance
var AppConfig *Config

func InitConfig() {
	// Load .env file (ignore error in production — env vars may be set directly)
	_ = godotenv.Load()

	AppConfig = &Config{
		AppPort:        getEnv("APP_PORT", "8080"),
		AppEnv:         getEnv("APP_ENV", "development"),
		DBHost:         getEnv("DB_HOST", "localhost"),
		DBPort:         getEnv("DB_PORT", "5432"),
		DBUser:         getEnv("DB_USER", "postgres"),
		DBPassword:     getEnv("DB_PASSWORD", "admin123"),
		DBName:         getEnv("DB_NAME", "forest_management"),
		DBSSLMode:      getEnv("DB_SSLMODE", "disable"),
		JWTSecret:      getEnv("JWT_SECRET", "JWT_SECRET=9c1f6a8e3d7b4f2e8a1c9d5b7f3e6a8c2d9f1b4e7c5a3d8f0b6c9e2a1d7f4b8c"),
		JWTExpiryHours: getEnv("JWT_EXPIRY_HOURS", "72"),
		SMSProvider:    getEnv("SMS_PROVIDER", "sparrow"),
		SMSAPIKey:      getEnv("SMS_API_KEY", ""),
		SMSSenderID:    getEnv("SMS_SENDER_ID", "BanSamiti"),
	}

	fmt.Println("✅ Configuration loaded")
}

func getEnv(key, fallback string) string {
	if value, exists := os.LookupEnv(key); exists {
		return value
	}
	return fallback
}
