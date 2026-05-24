package chatbot

import (
	"log/slog"

	"github.com/gin-gonic/gin"
	"github.com/imama2/Genzite-Backend/internal/core/broker"
	"github.com/imama2/Genzite-Backend/internal/core/config"
	"github.com/imama2/Genzite-Backend/internal/services/agents"
	"gorm.io/gorm"
)

type Module struct {
	logger       *slog.Logger
	cfg          *config.Config
	ChatService  *ChatService
	AgentService agents.AgentService
	BrokerClient broker.BrokerService
}

func NewModule(logger *slog.Logger) *Module {
	return &Module{logger: logger}
}

func (m *Module) Name() string {
	return "chatbot"
}

func (m *Module) Init(cfg *config.Config, _ *gorm.DB, brokerClient broker.BrokerService) error {
	m.cfg = cfg

	m.BrokerClient = brokerClient

	m.AgentService = agents.NewService() // Placeholder for agent service

	m.ChatService = NewService(m.logger, m.AgentService, m.BrokerClient)
	return nil
}

// Note: The old RegisterRoutes is removed as the chatbot service now operates over WebSockets,
// which are handled by the orchestrator's main function.

func (m *Module) RegisterRoutes(group *gin.RouterGroup, authMiddleware ...gin.HandlerFunc) {
	// The chatbot service now operates over WebSockets,
	// which are handled by the orchestrator's main function.
	// This method is here to satisfy the module interface.
}
