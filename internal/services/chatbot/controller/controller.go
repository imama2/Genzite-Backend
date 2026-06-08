package controller

import (
	"errors"

	"github.com/gin-gonic/gin"
	"github.com/imama2/Genzite-Backend/internal/core/middleware"
	"github.com/imama2/Genzite-Backend/internal/services/chatbot/models/dto"
	"github.com/imama2/Genzite-Backend/internal/services/chatbot/service"
	errorUtils "github.com/imama2/Genzite-Backend/internal/utils/errors"
	"github.com/imama2/Genzite-Backend/internal/utils/responses"
)

type ChatController struct {
	service *service.ChatService
}

func New(svc *service.ChatService) *ChatController {
	return &ChatController{service: svc}
}

func (h *ChatController) GenerateDraft(c *gin.Context) {
	userID, ok := getUserID(c)
	if !ok {
		responses.Unauthorized(c, "unauthorized")
		return
	}

	var req dto.GenerateDraftRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		responses.BadRequest(c, "invalid request")
		return
	}

	config, siteID, slug, status, err := h.service.GenerateDraft(c.Request.Context(), userID, dto.DraftInput{
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
		case errors.Is(err, errorUtils.ErrInvalidInput):
			responses.BadRequest(c, "invalid input")
		case errors.Is(err, errorUtils.ErrAIUnavailable):
			responses.ServiceUnavailableResponse(c, "ai client not configured")
		case errors.Is(err, errorUtils.ErrWebBuilderMissing):
			responses.ServiceUnavailableResponse(c, "web-builder service not configured")
		default:
			responses.ServerError(c, "failed to generate site")
		}
		return
	}

	responses.Ok(c, "draft generated", dto.DraftResponse{
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
