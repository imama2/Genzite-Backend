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
	Env                        string
	ServerPort                 string
	DatabaseURL                string
	DatabaseUsername           string
	DatabasePassword           string
	IAMEnabled                 bool
	MigrationsEnabled          bool
	MigrationsRequireAdminRole bool
	WebBuilderEnabled          bool
	ChatbotEnabled             bool
	PaymentEnabled             bool
	NotificationEnabled        bool
	TemplateEnabled            bool
	RedisURL                   string
	LogLevel                   string
	JWTSecret                  string
	JWTIssuer                  string
	JWTTTL                     time.Duration
	GoogleClientID             string
	GoogleClientSecret         string
	GoogleRedirectURL          string
	MigrationsPath             string
	AdminEmail                 string
	AdminPassword              string
	AdminName                  string
	ServeMode                  string
	WebBuilderPath             string
	OpenAIAPIKey               string
	OpenAIModel                string
	OpenAIBaseURL              string
	MidtransServerKey          string
	MidtransBaseURL            string
	RabbitMQURL                string
	RabbitMQQueue              string
	SMTPHost                   string
	SMTPPort                   string
	SMTPUser                   string
	SMTPPassword               string
	SMTPFrom                   string
	SMTPFromName               string
}

func Load() (*Config, error) {
	_ = godotenv.Load(".env")

	databaseURL := os.Getenv("DATABASE_URL")
	if databaseURL == "" {
		databaseURL = os.Getenv("POSTGRES_DSN")
	}
	if databaseURL == "" {
		return nil, errors.New("DATABASE_URL or POSTGRES_DSN is required")
	}

	databaseUsername := os.Getenv("DATABASE_USERNAME")
	if databaseUsername == "" {
		return nil, errors.New("DATABASE_USERNAME is required")
	}

	databasePassword := os.Getenv("DATABASE_PASSWORD")
	if databasePassword == "" {
		return nil, errors.New("DATABASE_PASSWORD is required")
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
		Env:                        getEnv("APP_ENV", "development"),
		ServerPort:                 getEnv("PORT", "8080"),
		DatabaseURL:                databaseURL,
		DatabaseUsername:           databaseUsername,
		DatabasePassword:           databasePassword,
		IAMEnabled:                 getBoolEnv("IAM_ENABLED"),
		MigrationsEnabled:          getBoolEnv("MIGRATIONS_ENABLED"),
		MigrationsRequireAdminRole: getBoolEnvWithDefault("MIGRATIONS_REQUIRE_ADMIN_ROLE", true),
		WebBuilderEnabled:          getBoolEnv("WEB_BUILDER_ENABLED"),
		ChatbotEnabled:             getBoolEnv("CHATBOT_ENABLED"),
		PaymentEnabled:             getBoolEnv("PAYMENT_ENABLED"),
		NotificationEnabled:        getBoolEnv("NOTIFICATION_ENABLED"),
		TemplateEnabled:            getBoolEnv("TEMPLATE_ENABLED"),
		RedisURL:                   getEnv("REDIS_URL", "redis://localhost:6379"),
		LogLevel:                   getEnv("LOG_LEVEL", "info"),
		JWTSecret:                  jwtSecret,
		JWTIssuer:                  getEnv("JWT_ISSUER", "genzite"),
		JWTTTL:                     time.Duration(jwtTTLMinutes) * time.Minute,
		GoogleClientID:             getEnv("GOOGLE_CLIENT_ID", ""),
		GoogleClientSecret:         getEnv("GOOGLE_CLIENT_SECRET", ""),
		GoogleRedirectURL:          getEnv("GOOGLE_REDIRECT_URL", ""),
		MigrationsPath:             getEnv("MIGRATIONS_PATH", "migrations"),
		AdminEmail:                 getEnv("ADMIN_EMAIL", ""),
		AdminPassword:              getEnv("ADMIN_PASSWORD", ""),
		AdminName:                  getEnv("ADMIN_NAME", "Administrator"),
		ServeMode:                  getEnv("SERVE_MODE", "manual"),
		WebBuilderPath:             getEnv("WEB_BUILDER_STORAGE_PATH", getEnv("PORTFOLIO_STORAGE_PATH", "storage/web-builder")),
		OpenAIAPIKey:               getEnv("OPENAI_API_KEY", ""),
		OpenAIModel:                getEnv("OPENAI_MODEL", "gpt-4o-mini"),
		OpenAIBaseURL:              getEnv("OPENAI_BASE_URL", "https://api.openai.com/v1"),
		MidtransServerKey:          getEnv("MIDTRANS_SERVER_KEY", ""),
		MidtransBaseURL:            getEnv("MIDTRANS_BASE_URL", "https://app.sandbox.midtrans.com"),
		RabbitMQURL:                getEnv("RABBITMQ_URL", ""),
		RabbitMQQueue:              getEnv("RABBITMQ_QUEUE", "notifications.email"),
		SMTPHost:                   getEnv("SMTP_HOST", ""),
		SMTPPort:                   getEnv("SMTP_PORT", "587"),
		SMTPUser:                   getEnv("SMTP_USER", ""),
		SMTPPassword:               getEnv("SMTP_PASSWORD", ""),
		SMTPFrom:                   getEnv("SMTP_FROM", ""),
		SMTPFromName:               getEnv("SMTP_FROM_NAME", ""),
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

func getBoolEnv(key string) bool {
	value := strings.TrimSpace(strings.ToLower(os.Getenv(key)))
	switch value {
	case "1", "true", "yes", "on":
		return true
	default:
		return false
	}
}

func getBoolEnvWithDefault(key string, fallback bool) bool {
	value := strings.TrimSpace(strings.ToLower(os.Getenv(key)))
	if value == "" {
		return fallback
	}
	switch value {
	case "1", "true", "yes", "on":
		return true
	default:
		return false
	}
}
