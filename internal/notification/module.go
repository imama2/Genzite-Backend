package notification

import (
	"context"
	"log/slog"

	"github.com/gin-gonic/gin"
	"github.com/imama2/Genzite-Backend/internal/core/broker"
	"github.com/imama2/Genzite-Backend/internal/core/config"
	"github.com/imama2/Genzite-Backend/internal/notification/service"
	"gorm.io/gorm"
)

type Module struct {
	logger  *slog.Logger
	cfg     *config.Config
	service *service.EmailService
}

func NewModule(logger *slog.Logger) *Module {
	return &Module{logger: logger}
}

func (m *Module) Name() string {
	return "notification"
}

func (m *Module) Init(cfg *config.Config, _ *gorm.DB, brokerClient *broker.Client) error {
	m.cfg = cfg
	m.service = service.New(cfg, brokerClient, m.logger)

	go func() {
		if err := m.service.StartConsumer(context.Background()); err != nil {
			m.logger.Error("notification consumer stopped", "error", err)
		}
	}()

	return nil
}

func (m *Module) RegisterRoutes(_ *gin.RouterGroup, _ ...gin.HandlerFunc) {}

func (m *Module) Notifier() *service.EmailService {
	return m.service
}
