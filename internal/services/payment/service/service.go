package service

import (
	"context"
	"crypto/rand"
	"crypto/sha512"
	"encoding/hex"
	"errors"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/imama2/Genzite-Backend/internal/services/payment/gateway/midtrans"
	"github.com/imama2/Genzite-Backend/internal/services/payment/models"
	"github.com/imama2/Genzite-Backend/internal/services/payment/models/dto"
	"github.com/imama2/Genzite-Backend/internal/services/payment/models/entities"
	errorUtils "github.com/imama2/Genzite-Backend/internal/utils/errors"
	"gorm.io/gorm"
)

const paymentAmountIDR = int64(99000)

func (s *PaymentService) CreateTransaction(ctx context.Context, userID uint, siteID uint) (*dto.CreateTransactionResult, error) {
	if s.webBuilder == nil {
		return nil, errorUtils.ErrWebBuilderMissing
	}

	_, err := s.webBuilder.GetSiteForUser(ctx, userID, siteID)
	if err != nil {
		return nil, err
	}

	orderID := buildOrderID(userID, siteID)
	request := midtrans.TransactionRequest{
		TransactionDetails: midtrans.TransactionDetails{
			OrderID:     orderID,
			GrossAmount: paymentAmountIDR,
		},
		ItemDetails: []midtrans.ItemDetail{
			{
				ID:       "web-builder",
				Price:    paymentAmountIDR,
				Quantity: 1,
				Name:     "Web Builder",
			},
		},
	}

	response, err := s.client.CreateTransaction(ctx, request)
	if err != nil {
		return nil, err
	}

	payment := &models.Payment{
		UserID:      userID,
		SiteID:      siteID,
		OrderID:     orderID,
		Amount:      paymentAmountIDR,
		Status:      entities.StatusPending,
		Provider:    entities.ProviderMidtrans,
		RedirectURL: response.RedirectURL,
		SnapToken:   response.Token,
	}

	if err := s.repo.Create(ctx, payment); err != nil {
		return nil, err
	}

	return &dto.CreateTransactionResult{
		OrderID:     orderID,
		RedirectURL: response.RedirectURL,
	}, nil
}

func (s *PaymentService) HandleWebhook(ctx context.Context, payload dto.WebhookPayload) (string, error) {
	if payload.OrderID == "" || payload.SignatureKey == "" || payload.StatusCode == "" || payload.GrossAmount == "" {
		return "", errorUtils.ErrInvalidInput
	}

	if !s.verifySignature(payload) {
		return "", errorUtils.ErrInvalidSignature
	}

	payment, err := s.repo.GetByOrderID(ctx, payload.OrderID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return "", errorUtils.ErrPaymentNotFound
		}
		return "", err
	}

	newStatus := mapStatus(payload.TransactionStatus, payload.FraudStatus)
	updated := false

	if newStatus == entities.StatusPaid {
		if payment.Status != entities.StatusPaid {
			if s.webBuilder == nil {
				return "", errorUtils.ErrWebBuilderMissing
			}
			if _, err := s.webBuilder.PublishSite(ctx, payment.UserID, payment.SiteID); err != nil {
				return "", err
			}
			payment.Status = entities.StatusPaid
			updated = true
		}
	} else if newStatus != "" && payment.Status != newStatus {
		payment.Status = newStatus
		updated = true
	}

	if amount, err := parseAmount(payload.GrossAmount); err == nil && amount > 0 && payment.Amount != amount {
		payment.Amount = amount
		updated = true
	}

	if updated {
		if err := s.repo.Update(ctx, payment); err != nil {
			return "", err
		}
	}

	return payment.Status, nil
}

func mapStatus(transactionStatus, fraudStatus string) string {
	switch transactionStatus {
	case "settlement":
		return entities.StatusPaid
	case "capture":
		if strings.EqualFold(fraudStatus, "accept") || fraudStatus == "" {
			return entities.StatusPaid
		}
		return entities.StatusPending
	case "pending":
		return entities.StatusPending
	case "deny":
		return entities.StatusFailed
	case "expire":
		return entities.StatusExpired
	case "cancel":
		return entities.StatusCanceled
	default:
		return ""
	}
}

func (s *PaymentService) verifySignature(payload dto.WebhookPayload) bool {
	raw := payload.OrderID + payload.StatusCode + payload.GrossAmount + s.cfg.MidtransServerKey
	hash := sha512.Sum512([]byte(raw))
	expected := hex.EncodeToString(hash[:])
	return strings.EqualFold(expected, payload.SignatureKey)
}

func buildOrderID(userID uint, siteID uint) string {
	nonce := make([]byte, 4)
	if _, err := rand.Read(nonce); err != nil {
		return fmt.Sprintf("wb-%d-%d-%s", userID, siteID, time.Now().UTC().Format("20060102150405"))
	}
	return fmt.Sprintf("wb-%d-%d-%s-%s", userID, siteID, time.Now().UTC().Format("20060102150405"), hex.EncodeToString(nonce))
}

func parseAmount(value string) (int64, error) {
	trimmed := strings.TrimSpace(value)
	if trimmed == "" {
		return 0, errorUtils.ErrInvalidInput
	}
	if strings.Contains(trimmed, ".") {
		parts := strings.Split(trimmed, ".")
		trimmed = parts[0]
	}
	return strconv.ParseInt(trimmed, 10, 64)
}
