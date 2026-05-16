package main

import (
	"log/slog"

	"github.com/imama2/Genzite-Backend/internal/core/config"
	"github.com/imama2/Genzite-Backend/internal/core/module"
	"github.com/imama2/Genzite-Backend/internal/services/chatbot"
	"github.com/imama2/Genzite-Backend/internal/services/iam"
	"github.com/imama2/Genzite-Backend/internal/services/migrations"
	"github.com/imama2/Genzite-Backend/internal/services/notification"
	"github.com/imama2/Genzite-Backend/internal/services/payment"
	"github.com/imama2/Genzite-Backend/internal/services/template"
	"github.com/imama2/Genzite-Backend/internal/services/webbuilder"
)

type serviceModules struct {
	iam          *iam.Module
	migrations   *migrations.Module
	webBuilder   *webbuilder.Module
	chatbot      *chatbot.Module
	payment      *payment.Module
	notification *notification.Module
	template     *template.Module
}

func newServiceModules(logger *slog.Logger) *serviceModules {
	return &serviceModules{
		iam:          iam.NewModule(logger),
		migrations:   migrations.NewModule(logger),
		webBuilder:   webbuilder.NewModule(logger),
		chatbot:      chatbot.NewModule(logger),
		payment:      payment.NewModule(logger),
		notification: notification.NewModule(logger),
		template:     template.NewModule(logger),
	}
}

func initModules(runtime *serverRuntime) error {
	mods := []module.Module{
		runtime.modules.iam,
		runtime.modules.migrations,
		runtime.modules.webBuilder,
		runtime.modules.chatbot,
		runtime.modules.payment,
		runtime.modules.notification,
		runtime.modules.template,
	}

	for _, mod := range mods {
		if !moduleEnabled(runtime.cfg, mod.Name()) {
			continue
		}

		if err := mod.Init(runtime.cfg, runtime.database, runtime.brokerClient); err != nil {
			runtime.logger.Error("failed to init module", "module", mod.Name(), "error", err)
			return err
		}
	}

	if runtime.cfg.IAMEnabled {
		runtime.authMiddleware = runtimeAuthMiddleware(runtime)
	}

	return nil
}

func moduleEnabled(cfg *config.Config, name string) bool {
	switch name {
	case "iam":
		return cfg.IAMEnabled
	case "migrations":
		return cfg.MigrationsEnabled
	case "web-builder":
		return cfg.WebBuilderEnabled
	case "chatbot":
		return cfg.ChatbotEnabled
	case "payment":
		return cfg.PaymentEnabled
	case "notification":
		return cfg.NotificationEnabled
	case "template":
		return cfg.TemplateEnabled
	default:
		return false
	}
}
