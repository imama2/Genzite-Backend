package payment

import (
	"log/slog"

	"github.com/gin-gonic/gin"
	"github.com/imama2/Genzite-Backend/internal/core/broker"
	"github.com/imama2/Genzite-Backend/internal/core/config"
	"github.com/imama2/Genzite-Backend/internal/services/payment/controller"
	"github.com/imama2/Genzite-Backend/internal/services/payment/midtrans"
	"github.com/imama2/Genzite-Backend/internal/services/payment/repository"
	"github.com/imama2/Genzite-Backend/internal/services/payment/service"
	webbuilder "github.com/imama2/Genzite-Backend/internal/services/webbuilder/service"
	"gorm.io/gorm"
)

type Module struct {
	logger     *slog.Logger
	cfg        *config.Config
	repo       repository.Repository
	client     *midtrans.Client
	service    *service.PaymentService
	controller *controller.PaymentController
	webBuilder webbuilder.SiteManager
}

func NewModule(logger *slog.Logger) *Module {
	return &Module{logger: logger}
}

func (m *Module) Name() string {
	return "payment"
}

func (m *Module) Init(cfg *config.Config, db *gorm.DB, broker broker.BrokerService) error {
	m.cfg = cfg
	m.repo = repository.New(db)
	client, err := midtrans.New(cfg)
	if err != nil {
		return err
	}
	m.client = client
	return nil
}

func (m *Module) SetWebBuilder(manager webbuilder.SiteManager) error {
	if manager == nil {
		return service.ErrWebBuilderMissing
	}
	m.webBuilder = manager
	m.service = service.New(m.cfg, m.repo, m.client, manager, m.logger)
	m.controller = controller.New(m.service)
	return nil
}

func (m *Module) RegisterRoutes(router *gin.RouterGroup, middlewares ...gin.HandlerFunc) {
	protected := router.Group("/api/v1/payment", middlewares...)
	protected.POST("/transactions", m.controller.CreateTransaction)

	router.POST("/api/v1/payment/webhook", m.controller.Webhook)
}
