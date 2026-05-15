package handler

import (
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/imama2/Genzite-Backend/internal/core/middleware"
	"github.com/imama2/Genzite-Backend/internal/services/payment/service"
	webbuilder "github.com/imama2/Genzite-Backend/internal/services/webbuilder/service"
)

type PaymentHandler struct {
	service *service.PaymentService
}

func New(service *service.PaymentService) *PaymentHandler {
	return &PaymentHandler{service: service}
}

type createRequest struct {
	SiteID uint `json:"site_id" binding:"required"`
}

type createResponse struct {
	OrderID     string `json:"order_id"`
	RedirectURL string `json:"redirect_url"`
}

func (h *PaymentHandler) CreateTransaction(c *gin.Context) {
	userID, ok := getUserID(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

	var req createRequest
	if err := c.ShouldBindJSON(&req); err != nil || req.SiteID == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request"})
		return
	}

	result, err := h.service.CreateTransaction(c.Request.Context(), userID, req.SiteID)
	if err != nil {
		switch {
		case errors.Is(err, service.ErrInvalidInput):
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid input"})
		case errors.Is(err, webbuilder.ErrSiteNotFound):
			c.JSON(http.StatusNotFound, gin.H{"error": "site not found"})
		case errors.Is(err, webbuilder.ErrUnauthorized):
			c.JSON(http.StatusForbidden, gin.H{"error": "forbidden"})
		case errors.Is(err, service.ErrWebBuilderMissing):
			c.JSON(http.StatusServiceUnavailable, gin.H{"error": "web-builder service not configured"})
		default:
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to create payment"})
		}
		return
	}

	c.JSON(http.StatusOK, createResponse{
		OrderID:     result.OrderID,
		RedirectURL: result.RedirectURL,
	})
}

func (h *PaymentHandler) Webhook(c *gin.Context) {
	var payload service.WebhookPayload
	if err := c.ShouldBindJSON(&payload); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request"})
		return
	}

	status, err := h.service.HandleWebhook(c.Request.Context(), payload)
	if err != nil {
		switch {
		case errors.Is(err, service.ErrInvalidSignature):
			c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid signature"})
		case errors.Is(err, service.ErrPaymentNotFound):
			c.JSON(http.StatusNotFound, gin.H{"error": "payment not found"})
		case errors.Is(err, service.ErrInvalidInput):
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid input"})
		default:
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to process webhook"})
		}
		return
	}

	c.JSON(http.StatusOK, gin.H{"status": status})
}

func getUserID(c *gin.Context) (uint, bool) {
	value, ok := c.Get(middleware.ContextUserIDKey)
	if !ok {
		return 0, false
	}
	userID, ok := value.(uint)
	return userID, ok
}


