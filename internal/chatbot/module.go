package chatbot

import (
	"log/slog"

	"github.com/gin-gonic/gin"
	"github.com/imama2/Genzite-Backend/internal/chatbot/ai"
	"github.com/imama2/Genzite-Backend/internal/chatbot/handler"
	"github.com/imama2/Genzite-Backend/internal/chatbot/service"
	"github.com/imama2/Genzite-Backend/internal/core/broker"
	"github.com/imama2/Genzite-Backend/internal/core/config"
	webbuilder "github.com/imama2/Genzite-Backend/internal/webbuilder/service"
	"gorm.io/gorm"
)

type Module struct {
	logger     *slog.Logger
	cfg        *config.Config
	aiClient   service.AIClient
	webBuilder webbuilder.SiteManager
	service    *service.ChatService
	handler    *handler.ChatHandler
}

func NewModule(logger *slog.Logger) *Module {
	return &Module{logger: logger}
}

func (m *Module) Name() string {
	return "chatbot"
}

func (m *Module) Init(cfg *config.Config, _ *gorm.DB, _ *broker.Client) error {
	m.cfg = cfg

	client, err := ai.NewOpenAIClient(cfg, m.logger)
	if err != nil {
		return err
	}
	m.aiClient = client
	return nil
}

func (m *Module) SetWebBuilder(manager webbuilder.SiteManager) error {
	if manager == nil {
		return service.ErrWebBuilderMissing
	}
	m.webBuilder = manager
	m.service = service.New(m.aiClient, manager)
	m.handler = handler.New(m.service)
	return nil
}

func (m *Module) RegisterRoutes(router *gin.RouterGroup, middlewares ...gin.HandlerFunc) {
	api := router.Group("/api/v1/chatbot", middlewares...)
	api.POST("/generate", m.handler.GenerateDraft)
}
