package service

import (
	"context"
	"errors"

	"github.com/imama2/Genzite-Backend/internal/services/template/repository"
)

type Service struct {
	repo repository.Repository
}

func New(repo repository.Repository) *Service {
	return &Service{repo: repo}
}

func (s *Service) Ping(ctx context.Context) error {
	if s.repo == nil {
		return errors.New("repository not configured")
	}
	return s.repo.Ping(ctx)
}

