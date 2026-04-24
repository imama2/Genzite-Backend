package config

import (
	"errors"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/joho/godotenv"
)

type Config struct {
	Env                string
	ServerPort         string
	DatabaseURL        string
	EnabledServices    []string
	LogLevel           string
	JWTSecret          string
	JWTIssuer          string
	JWTTTL             time.Duration
	GoogleClientID     string
	GoogleClientSecret string
	GoogleRedirectURL  string
	MigrationsPath     string
	AdminEmail         string
	AdminPassword      string
	AdminName          string
	ServeMode          string
	WebBuilderPath     string
	OpenAIAPIKey       string
	OpenAIModel        string
	OpenAIBaseURL      string
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

	jwtSecret := os.Getenv("JWT_SECRET")
	if jwtSecret == "" {
		return nil, errors.New("JWT_SECRET is required")
	}

	jwtTTLMinutesRaw := getEnv("JWT_TTL_MINUTES", "60")
	jwtTTLMinutes, err := strconv.Atoi(jwtTTLMinutesRaw)
	if err != nil || jwtTTLMinutes <= 0 {
		return nil, errors.New("JWT_TTL_MINUTES must be a positive integer")
	}

	cfg := &Config{
		Env:                getEnv("APP_ENV", "development"),
		ServerPort:         getEnv("PORT", "8080"),
		DatabaseURL:        databaseURL,
		EnabledServices:    splitEnv("ENABLED_SERVICES"),
		LogLevel:           getEnv("LOG_LEVEL", "info"),
		JWTSecret:          jwtSecret,
		JWTIssuer:          getEnv("JWT_ISSUER", "genzite"),
		JWTTTL:             time.Duration(jwtTTLMinutes) * time.Minute,
		GoogleClientID:     getEnv("GOOGLE_CLIENT_ID", ""),
		GoogleClientSecret: getEnv("GOOGLE_CLIENT_SECRET", ""),
		GoogleRedirectURL:  getEnv("GOOGLE_REDIRECT_URL", ""),
		MigrationsPath:     getEnv("MIGRATIONS_PATH", "migrations"),
		AdminEmail:         getEnv("ADMIN_EMAIL", ""),
		AdminPassword:      getEnv("ADMIN_PASSWORD", ""),
		AdminName:          getEnv("ADMIN_NAME", "Administrator"),
		ServeMode:          getEnv("SERVE_MODE", "manual"),
		WebBuilderPath:     getEnv("WEB_BUILDER_STORAGE_PATH", getEnv("PORTFOLIO_STORAGE_PATH", "storage/web-builder")),
		OpenAIAPIKey:       getEnv("OPENAI_API_KEY", ""),
		OpenAIModel:        getEnv("OPENAI_MODEL", "gpt-4o-mini"),
		OpenAIBaseURL:      getEnv("OPENAI_BASE_URL", "https://api.openai.com/v1"),
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
