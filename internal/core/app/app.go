package app

import (
	"log/slog"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/imama2/Genzite-Backend/internal/core/config"
	"gorm.io/gorm"
)

type App struct {
	Config *config.Config
	DB     *gorm.DB
	Router *gin.Engine
	Logger *slog.Logger
}

func NewApp(cfg *config.Config, database *gorm.DB, logger *slog.Logger) *App {
	if cfg.Env == "production" {
		gin.SetMode(gin.ReleaseMode)
	}

	router := gin.New()
	router.Use(gin.Recovery())

	return &App{
		Config: cfg,
		DB:     database,
		Router: router,
		Logger: logger,
	}
}

func (a *App) RegisterHealth() {
	a.Router.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "ok"})
	})
}
