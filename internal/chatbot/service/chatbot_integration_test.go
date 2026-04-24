package service_test

import (
	"context"
	"errors"
	"io"
	"log/slog"
	"os"
	"path/filepath"
	"testing"

	"github.com/imama2/Genzite-Backend/internal/chatbot/ai"
	"github.com/imama2/Genzite-Backend/internal/chatbot/service"
	"github.com/imama2/Genzite-Backend/internal/core/config"
	"github.com/imama2/Genzite-Backend/internal/webbuilder/models"
	webbuilder "github.com/imama2/Genzite-Backend/internal/webbuilder/service"
	"gorm.io/gorm"
)

type fakeAI struct {
	config webbuilder.SiteConfig
}

func (f *fakeAI) GenerateSiteConfig(_ context.Context, _ ai.BrandingInput) (webbuilder.SiteConfig, error) {
	return f.config, nil
}

type memoryRepo struct {
	nextID uint
	byID   map[uint]*models.Site
	bySlug map[string]*models.Site
}

func newMemoryRepo() *memoryRepo {
	return &memoryRepo{
		nextID: 1,
		byID:   make(map[uint]*models.Site),
		bySlug: make(map[string]*models.Site),
	}
}

func (r *memoryRepo) CreateSite(_ context.Context, site *models.Site) error {
	if site.ID == 0 {
		site.ID = r.nextID
		r.nextID++
	}
	r.byID[site.ID] = site
	r.bySlug[site.Slug] = site
	return nil
}

func (r *memoryRepo) UpdateSite(_ context.Context, site *models.Site) error {
	if _, ok := r.byID[site.ID]; !ok {
		return gorm.ErrRecordNotFound
	}
	r.byID[site.ID] = site
	r.bySlug[site.Slug] = site
	return nil
}

func (r *memoryRepo) GetSiteByID(_ context.Context, siteID uint) (*models.Site, error) {
	site, ok := r.byID[siteID]
	if !ok {
		return nil, gorm.ErrRecordNotFound
	}
	return site, nil
}

func (r *memoryRepo) GetSiteBySlug(_ context.Context, slug string) (*models.Site, error) {
	site, ok := r.bySlug[slug]
	if !ok {
		return nil, gorm.ErrRecordNotFound
	}
	return site, nil
}

func (r *memoryRepo) GetPublishedSiteBySlug(_ context.Context, slug string) (*models.Site, error) {
	site, ok := r.bySlug[slug]
	if !ok {
		return nil, gorm.ErrRecordNotFound
	}
	if site.Status != "published" {
		return nil, gorm.ErrRecordNotFound
	}
	return site, nil
}

func TestChatbotGeneratesDraftSite(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	outputDir := t.TempDir()
	cfg := &config.Config{WebBuilderPath: outputDir}
	logger := slog.New(slog.NewTextHandler(io.Discard, nil))

	repo := newMemoryRepo()
	siteService, err := webbuilder.NewSiteService(cfg, repo, logger)
	if err != nil {
		t.Fatalf("new site service: %v", err)
	}

	aiClient := &fakeAI{config: webbuilder.SiteConfig{
		Title:    "Jane Doe",
		Name:     "Jane Doe",
		Headline: "Product designer",
		Bio:      "Building thoughtful digital experiences.",
	}}

	chatService := service.New(aiClient, siteService)
	configResult, siteID, slug, status, err := chatService.GenerateDraft(ctx, 1, service.DraftInput{
		Slug:   "jane-doe",
		Prompt: "Minimal, friendly, modern portfolio.",
		Name:   "Jane Doe",
	})
	if err != nil {
		t.Fatalf("generate draft: %v", err)
	}

	if siteID == 0 || slug != "jane-doe" || status != "draft" {
		t.Fatalf("unexpected draft result: id=%d slug=%s status=%s", siteID, slug, status)
	}

	if configResult == nil || configResult.Name != "Jane Doe" {
		t.Fatalf("unexpected config result: %+v", configResult)
	}

	outputPath := filepath.Join(outputDir, "jane-doe", "index.html")
	if _, err := os.Stat(outputPath); err != nil {
		if errors.Is(err, os.ErrNotExist) {
			t.Fatalf("expected generated html at %s", outputPath)
		}
		t.Fatalf("stat generated html: %v", err)
	}
}
