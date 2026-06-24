package config

import (
	"log"
	"os"
	"strconv"
	"time"
)

type Config struct {
	Env           string
	Port          string
	AllowedOrigin string

	DBHost     string
	DBPort     string
	DBName     string
	DBUser     string
	DBPassword string
	DBSSLMode  string

	JWTSecret     string
	JWTExpiry     time.Duration
	RefreshExpiry time.Duration
}

// IsProd returns true if the application is running in production mode.
func (c *Config) IsProd() bool {
	return c.Env == "production"
}

// Load reads configuration from environment variables with default fallbacks.
func Load() *Config {
	jwtMinutes, err := strconv.Atoi(getEnv("JWT_EXPIRY_MINUTES", "15"))
	if err != nil {
		log.Printf("Warning: invalid JWT_EXPIRY_MINUTES. Using default: 15")
		jwtMinutes = 15
	}

	refreshDays, err := strconv.Atoi(getEnv("REFRESH_EXPIRY_DAYS", "30"))
	if err != nil {
		log.Printf("Warning: invalid REFRESH_EXPIRY_DAYS. Using default: 30")
		refreshDays = 30
	}

	return &Config{
		Env:           getEnv("ENV", "development"),
		Port:          getEnv("PORT", "8080"),
		AllowedOrigin: getEnv("ALLOWED_ORIGIN", "http://localhost:5173"),

		DBHost:     getEnv("DB_HOST", "127.0.0.1"),
		DBPort:     getEnv("DB_PORT", "5432"),
		DBName:     getEnv("DB_NAME", "lightchat_db"),
		DBUser:     getEnv("DB_USER", "postgres"),
		DBPassword: getEnv("DB_PASSWORD", ""),
		DBSSLMode:  getEnv("DB_SSLMODE", "disable"),

		JWTSecret:     getEnv("JWT_SECRET", "super-secret-key"),
		JWTExpiry:     time.Duration(jwtMinutes) * time.Minute,
		RefreshExpiry: time.Duration(refreshDays) * 24 * time.Hour,
	}
}

func getEnv(key, fallback string) string {
	if v, ok := os.LookupEnv(key); ok {
		return v
	}
	return fallback
}
