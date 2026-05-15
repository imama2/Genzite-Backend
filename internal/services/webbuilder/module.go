package webbuilder

import (
	"log/slog"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/imama2/Genzite-Backend/internal/core/broker"
	"github.com/imama2/Genzite-Backend/internal/core/config"
	"github.com/imama2/Genzite-Backend/internal/services/webbuilder/handler"
	"github.com/imama2/Genzite-Backend/internal/services/webbuilder/repository"
	"github.com/imama2/Genzite-Backend/internal/services/webbuilder/service"
	"gorm.io/gorm"
)

type Module struct {
	logger  *slog.Logger
	cfg     *config.Config
	repo    repository.Repository
	service *service.SiteService
	handler *handler.SiteHandler
}

func NewModule(logger *slog.Logger) *Module {
	return &Module{logger: logger}
}

func (m *Module) Name() string {
	return "web-builder"
}

func (m *Module) Init(cfg *config.Config, db *gorm.DB, _ *broker.Client) error {
	m.cfg = cfg
	m.repo = repository.New(db)
	siteService, err := service.NewSiteService(cfg, m.repo, m.logger)
	if err != nil {
		return err
	}
	m.service = siteService
	m.handler = handler.New(siteService)
	return nil
}

func (m *Module) RegisterRoutes(router *gin.RouterGroup, middlewares ...gin.HandlerFunc) {
	api := router.Group("/api/v1/web-builder", middlewares...)
	api.POST("/sites", m.handler.CreateSite)
	api.POST("/sites/:id/publish", m.handler.PublishSite)

	if isAutoServe(m.cfg.ServeMode) {
		router.GET("/:slug", m.handler.ServeSite)
	}
}

func (m *Module) Manager() service.SiteManager {
	return m.service
}

func isAutoServe(mode string) bool {
	return strings.EqualFold(strings.TrimSpace(mode), "auto")
}

