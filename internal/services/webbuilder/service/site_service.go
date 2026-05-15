package service

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"html/template"
	"log/slog"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/imama2/Genzite-Backend/internal/core/config"
	"github.com/imama2/Genzite-Backend/internal/services/webbuilder/models"
	"github.com/imama2/Genzite-Backend/internal/services/webbuilder/repository"
	"gorm.io/gorm"
)

const (
	statusDraft     = "draft"
	statusPublished = "published"
)

var (
	ErrInvalidInput = errors.New("invalid input")
	ErrSlugTaken    = errors.New("slug already taken")
	ErrSiteNotFound = errors.New("site not found")
	ErrUnauthorized = errors.New("unauthorized")
	ErrRenderFailed = errors.New("render failed")
	ErrNotPublished = errors.New("site not published")
)

type SiteManager interface {
	CreateSite(ctx context.Context, userID uint, input CreateSiteInput) (*models.Site, error)
	PublishSite(ctx context.Context, userID uint, siteID uint) (*models.Site, error)
	GetSiteForUser(ctx context.Context, userID uint, siteID uint) (*models.Site, error)
}

type SiteService struct {
	cfg      *config.Config
	repo     repository.Repository
	logger   *slog.Logger
	template *template.Template
}

type CreateSiteInput struct {
	Slug   string
	Config SiteConfig
}

type SiteConfig struct {
	Title     string `json:"title"`
	Name      string `json:"name"`
	Headline  string `json:"headline"`
	Bio       string `json:"bio"`
	AvatarURL string `json:"avatar_url"`
	Links     []Link `json:"links"`
}

type Link struct {
	Label string `json:"label"`
	URL   string `json:"url"`
}

type templateData struct {
	Title     string
	Name      string
	Headline  string
	Bio       string
	AvatarURL string
	Links     []Link
}

func NewSiteService(cfg *config.Config, repo repository.Repository, logger *slog.Logger) (*SiteService, error) {
	tmpl, err := template.New("web-builder").Parse(defaultTemplate)
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

func (s *SiteService) CreateSite(ctx context.Context, userID uint, input CreateSiteInput) (*models.Site, error) {
	slug := sanitizeSlug(input.Slug)
	if slug == "" || strings.TrimSpace(input.Config.Name) == "" {
		return nil, ErrInvalidInput
	}

	site, err := s.repo.GetSiteBySlug(ctx, slug)
	if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, err
	}

	if site != nil && site.UserID != userID {
		return nil, ErrSlugTaken
	}

	configJSON, err := json.Marshal(input.Config)
	if err != nil {
		return nil, ErrInvalidInput
	}

	outputPath := s.outputPath(slug)
	if outputPath == "" {
		return nil, ErrInvalidInput
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
			Status:     statusDraft,
			OutputPath: outputPath,
		}
		if err := s.repo.CreateSite(ctx, site); err != nil {
			return nil, err
		}
	} else {
		site.Title = input.Config.Title
		site.ConfigJSON = string(configJSON)
		site.Status = statusDraft
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
			return nil, ErrSiteNotFound
		}
		return nil, err
	}

	if site.UserID != userID {
		return nil, ErrUnauthorized
	}

	config, err := s.decodeConfig(site.ConfigJSON)
	if err != nil {
		return nil, ErrInvalidInput
	}

	if err := s.renderToFile(site.OutputPath, config); err != nil {
		return nil, err
	}

	now := time.Now()
	site.Status = statusPublished
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
			return nil, ErrSiteNotFound
		}
		return nil, err
	}

	if site.UserID != userID {
		return nil, ErrUnauthorized
	}

	return site, nil
}

func (s *SiteService) GetPublishedSiteBySlug(ctx context.Context, slug string) (*models.Site, error) {
	site, err := s.repo.GetPublishedSiteBySlug(ctx, sanitizeSlug(slug))
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrSiteNotFound
		}
		return nil, err
	}
	return site, nil
}

func (s *SiteService) renderToFile(outputPath string, config SiteConfig) error {
	if outputPath == "" {
		return ErrInvalidInput
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

	data := templateData{
		Title:     config.Title,
		Name:      config.Name,
		Headline:  config.Headline,
		Bio:       config.Bio,
		AvatarURL: config.AvatarURL,
		Links:     config.Links,
	}

	if err := s.template.Execute(file, data); err != nil {
		return ErrRenderFailed
	}

	return nil
}

func (s *SiteService) decodeConfig(raw string) (SiteConfig, error) {
	var config SiteConfig
	if err := json.Unmarshal([]byte(raw), &config); err != nil {
		return SiteConfig{}, err
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

const defaultTemplate = `<!doctype html>
<html lang="en">
  <head>
    <meta charset="utf-8"/>
    <meta name="viewport" content="width=device-width, initial-scale=1"/>
    <title>{{ if .Title }}{{ .Title }}{{ else }}{{ .Name }}{{ end }}</title>
    <style>
      body { font-family: Arial, sans-serif; max-width: 720px; margin: 48px auto; padding: 0 16px; }
      header { display: flex; align-items: center; gap: 16px; }
      img.avatar { width: 96px; height: 96px; border-radius: 50%; object-fit: cover; }
      ul.links { list-style: none; padding: 0; }
      ul.links li { margin: 8px 0; }
    </style>
  </head>
  <body>
    <header>
      {{ if .AvatarURL }}<img class="avatar" src="{{ .AvatarURL }}" alt="{{ .Name }}"/>{{ end }}
      <div>
        <h1>{{ .Name }}</h1>
        {{ if .Headline }}<p>{{ .Headline }}</p>{{ end }}
      </div>
    </header>
    {{ if .Bio }}<p>{{ .Bio }}</p>{{ end }}
    {{ if .Links }}
      <h2>Links</h2>
      <ul class="links">
        {{ range .Links }}
          <li><a href="{{ .URL }}">{{ .Label }}</a></li>
        {{ end }}
      </ul>
    {{ end }}
  </body>
</html>
`

