package migrations

import (
	"log/slog"

	"github.com/gin-gonic/gin"
	"github.com/imama2/Genzite-Backend/internal/core/broker"
	"github.com/imama2/Genzite-Backend/internal/core/config"
	"github.com/imama2/Genzite-Backend/internal/core/middleware"
	"github.com/imama2/Genzite-Backend/internal/services/migrations/controller"
	"github.com/imama2/Genzite-Backend/internal/services/migrations/service"
	"gorm.io/gorm"
)

type Module struct {
	logger     *slog.Logger
	cfg        *config.Config
	service    *service.MigrationService
	controller *controller.MigrationController
}

func NewModule(logger *slog.Logger) *Module {
	return &Module{logger: logger}
}

func (m *Module) Name() string {
	return "migrations"
}

func (m *Module) Init(cfg *config.Config, db *gorm.DB, _ *broker.Client) error {
	m.cfg = cfg
	m.service = service.New(cfg, db, m.logger)
	m.controller = controller.New(m.service)
	return nil
}

func (m *Module) RegisterRoutes(router *gin.RouterGroup, middlewares ...gin.HandlerFunc) {
	protected := router.Group(
		"/migrations",
		append(middlewares, middleware.RequireRoles("admin"), middleware.RequirePermissions("migrations:*"))...,
	)
	protected.POST("/up", m.controller.Up)
	protected.POST("/down", m.controller.Down)
	protected.POST("/seed", m.controller.Seed)
	protected.GET("/version", m.controller.Version)
}
