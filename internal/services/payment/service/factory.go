package service

import (
	"log/slog"

	"github.com/imama2/Genzite-Backend/internal/core/config"
	"github.com/imama2/Genzite-Backend/internal/services/payment/gateway/midtrans"
	"github.com/imama2/Genzite-Backend/internal/services/payment/repository"
	webbuilder "github.com/imama2/Genzite-Backend/internal/services/webbuilder/service"
)

type PaymentService struct {
	cfg        *config.Config
	repo       repository.Repository
	client     *midtrans.Client
	webBuilder webbuilder.SiteManager
	logger     *slog.Logger
}

func New(cfg *config.Config, repo repository.Repository, client *midtrans.Client, webBuilder webbuilder.SiteManager, logger *slog.Logger) *PaymentService {
	return &PaymentService{
		cfg:        cfg,
		repo:       repo,
		client:     client,
		webBuilder: webBuilder,
		logger:     logger,
	}
}
