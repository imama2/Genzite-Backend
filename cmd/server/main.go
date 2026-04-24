package main

import (
	"os"

	"github.com/imama2/Genzite-Backend/internal/core/app"
	"github.com/imama2/Genzite-Backend/internal/core/config"
	"github.com/imama2/Genzite-Backend/internal/core/db"
	"github.com/imama2/Genzite-Backend/internal/core/logging"
)

func main() {
	logger := logging.NewLogger("info")

	cfg, err := config.Load()
	if err != nil {
		logger.Error("failed to load config", "error", err)
		os.Exit(1)
	}

	logger = logging.NewLogger(cfg.LogLevel)

	database, err := db.Connect(cfg, logger)
	if err != nil {
		logger.Error("failed to connect to database", "error", err)
		os.Exit(1)
	}

	application := app.NewApp(cfg, database, logger)
	application.RegisterHealth()

	addr := ":" + cfg.ServerPort
	logger.Info("starting server", "addr", addr)
	if err := application.Router.Run(addr); err != nil {
		logger.Error("server stopped with error", "error", err)
		os.Exit(1)
	}
}
