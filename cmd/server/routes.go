package main

import (
	"os"

	"github.com/gin-gonic/gin"
	"github.com/imama2/Genzite-Backend/internal/core/middleware"
)

func runtimeAuthMiddleware(runtime *serverRuntime) gin.HandlerFunc {
	return middleware.JWTAuth(runtime.cfg, runtime.modules.iam.AuthProvider())
}

func registerRoutes(runtime *serverRuntime) {
	api := runtime.application.Router.Group("/api/v1")

	if runtime.cfg.IAMEnabled {
		runtime.modules.iam.RegisterRoutes(api, runtime.authMiddleware)
	}

	if runtime.cfg.MigrationsEnabled {
		if !runtime.cfg.IAMEnabled {
			runtime.logger.Error("migrations requires iam service")
			os.Exit(1)
		}
		runtime.modules.migrations.RegisterRoutes(api, runtime.authMiddleware)
	}

	if runtime.cfg.WebBuilderEnabled {
		if !runtime.cfg.IAMEnabled {
			runtime.logger.Error("web-builder requires iam service")
			os.Exit(1)
		}
		runtime.modules.webBuilder.RegisterRoutes(runtime.application.Router.Group(""), runtime.authMiddleware)
	}

	if runtime.cfg.ChatbotEnabled {
		if !runtime.cfg.IAMEnabled || !runtime.cfg.WebBuilderEnabled {
			runtime.logger.Error("chatbot requires iam and web-builder services")
			os.Exit(1)
		}
		runtime.modules.chatbot.RegisterRoutes(api, runtime.authMiddleware)
	}

	if runtime.cfg.PaymentEnabled {
		if !runtime.cfg.IAMEnabled || !runtime.cfg.WebBuilderEnabled {
			runtime.logger.Error("payment requires iam and web-builder services")
			os.Exit(1)
		}
		if err := runtime.modules.payment.SetWebBuilder(runtime.modules.webBuilder.Manager()); err != nil {
			runtime.logger.Error("failed to link web-builder to payment", "error", err)
			os.Exit(1)
		}
		runtime.modules.payment.RegisterRoutes(runtime.application.Router.Group(""), runtime.authMiddleware)
	}

	if runtime.cfg.TemplateEnabled {
		if !runtime.cfg.IAMEnabled {
			runtime.logger.Error("template requires iam service")
			os.Exit(1)
		}
		runtime.modules.template.RegisterRoutes(api, runtime.authMiddleware)
	}
}
