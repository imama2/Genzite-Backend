package main

import (
	"os"

	"github.com/gin-gonic/gin"
	"github.com/imama2/Genzite-Backend/internal/core/app"
	"github.com/imama2/Genzite-Backend/internal/core/broker"
	"github.com/imama2/Genzite-Backend/internal/core/config"
	"github.com/imama2/Genzite-Backend/internal/core/db"
	"github.com/imama2/Genzite-Backend/internal/core/logging"
	"github.com/imama2/Genzite-Backend/internal/core/middleware"
	"github.com/imama2/Genzite-Backend/internal/core/module"
	"github.com/imama2/Genzite-Backend/internal/chatbot"
	"github.com/imama2/Genzite-Backend/internal/iam"
	"github.com/imama2/Genzite-Backend/internal/migrations"
	"github.com/imama2/Genzite-Backend/internal/notification"
	"github.com/imama2/Genzite-Backend/internal/payment"
	"github.com/imama2/Genzite-Backend/internal/webbuilder"
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
	migrationsModule := migrations.NewModule(logger)
	webBuilderModule := webbuilder.NewModule(logger)
	chatbotModule := chatbot.NewModule(logger)
	paymentModule := payment.NewModule(logger)
	notificationModule := notification.NewModule(logger)
	modules := []module.Module{
		iamModule,
		migrationsModule,
		webBuilderModule,
		chatbotModule,
		paymentModule,
		notificationModule,
	}

	enabled := make(map[string]struct{}, len(cfg.EnabledServices))
	for _, name := range cfg.EnabledServices {
		enabled[name] = struct{}{}
	}
	enableAll := len(enabled) == 0

	iamEnabled := enableAll || hasService(enabled, iamModule.Name())
	migrationsEnabled := enableAll || hasService(enabled, migrationsModule.Name())
	webBuilderEnabled := enableAll || hasService(enabled, webBuilderModule.Name())
	chatbotEnabled := enableAll || hasService(enabled, chatbotModule.Name())
	paymentEnabled := enableAll || hasService(enabled, paymentModule.Name())
	notificationEnabled := enableAll || hasService(enabled, notificationModule.Name())

	var brokerClient *broker.Client
	if notificationEnabled {
		client, err := broker.NewClient(cfg, logger)
		if err != nil {
			logger.Error("failed to connect broker", "error", err)
			os.Exit(1)
		}
		brokerClient = client
	}

	for _, mod := range modules {
		if !enableAll {
			if _, ok := enabled[mod.Name()]; !ok {
				logger.Info("service disabled", "service", mod.Name())
				continue
			}
		}

		if err := mod.Init(cfg, database, brokerClient); err != nil {
			logger.Error("failed to init module", "module", mod.Name(), "error", err)
			os.Exit(1)
		}
	}

	api := application.Router.Group("/api/v1")

	var authMiddleware gin.HandlerFunc
	if iamEnabled {
		authMiddleware = middleware.JWTAuth(cfg, iamModule.AuthProvider())
		iamModule.RegisterRoutes(api, authMiddleware)
	}

	if migrationsEnabled {
		if !iamEnabled {
			logger.Error("migrations requires iam service")
			os.Exit(1)
		}
		migrationsModule.RegisterRoutes(api, authMiddleware)
	}

	if webBuilderEnabled {
		if !iamEnabled {
			logger.Error("web-builder requires iam service")
			os.Exit(1)
		}
		webBuilderModule.RegisterRoutes(application.Router.Group(""), authMiddleware)
	}

	if chatbotEnabled {
		if !iamEnabled || !webBuilderEnabled {
			logger.Error("chatbot requires iam and web-builder services")
			os.Exit(1)
		}
		if err := chatbotModule.SetWebBuilder(webBuilderModule.Manager()); err != nil {
			logger.Error("failed to link web-builder to chatbot", "error", err)
			os.Exit(1)
		}
		chatbotModule.RegisterRoutes(api, authMiddleware)
	}

	if paymentEnabled {
		if !iamEnabled || !webBuilderEnabled {
			logger.Error("payment requires iam and web-builder services")
			os.Exit(1)
		}
		if err := paymentModule.SetWebBuilder(webBuilderModule.Manager()); err != nil {
			logger.Error("failed to link web-builder to payment", "error", err)
			os.Exit(1)
		}
		paymentModule.RegisterRoutes(application.Router.Group(""), authMiddleware)
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
