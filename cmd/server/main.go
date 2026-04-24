package main

import (
	"os"

	"github.com/imama2/Genzite-Backend/internal/core/app"
	"github.com/imama2/Genzite-Backend/internal/core/config"
	"github.com/imama2/Genzite-Backend/internal/core/db"
	"github.com/imama2/Genzite-Backend/internal/core/logging"
	"github.com/imama2/Genzite-Backend/internal/core/middleware"
	"github.com/imama2/Genzite-Backend/internal/core/module"
	"github.com/imama2/Genzite-Backend/internal/iam"
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

	iamModule := iam.NewModule(logger)
	modules := []module.Module{
		iamModule,
	}

	enabled := make(map[string]struct{}, len(cfg.EnabledServices))
	for _, name := range cfg.EnabledServices {
		enabled[name] = struct{}{}
	}
	enableAll := len(enabled) == 0

	for _, mod := range modules {
		if !enableAll {
			if _, ok := enabled[mod.Name()]; !ok {
				logger.Info("service disabled", "service", mod.Name())
				continue
			}
		}

		if err := mod.Init(cfg, database, nil); err != nil {
			logger.Error("failed to init module", "module", mod.Name(), "error", err)
			os.Exit(1)
		}
	}

	api := application.Router.Group("/api/v1")
	if enableAll || hasService(enabled, iamModule.Name()) {
		apiMiddleware := middleware.JWTAuth(cfg, iamModule.AuthProvider())
		iamModule.RegisterRoutes(api, apiMiddleware)
	}

	addr := ":" + cfg.ServerPort
	logger.Info("starting server", "addr", addr)
	if err := application.Router.Run(addr); err != nil {
		logger.Error("server stopped with error", "error", err)
		os.Exit(1)
	}
}

func hasService(enabled map[string]struct{}, name string) bool {
	_, ok := enabled[name]
	return ok
}
