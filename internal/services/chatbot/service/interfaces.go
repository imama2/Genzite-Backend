package service

import (
	"context"

	"github.com/imama2/Genzite-Backend/internal/services/chatbot/models/dto"
)

// ChatManager is the interface exposed to the controller.
type ChatManager interface {
	GenerateDraft(ctx context.Context, userID uint, input dto.DraftInput) (*dto.SiteConfig, uint, string, string, error)
}

// AIClient is the interface for AI-based site config generation.
type AIClient interface {
	GenerateSiteConfig(ctx context.Context, input dto.BrandingInput) (dto.SiteConfig, error)
}
