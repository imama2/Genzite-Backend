package iam

import (
	"github.com/imama2/Genzite-Backend/internal/core/broker"
	"github.com/imama2/Genzite-Backend/internal/core/config"
	"github.com/imama2/Genzite-Backend/internal/services/iam/controller"
	repository "github.com/imama2/Genzite-Backend/internal/services/iam/repository/database"
	"github.com/imama2/Genzite-Backend/internal/services/iam/services/authentication"
	"gorm.io/gorm"
)

func (m *Module) Init(cfg *config.Config, db *gorm.DB, _ *broker.Client) error {
	m.cfg = cfg
	m.repo = repository.New(db)
	m.service = authentication.New(cfg, m.repo, m.logger)
	m.controller = controller.NewAuthController(cfg, m.service)
	return nil
}
