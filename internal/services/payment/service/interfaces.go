package service

import (
	"context"

	"github.com/imama2/Genzite-Backend/internal/services/payment/models/dto"
)

type PaymentManager interface {
	CreateTransaction(ctx context.Context, userID uint, siteID uint) (*dto.CreateTransactionResult, error)
	HandleWebhook(ctx context.Context, payload dto.WebhookPayload) (string, error)
}
