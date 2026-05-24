package module

import (
	"github.com/gin-gonic/gin"
	"github.com/imama2/Genzite-Backend/internal/core/broker"
	"github.com/imama2/Genzite-Backend/internal/core/config"
	"gorm.io/gorm"
)

type Module interface {
	Name() string
	Init(cfg *config.Config, db *gorm.DB, broker broker.BrokerService) error
	RegisterRoutes(router *gin.RouterGroup, middlewares ...gin.HandlerFunc)
}
