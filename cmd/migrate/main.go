package main

import (
	"context"
	"log/slog"
	"os"

	"github.com/imama2/Genzite-Backend/internal/core/config"
	"github.com/imama2/Genzite-Backend/internal/core/db"
	"github.com/imama2/Genzite-Backend/internal/core/logging"
	"github.com/imama2/Genzite-Backend/internal/services/migrations/service"
)

func main() {
	logger := logging.NewLogger("info")

	cfg, err := config.Load()
	if err != nil {
		logger.Error("failed to load config", slog.Any("error", err))
		os.Exit(1)
	}

	database, err := db.Connect(cfg, logger)
	if err != nil {
		logger.Error("failed to connect database", slog.Any("error", err))
		os.Exit(1)
	}

	migrator := service.New(cfg, database, logger)

	if _, err := migrator.Up(context.Background(), 0); err != nil {
		logger.Error("failed to run migrations", slog.Any("error", err))
		os.Exit(1)
	}

	if err := migrator.Seed(context.Background()); err != nil {
		logger.Error("failed to run seeders", slog.Any("error", err))
		os.Exit(1)
	}
}
