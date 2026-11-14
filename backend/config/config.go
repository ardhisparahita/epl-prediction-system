package config

import (
	"errors"
	"fmt"
	"os"
	"strings"

	"github.com/joho/godotenv"
)

type Config struct {
	AppEnv  string
	AppPort string

	DBHost     string
	DBPort     string
	DBUser     string
	DBPassword string
	DBName     string
	DBSSLMode  string
	DBTimezone string

	MLServiceURL string
}

func Load() (Config, error) {
	if err := godotenv.Load("../.env"); err != nil &&
		!errors.Is(err, os.ErrNotExist) {
		return Config{}, fmt.Errorf("load env: %w", err)
	}

	cfg := Config{
		AppEnv:  getEnv("APP_ENV", "development"),
		AppPort: getEnv("APP_PORT", "8080"),

		DBHost:     getEnv("DB_HOST", "localhost"),
		DBPort:     getEnv("DB_PORT", "5433"),
		DBUser:     os.Getenv("DB_USER"),
		DBPassword: os.Getenv("DB_PASSWORD"),
		DBName:     os.Getenv("DB_NAME"),
		DBSSLMode:  getEnv("DB_SSLMODE", "disable"),
		DBTimezone: getEnv("DB_TIMEZONE", "Asia/Jakarta"),

		MLServiceURL: getEnv(
			"ML_SERVICE_URL",
			"http://localhost:8000",
		),
	}

	required := map[string]string{
		"DB_USER":     cfg.DBUser,
		"DB_PASSWORD": cfg.DBPassword,
		"DB_NAME":     cfg.DBName,
	}

	for key, value := range required {
		if strings.TrimSpace(value) == "" {
			return Config{}, fmt.Errorf(
				"environment variable %s is required", key,
			)
		}
	}

	return cfg, nil
}

func getEnv(key, fallback string) string {
	value := strings.TrimSpace(os.Getenv(key))

	if value == "" {
		return fallback
	}

	return value
}
