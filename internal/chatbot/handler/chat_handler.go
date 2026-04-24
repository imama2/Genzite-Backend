package handler

import (
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/imama2/Genzite-Backend/internal/chatbot/service"
	"github.com/imama2/Genzite-Backend/internal/core/middleware"
	webbuilder "github.com/imama2/Genzite-Backend/internal/webbuilder/service"
)

type ChatHandler struct {
	service *service.ChatService
}

func New(service *service.ChatService) *ChatHandler {
	return &ChatHandler{service: service}
}

type chatRequest struct {
	Slug      string              `json:"slug" binding:"required"`
	Prompt    string              `json:"prompt" binding:"required"`
	Name      string              `json:"name"`
	Headline  string              `json:"headline"`
	Bio       string              `json:"bio"`
	AvatarURL string              `json:"avatar_url"`
	Links     []webbuilder.Link   `json:"links"`
}

type chatResponse struct {
	SiteID uint                   `json:"site_id"`
	Slug   string                 `json:"slug"`
	Status string                 `json:"status"`
	Config webbuilder.SiteConfig  `json:"config"`
}

func (h *ChatHandler) GenerateDraft(c *gin.Context) {
	userID, ok := getUserID(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

	var req chatRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request"})
		return
	}

	config, siteID, slug, status, err := h.service.GenerateDraft(c.Request.Context(), userID, service.DraftInput{
		Slug:      req.Slug,
		Prompt:    req.Prompt,
		Name:      req.Name,
		Headline:  req.Headline,
		Bio:       req.Bio,
		AvatarURL: req.AvatarURL,
		Links:     req.Links,
	})
	if err != nil {
		switch {
		case errors.Is(err, service.ErrInvalidInput):
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid input"})
		case errors.Is(err, service.ErrAIUnavailable):
			c.JSON(http.StatusServiceUnavailable, gin.H{"error": "ai client not configured"})
		case errors.Is(err, service.ErrWebBuilderMissing):
			c.JSON(http.StatusServiceUnavailable, gin.H{"error": "web-builder service not configured"})
		default:
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to generate site"})
		}
		return
	}

	c.JSON(http.StatusOK, chatResponse{
		SiteID: siteID,
		Slug:   slug,
		Status: status,
		Config: *config,
	})
}

func getUserID(c *gin.Context) (uint, bool) {
	value, ok := c.Get(middleware.ContextUserIDKey)
	if !ok {
		return 0, false
	}
	userID, ok := value.(uint)
	return userID, ok
}
