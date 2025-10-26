package config

import (
	"fmt"
	"os"

	"github.com/joho/godotenv"
)

type Config struct {
	// Server
	Port     string
	GinMode  string
	LogLevel string

	// Database
	DBHost     string
	DBPort     string
	DBName     string
	DBUser     string
	DBPassword string
	DBSSLMode  string

	// JWT
	JWTSecret          string
	JWTExpiry          string
	RefreshTokenExpiry string

	// Redis
	RedisURL      string
	RedisPassword string

	// OCR Service
	OCRServiceURL string

	// Storage
	StoragePath       string
	MaxFileSize       int64
	AllowedExtensions []string

	// Rate Limiting
	RateLimitRequests int
	RateLimitWindow   string

	// Features
	EnableRegistration      bool
	EnableEmailVerification bool
	EnableAPIKeys           bool
}

func Load() (*Config, error) {
	// Load .env file if it exists
	_ = godotenv.Load()

	cfg := &Config{
		Port:                    getEnv("PORT", "8080"),
		GinMode:                 getEnv("GIN_MODE", "debug"),
		LogLevel:                getEnv("LOG_LEVEL", "info"),
		DBHost:                  getEnv("DB_HOST", "localhost"),
		DBPort:                  getEnv("DB_PORT", "5432"),
		DBName:                  getEnv("POSTGRES_DB", "ocr_db"),
		DBUser:                  getEnv("POSTGRES_USER", "ocr_user"),
		DBPassword:              getEnv("POSTGRES_PASSWORD", ""),
		DBSSLMode:               getEnv("DB_SSLMODE", "disable"),
		JWTSecret:               getEnv("JWT_SECRET", ""),
		JWTExpiry:               getEnv("JWT_EXPIRY", "24h"),
		RefreshTokenExpiry:      getEnv("REFRESH_TOKEN_EXPIRY", "168h"),
		RedisURL:                getEnv("REDIS_URL", "redis://localhost:6379"),
		RedisPassword:           getEnv("REDIS_PASSWORD", ""),
		OCRServiceURL:           getEnv("OCR_SERVICE_URL", "http://localhost:8000"),
		StoragePath:             getEnv("STORAGE_PATH", "./storage"),
		MaxFileSize:             52428800, // 50MB default
		EnableRegistration:      getEnvBool("ENABLE_REGISTRATION", true),
		EnableEmailVerification: getEnvBool("ENABLE_EMAIL_VERIFICATION", false),
		EnableAPIKeys:           getEnvBool("ENABLE_API_KEYS", true),
	}

	// Validate required fields
	if cfg.JWTSecret == "" {
		return nil, fmt.Errorf("JWT_SECRET is required")
	}

	if cfg.DBPassword == "" {
		return nil, fmt.Errorf("POSTGRES_PASSWORD is required")
	}

	return cfg, nil
}

func getEnv(key, defaultValue string) string {
	value := os.Getenv(key)
	if value == "" {
		return defaultValue
	}
	return value
}

func getEnvBool(key string, defaultValue bool) bool {
	value := os.Getenv(key)
	if value == "" {
		return defaultValue
	}
	return value == "true" || value == "1"
}
