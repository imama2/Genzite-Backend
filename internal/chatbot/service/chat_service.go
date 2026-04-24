package service

import (
	"context"
	"errors"
	"strings"

	"github.com/imama2/Genzite-Backend/internal/chatbot/ai"
	webbuilder "github.com/imama2/Genzite-Backend/internal/webbuilder/service"
)

var (
	ErrInvalidInput        = errors.New("invalid input")
	ErrAIUnavailable       = errors.New("ai client not configured")
	ErrWebBuilderMissing   = errors.New("web-builder service not configured")
)

type ChatService struct {
	ai         AIClient
	webBuilder webbuilder.SiteManager
}

type AIClient interface {
	GenerateSiteConfig(ctx context.Context, input ai.BrandingInput) (webbuilder.SiteConfig, error)
}

type DraftInput struct {
	Slug      string
	Prompt    string
	Name      string
	Headline  string
	Bio       string
	AvatarURL string
	Links     []webbuilder.Link
}

func New(aiClient AIClient, webBuilder webbuilder.SiteManager) *ChatService {
	return &ChatService{
		ai:         aiClient,
		webBuilder: webBuilder,
	}
}

func (s *ChatService) GenerateDraft(ctx context.Context, userID uint, input DraftInput) (*webbuilder.SiteConfig, uint, string, string, error) {
	if s.ai == nil {
		return nil, 0, "", "", ErrAIUnavailable
	}
	if s.webBuilder == nil {
		return nil, 0, "", "", ErrWebBuilderMissing
	}
	if strings.TrimSpace(input.Slug) == "" || strings.TrimSpace(input.Prompt) == "" {
		return nil, 0, "", "", ErrInvalidInput
	}

	config, err := s.ai.GenerateSiteConfig(ctx, ai.BrandingInput{
		Prompt:    input.Prompt,
		Name:      input.Name,
		Headline:  input.Headline,
		Bio:       input.Bio,
		AvatarURL: input.AvatarURL,
		Links:     input.Links,
	})
	if err != nil {
		return nil, 0, "", "", err
	}

	applyOverrides(&config, input)
	if strings.TrimSpace(config.Name) == "" {
		return nil, 0, "", "", ErrInvalidInput
	}

	site, err := s.webBuilder.CreateSite(ctx, userID, webbuilder.CreateSiteInput{
		Slug:   input.Slug,
		Config: config,
	})
	if err != nil {
		return nil, 0, "", "", err
	}

	return &config, site.ID, site.Slug, site.Status, nil
}

func applyOverrides(config *webbuilder.SiteConfig, input DraftInput) {
	if strings.TrimSpace(input.Name) != "" {
		config.Name = input.Name
	}
	if strings.TrimSpace(input.Headline) != "" {
		config.Headline = input.Headline
	}
	if strings.TrimSpace(input.Bio) != "" {
		config.Bio = input.Bio
	}
	if strings.TrimSpace(input.AvatarURL) != "" {
		config.AvatarURL = input.AvatarURL
	}
	if len(input.Links) > 0 {
		config.Links = input.Links
	}
	if strings.TrimSpace(input.Prompt) != "" && strings.TrimSpace(config.Title) == "" {
		if strings.TrimSpace(input.Name) != "" {
			config.Title = strings.TrimSpace(input.Name)
		}
	}
}
