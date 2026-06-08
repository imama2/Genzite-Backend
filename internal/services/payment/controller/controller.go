package controller

import (
	"errors"

	"github.com/gin-gonic/gin"
	"github.com/imama2/Genzite-Backend/internal/core/middleware"
	"github.com/imama2/Genzite-Backend/internal/services/payment/models/dto"
	"github.com/imama2/Genzite-Backend/internal/services/payment/service"
	errorUtils "github.com/imama2/Genzite-Backend/internal/utils/errors"
	"github.com/imama2/Genzite-Backend/internal/utils/responses"
)

type PaymentController struct {
	service *service.PaymentService
}

func New(svc *service.PaymentService) *PaymentController {
	return &PaymentController{service: svc}
}

func (h *PaymentController) CreateTransaction(c *gin.Context) {
	userID, ok := getUserID(c)
	if !ok {
		responses.Unauthorized(c, "unauthorized")
		return
	}

	var req dto.CreateTransactionRequest
	if err := c.ShouldBindJSON(&req); err != nil || req.SiteID == 0 {
		responses.BadRequest(c, "invalid request")
		return
	}

	result, err := h.service.CreateTransaction(c.Request.Context(), userID, req.SiteID)
	if err != nil {
		switch {
		case errors.Is(err, errorUtils.ErrInvalidInput):
			responses.BadRequest(c, "invalid input")
		case errors.Is(err, errorUtils.ErrSiteNotFound):
			responses.NotFound(c, "site not found")
		case errors.Is(err, errorUtils.ErrUnauthorized):
			responses.ForbiddenResponse(c, "forbidden")
		case errors.Is(err, errorUtils.ErrWebBuilderMissing):
			responses.ServiceUnavailableResponse(c, "web-builder service not configured")
		default:
			responses.ServerError(c, "failed to create payment")
		}
		return
	}

	responses.Ok(c, "payment created", dto.CreateTransactionResponse{
		OrderID:     result.OrderID,
		RedirectURL: result.RedirectURL,
	})
}

func (h *PaymentController) Webhook(c *gin.Context) {
	var payload dto.WebhookPayload
	if err := c.ShouldBindJSON(&payload); err != nil {
		responses.BadRequest(c, "invalid request")
		return
	}

	status, err := h.service.HandleWebhook(c.Request.Context(), payload)
	if err != nil {
		switch {
		case errors.Is(err, errorUtils.ErrInvalidSignature):
			responses.Unauthorized(c, "invalid signature")
		case errors.Is(err, errorUtils.ErrPaymentNotFound):
			responses.NotFound(c, "payment not found")
		case errors.Is(err, errorUtils.ErrInvalidInput):
			responses.BadRequest(c, "invalid input")
		default:
			responses.ServerError(c, "failed to process webhook")
		}
		return
	}

	responses.Ok(c, "webhook processed", gin.H{"status": status})
}

func getUserID(c *gin.Context) (uint, bool) {
	value, ok := c.Get(middleware.ContextUserIDKey)
	if !ok {
		return 0, false
	}
	userID, ok := value.(uint)
	return userID, ok
}
