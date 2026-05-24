package main

import (
	"fmt"
	"log/slog"

	"github.com/gin-gonic/gin"
	"github.com/imama2/Genzite-Backend/internal/core/app"
	"github.com/imama2/Genzite-Backend/internal/core/broker"
	"github.com/imama2/Genzite-Backend/internal/core/config"
	"github.com/imama2/Genzite-Backend/internal/core/db"
	"github.com/imama2/Genzite-Backend/internal/core/logging"
	"gorm.io/gorm"
)

type serverRuntime struct {
	cfg            *config.Config
	logger         *slog.Logger
	database       *gorm.DB
	application    *app.App
	brokerClient   broker.BrokerService
	modules        *serviceModules
	authMiddleware gin.HandlerFunc
}

func run() error {
	runtime, err := bootstrap()
	if err != nil {
		return err
	}

	if err := initModules(runtime); err != nil {
		return err
	}

	registerRoutes(runtime)

	return startServer(runtime)
}

func bootstrap() (*serverRuntime, error) {
	bootstrapLogger := logging.NewLogger("info")

	cfg, err := config.Load()
	if err != nil {
		bootstrapLogger.Error("failed to load config", "error", err)
		return nil, err
	}

	logger := logging.NewLogger(cfg.LogLevel)
	database, err := db.Connect(cfg, logger)
	if err != nil {
		logger.Error("failed to connect to database", "error", err)
		return nil, err
	}

	application := app.NewApp(cfg, database, logger)
	application.RegisterHealth()

	runtime := &serverRuntime{
		cfg:         cfg,
		logger:      logger,
		database:    database,
		application: application,
		modules:     newServiceModules(logger),
	}

	if cfg.NotificationEnabled {
		client, err := broker.NewClient(cfg, logger)
		if err != nil {
			logger.Error("failed to connect broker", "error", err)
			return nil, err
		}
		runtime.brokerClient = client
	}

	return runtime, nil
}

func startServer(runtime *serverRuntime) error {
	addr := fmt.Sprintf(":%s", runtime.cfg.ServerPort)
	runtime.logger.Info("starting server", "addr", addr)
	if err := runtime.application.Router.Run(addr); err != nil {
		runtime.logger.Error("server stopped with error", "error", err)
		return err
	}
	return nil
}
