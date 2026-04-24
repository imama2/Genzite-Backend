package repository

import (
	"context"

	"github.com/imama2/Genzite-Backend/internal/webbuilder/models"
	"gorm.io/gorm"
)

type Repository interface {
	CreateSite(ctx context.Context, site *models.Site) error
	UpdateSite(ctx context.Context, site *models.Site) error
	GetSiteByID(ctx context.Context, siteID uint) (*models.Site, error)
	GetSiteBySlug(ctx context.Context, slug string) (*models.Site, error)
	GetPublishedSiteBySlug(ctx context.Context, slug string) (*models.Site, error)
}

type GormRepository struct {
	db *gorm.DB
}

func New(db *gorm.DB) *GormRepository {
	return &GormRepository{db: db}
}

func (r *GormRepository) CreateSite(ctx context.Context, site *models.Site) error {
	return r.db.WithContext(ctx).Create(site).Error
}

func (r *GormRepository) UpdateSite(ctx context.Context, site *models.Site) error {
	return r.db.WithContext(ctx).Save(site).Error
}

func (r *GormRepository) GetSiteByID(ctx context.Context, siteID uint) (*models.Site, error) {
	var site models.Site
	if err := r.db.WithContext(ctx).First(&site, siteID).Error; err != nil {
		return nil, err
	}
	return &site, nil
}

func (r *GormRepository) GetSiteBySlug(ctx context.Context, slug string) (*models.Site, error) {
	var site models.Site
	if err := r.db.WithContext(ctx).Where("slug = ?", slug).First(&site).Error; err != nil {
		return nil, err
	}
	return &site, nil
}

func (r *GormRepository) GetPublishedSiteBySlug(ctx context.Context, slug string) (*models.Site, error) {
	var site models.Site
	if err := r.db.WithContext(ctx).
		Where("slug = ? AND status = ?", slug, "published").
		First(&site).Error; err != nil {
		return nil, err
	}
	return &site, nil
}
