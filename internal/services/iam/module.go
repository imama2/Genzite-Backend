package iam

import (
	"log/slog"

	"github.com/imama2/Genzite-Backend/internal/core/config"
	"github.com/imama2/Genzite-Backend/internal/services/iam/controller"
	repository "github.com/imama2/Genzite-Backend/internal/services/iam/repository/database"
	"github.com/imama2/Genzite-Backend/internal/services/iam/services/authentication"
)

type Module struct {
	logger     *slog.Logger
	cfg        *config.Config
	repo       repository.Repository
	service    *authentication.Service
	controller *controller.AuthController
}

func NewModule(logger *slog.Logger) *Module {
	return &Module{logger: logger}
}

func (m *Module) Name() string {
	return "iam"
}
