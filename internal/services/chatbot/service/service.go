package service

import (
	"context"
	"strings"

	"github.com/imama2/Genzite-Backend/internal/services/chatbot/models/dto"
	"github.com/imama2/Genzite-Backend/internal/services/chatbot/models/entities"
	webbuilder_dto "github.com/imama2/Genzite-Backend/internal/services/webbuilder/models/dto"
	webbuilder_entities "github.com/imama2/Genzite-Backend/internal/services/webbuilder/models/entities"
	errorUtils "github.com/imama2/Genzite-Backend/internal/utils/errors"
)

func (s *ChatService) GenerateDraft(ctx context.Context, userID uint, input dto.DraftInput) (*dto.SiteConfig, uint, string, string, error) {
	if s.ai == nil {
		return nil, 0, "", "", errorUtils.ErrAIUnavailable
	}
	if s.webBuilder == nil {
		return nil, 0, "", "", errorUtils.ErrWebBuilderMissing
	}
	if strings.TrimSpace(input.Slug) == "" || strings.TrimSpace(input.Prompt) == "" {
		return nil, 0, "", "", errorUtils.ErrInvalidInput
	}

	config, err := s.ai.GenerateSiteConfig(ctx, dto.BrandingInput{
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
		return nil, 0, "", "", errorUtils.ErrInvalidInput
	}

	// Map chatbot's local SiteConfig to webbuilder's DTO at the service boundary.
	site, err := s.webBuilder.CreateSite(ctx, userID, webbuilder_dto.CreateSiteInput{
		Slug: input.Slug,
		Config: webbuilder_dto.SiteConfig{
			Title:     config.Title,
			Name:      config.Name,
			Headline:  config.Headline,
			Bio:       config.Bio,
			AvatarURL: config.AvatarURL,
			Links:     toWebbuilderLinks(config.Links),
		},
	})
	if err != nil {
		return nil, 0, "", "", err
	}

	return &config, site.ID, site.Slug, site.Status, nil
}

func applyOverrides(config *dto.SiteConfig, input dto.DraftInput) {
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

// toWebbuilderLinks converts chatbot-local Link slice to webbuilder's entity type.
func toWebbuilderLinks(links []entities.Link) []webbuilder_entities.Link {
	out := make([]webbuilder_entities.Link, len(links))
	for i, l := range links {
		out[i] = webbuilder_entities.Link{Label: l.Label, URL: l.URL}
	}
	return out
}
