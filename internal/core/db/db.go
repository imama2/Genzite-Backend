package db

import (
	"fmt"
	"log/slog"
	"net/url"

	"github.com/imama2/Genzite-Backend/internal/core/config"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

func Connect(cfg *config.Config, logger *slog.Logger) (*gorm.DB, error) {
	dsn, err := buildDatabaseDSN(cfg.DatabaseURL, cfg.DatabaseUsername, cfg.DatabasePassword)
	if err != nil {
		return nil, err
	}

	database, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		return nil, fmt.Errorf("open database: %w", err)
	}

	sqlDB, err := database.DB()
	if err != nil {
		return nil, fmt.Errorf("get sql db: %w", err)
	}

	if err := sqlDB.Ping(); err != nil {
		return nil, fmt.Errorf("ping database: %w", err)
	}

	logger.Info("database connection established")
	return database, nil
}

func buildDatabaseDSN(baseURL, username, password string) (string, error) {
	parsed, err := url.Parse(baseURL)
	if err != nil {
		return "", fmt.Errorf("parse database url: %w", err)
	}

	parsed.User = url.UserPassword(username, password)
	return parsed.String(), nil
}
