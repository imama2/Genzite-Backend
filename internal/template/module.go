package template

import (
	"log/slog"

	"github.com/gin-gonic/gin"
	"github.com/imama2/Genzite-Backend/internal/core/broker"
	"github.com/imama2/Genzite-Backend/internal/core/config"
	"github.com/imama2/Genzite-Backend/internal/template/handler"
	"github.com/imama2/Genzite-Backend/internal/template/repository"
	"github.com/imama2/Genzite-Backend/internal/template/service"
	"gorm.io/gorm"
)

type Module struct {
	logger  *slog.Logger
	cfg     *config.Config
	repo    repository.Repository
	service *service.Service
	handler *handler.Handler
}

func NewModule(logger *slog.Logger) *Module {
	return &Module{logger: logger}
}

func (m *Module) Name() string {
	return "template"
}

func (m *Module) Init(cfg *config.Config, _ *gorm.DB, _ *broker.Client) error {
	m.cfg = cfg
	m.repo = repository.New()
	m.service = service.New(m.repo)
	m.handler = handler.New(m.service)
	return nil
}

func (m *Module) RegisterRoutes(router *gin.RouterGroup, middlewares ...gin.HandlerFunc) {
	api := router.Group("/api/v1/template", middlewares...)
	api.GET("/ping", m.handler.Ping)
}
