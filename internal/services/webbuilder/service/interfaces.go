package service

import (
	"context"

	"github.com/imama2/Genzite-Backend/internal/services/webbuilder/models"
	"github.com/imama2/Genzite-Backend/internal/services/webbuilder/models/dto"
)

type SiteManager interface {
	CreateSite(ctx context.Context, userID uint, input dto.CreateSiteInput) (*models.Site, error)
	PublishSite(ctx context.Context, userID uint, siteID uint) (*models.Site, error)
	GetSiteForUser(ctx context.Context, userID uint, siteID uint) (*models.Site, error)
	GetPublishedSiteBySlug(ctx context.Context, slug string) (*models.Site, error)
}
