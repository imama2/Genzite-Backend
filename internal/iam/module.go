package iam

import (
	"log/slog"

	"github.com/gin-gonic/gin"
	"github.com/imama2/Genzite-Backend/internal/core/broker"
	"github.com/imama2/Genzite-Backend/internal/core/config"
	"github.com/imama2/Genzite-Backend/internal/core/middleware"
	"github.com/imama2/Genzite-Backend/internal/iam/controller"
	"github.com/imama2/Genzite-Backend/internal/iam/repository"
	"github.com/imama2/Genzite-Backend/internal/iam/service"
	"gorm.io/gorm"
)

type Module struct {
	logger     *slog.Logger
	cfg        *config.Config
	repo       repository.Repository
	service    *service.AuthService
	controller *controller.AuthController
}

func NewModule(logger *slog.Logger) *Module {
	return &Module{logger: logger}
}

func (m *Module) Name() string {
	return "iam"
}

func (m *Module) Init(cfg *config.Config, db *gorm.DB, _ *broker.Client) error {
	m.cfg = cfg
	m.repo = repository.New(db)
	m.service = service.NewAuthService(cfg, m.repo, m.logger)
	m.controller = controller.NewAuthController(cfg, m.service)
	return nil
}

func (m *Module) RegisterRoutes(router *gin.RouterGroup, middlewares ...gin.HandlerFunc) {
	public := router.Group("/auth")
	public.POST("/register", m.controller.Register)
	public.POST("/login", m.controller.Login)
	public.GET("/google/login", m.controller.GoogleLogin)
	public.GET("/google/callback", m.controller.GoogleCallback)

	protected := router.Group("/auth", middlewares...)
	protected.GET("/me", m.controller.Me)
}

func (m *Module) AuthProvider() middleware.AuthProvider {
	return m.service
}
