package dto

import "github.com/imama2/Genzite-Backend/internal/services/chatbot/models/entities"

// SiteConfig is the chatbot-local representation of a generated site configuration.
// Do not replace with webbuilder's SiteConfig — Rule 4 (no cross-service model imports).
type SiteConfig struct {
	Title     string          `json:"title"`
	Name      string          `json:"name"`
	Headline  string          `json:"headline"`
	Bio       string          `json:"bio"`
	AvatarURL string          `json:"avatar_url"`
	Links     []entities.Link `json:"links"`
}

// BrandingInput is the input to the AI client for site generation.
type BrandingInput struct {
	Prompt    string
	Name      string
	Headline  string
	Bio       string
	AvatarURL string
	Links     []entities.Link
}

// DraftInput is the input to the ChatService for generating a draft site.
type DraftInput struct {
	Slug      string
	Prompt    string
	Name      string
	Headline  string
	Bio       string
	AvatarURL string
	Links     []entities.Link
}

// GenerateDraftRequest is the HTTP request body for the GenerateDraft endpoint.
type GenerateDraftRequest struct {
	Slug      string          `json:"slug"    binding:"required"`
	Prompt    string          `json:"prompt"  binding:"required"`
	Name      string          `json:"name"`
	Headline  string          `json:"headline"`
	Bio       string          `json:"bio"`
	AvatarURL string          `json:"avatar_url"`
	Links     []entities.Link `json:"links"`
}
