package service

import (
	"fmt"
	"html/template"
	"log/slog"

	"github.com/imama2/Genzite-Backend/internal/core/config"
	"github.com/imama2/Genzite-Backend/internal/services/webbuilder/models/entities"
	"github.com/imama2/Genzite-Backend/internal/services/webbuilder/repository"
)

type SiteService struct {
	cfg      *config.Config
	repo     repository.Repository
	logger   *slog.Logger
	template *template.Template
}

func NewSiteService(cfg *config.Config, repo repository.Repository, logger *slog.Logger) (*SiteService, error) {
	tmpl, err := template.New("web-builder").Parse(entities.DefaultTemplate)
	if err != nil {
		return nil, fmt.Errorf("parse template: %w", err)
	}

	return &SiteService{
		cfg:      cfg,
		repo:     repo,
		logger:   logger,
		template: tmpl,
	}, nil
}
