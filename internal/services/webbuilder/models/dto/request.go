package dto

import "github.com/imama2/Genzite-Backend/internal/services/webbuilder/models/entities"

type CreateSiteRequest struct {
	Slug   string     `json:"slug" binding:"required"`
	Config SiteConfig `json:"config" binding:"required"`
}

type CreateSiteInput struct {
	Slug   string
	Config SiteConfig
}

type SiteConfig struct {
	Title     string          `json:"title"`
	Name      string          `json:"name"`
	Headline  string          `json:"headline"`
	Bio       string          `json:"bio"`
	AvatarURL string          `json:"avatar_url"`
	Links     []entities.Link `json:"links"`
}
