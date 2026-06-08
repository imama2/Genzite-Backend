package service

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/imama2/Genzite-Backend/internal/services/webbuilder/models"
	"github.com/imama2/Genzite-Backend/internal/services/webbuilder/models/dto"
	"github.com/imama2/Genzite-Backend/internal/services/webbuilder/models/entities"
	errorUtils "github.com/imama2/Genzite-Backend/internal/utils/errors"
	"gorm.io/gorm"
)

func (s *SiteService) CreateSite(ctx context.Context, userID uint, input dto.CreateSiteInput) (*models.Site, error) {
	slug := sanitizeSlug(input.Slug)
	if slug == "" || strings.TrimSpace(input.Config.Name) == "" {
		return nil, errorUtils.ErrInvalidInput
	}

	site, err := s.repo.GetSiteBySlug(ctx, slug)
	if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, err
	}

	if site != nil && site.UserID != userID {
		return nil, errorUtils.ErrSlugTaken
	}

	configJSON, err := json.Marshal(input.Config)
	if err != nil {
		return nil, errorUtils.ErrInvalidInput
	}

	outputPath := s.outputPath(slug)
	if outputPath == "" {
		return nil, errorUtils.ErrInvalidInput
	}

	if err := s.renderToFile(outputPath, input.Config); err != nil {
		return nil, err
	}

	if site == nil {
		site = &models.Site{
			UserID:     userID,
			Slug:       slug,
			Title:      input.Config.Title,
			ConfigJSON: string(configJSON),
			Status:     entities.StatusDraft,
			OutputPath: outputPath,
		}
		if err := s.repo.CreateSite(ctx, site); err != nil {
			return nil, err
		}
	} else {
		site.Title = input.Config.Title
		site.ConfigJSON = string(configJSON)
		site.Status = entities.StatusDraft
		site.PublishedAt = nil
		site.OutputPath = outputPath
		if err := s.repo.UpdateSite(ctx, site); err != nil {
			return nil, err
		}
	}

	return site, nil
}

func (s *SiteService) PublishSite(ctx context.Context, userID uint, siteID uint) (*models.Site, error) {
	site, err := s.repo.GetSiteByID(ctx, siteID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errorUtils.ErrSiteNotFound
		}
		return nil, err
	}

	if site.UserID != userID {
		return nil, errorUtils.ErrUnauthorized
	}

	config, err := s.decodeConfig(site.ConfigJSON)
	if err != nil {
		return nil, errorUtils.ErrInvalidInput
	}

	if err := s.renderToFile(site.OutputPath, config); err != nil {
		return nil, err
	}

	now := time.Now()
	site.Status = entities.StatusPublished
	site.PublishedAt = &now

	if err := s.repo.UpdateSite(ctx, site); err != nil {
		return nil, err
	}

	return site, nil
}

func (s *SiteService) GetSiteForUser(ctx context.Context, userID uint, siteID uint) (*models.Site, error) {
	site, err := s.repo.GetSiteByID(ctx, siteID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errorUtils.ErrSiteNotFound
		}
		return nil, err
	}

	if site.UserID != userID {
		return nil, errorUtils.ErrUnauthorized
	}

	return site, nil
}

func (s *SiteService) GetPublishedSiteBySlug(ctx context.Context, slug string) (*models.Site, error) {
	site, err := s.repo.GetPublishedSiteBySlug(ctx, sanitizeSlug(slug))
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errorUtils.ErrSiteNotFound
		}
		return nil, err
	}
	return site, nil
}

func (s *SiteService) renderToFile(outputPath string, config dto.SiteConfig) error {
	if outputPath == "" {
		return errorUtils.ErrInvalidInput
	}

	dir := filepath.Dir(outputPath)
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return fmt.Errorf("create output dir: %w", err)
	}

	file, err := os.Create(outputPath)
	if err != nil {
		return fmt.Errorf("create output file: %w", err)
	}
	defer file.Close()

	data := entities.TemplateData{
		Title:     config.Title,
		Name:      config.Name,
		Headline:  config.Headline,
		Bio:       config.Bio,
		AvatarURL: config.AvatarURL,
		Links:     config.Links,
	}

	if err := s.template.Execute(file, data); err != nil {
		return errorUtils.ErrRenderFailed
	}

	return nil
}

func (s *SiteService) decodeConfig(raw string) (dto.SiteConfig, error) {
	var config dto.SiteConfig
	if err := json.Unmarshal([]byte(raw), &config); err != nil {
		return dto.SiteConfig{}, err
	}
	return config, nil
}

func (s *SiteService) outputPath(slug string) string {
	if slug == "" {
		return ""
	}

	root := strings.TrimSpace(s.cfg.WebBuilderPath)
	if root == "" {
		root = "storage/web-builder"
	}

	return filepath.Join(root, slug, "index.html")
}

func sanitizeSlug(value string) string {
	value = strings.ToLower(strings.TrimSpace(value))
	if value == "" {
		return ""
	}

	var builder strings.Builder
	builder.Grow(len(value))
	lastDash := false
	for _, r := range value {
		switch {
		case r >= 'a' && r <= 'z':
			builder.WriteRune(r)
			lastDash = false
		case r >= '0' && r <= '9':
			builder.WriteRune(r)
			lastDash = false
		case r == '-' || r == '_' || r == ' ':
			if !lastDash && builder.Len() > 0 {
				builder.WriteByte('-')
				lastDash = true
			}
		}
	}

	result := strings.Trim(builder.String(), "-")
	return result
}
