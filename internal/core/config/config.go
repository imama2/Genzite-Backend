package config

import (
	"errors"
	"os"
	"strings"

	"github.com/joho/godotenv"
)

type Config struct {
	Env             string
	ServerPort      string
	DatabaseURL     string
	EnabledServices []string
	LogLevel        string
}

func Load() (*Config, error) {
	_ = godotenv.Load("config/.env", ".env")

	databaseURL := os.Getenv("DATABASE_URL")
	if databaseURL == "" {
		databaseURL = os.Getenv("POSTGRES_DSN")
	}
	if databaseURL == "" {
		return nil, errors.New("DATABASE_URL or POSTGRES_DSN is required")
	}

	cfg := &Config{
		Env:             getEnv("APP_ENV", "development"),
		ServerPort:      getEnv("PORT", "8080"),
		DatabaseURL:     databaseURL,
		EnabledServices: splitEnv("ENABLED_SERVICES"),
		LogLevel:        getEnv("LOG_LEVEL", "info"),
	}

	return cfg, nil
}

func getEnv(key, fallback string) string {
	value := os.Getenv(key)
	if value == "" {
		return fallback
	}
	return value
}

func splitEnv(key string) []string {
	value := strings.TrimSpace(os.Getenv(key))
	if value == "" {
		return nil
	}

	parts := strings.Split(value, ",")
	services := make([]string, 0, len(parts))
	for _, part := range parts {
		trimmed := strings.TrimSpace(part)
		if trimmed != "" {
			services = append(services, trimmed)
		}
	}

	return services
}
