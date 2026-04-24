package service

import (
	"context"
	"crypto/rand"
	"crypto/sha512"
	"encoding/hex"
	"errors"
	"fmt"
	"log/slog"
	"strconv"
	"strings"
	"time"

	"github.com/imama2/Genzite-Backend/internal/core/config"
	"github.com/imama2/Genzite-Backend/internal/payment/midtrans"
	"github.com/imama2/Genzite-Backend/internal/payment/models"
	"github.com/imama2/Genzite-Backend/internal/payment/repository"
	webbuilder "github.com/imama2/Genzite-Backend/internal/webbuilder/service"
	"gorm.io/gorm"
)

const (
	providerMidtrans = "midtrans"
	statusPending    = "pending"
	statusPaid       = "paid"
	statusFailed     = "failed"
	statusCanceled   = "canceled"
	statusExpired    = "expired"
)

const paymentAmountIDR = int64(99000)

var (
	ErrInvalidInput      = errors.New("invalid input")
	ErrPaymentNotFound   = errors.New("payment not found")
	ErrInvalidSignature  = errors.New("invalid signature")
	ErrWebBuilderMissing = errors.New("web-builder service not configured")
)

type PaymentService struct {
	cfg        *config.Config
	repo       repository.Repository
	client     *midtrans.Client
	webBuilder webbuilder.SiteManager
	logger     *slog.Logger
}

type CreateResult struct {
	OrderID     string
	RedirectURL string
}

type WebhookPayload struct {
	OrderID           string `json:"order_id"`
	StatusCode        string `json:"status_code"`
	GrossAmount       string `json:"gross_amount"`
	SignatureKey      string `json:"signature_key"`
	TransactionStatus string `json:"transaction_status"`
	FraudStatus       string `json:"fraud_status"`
}

func New(cfg *config.Config, repo repository.Repository, client *midtrans.Client, webBuilder webbuilder.SiteManager, logger *slog.Logger) *PaymentService {
	return &PaymentService{
		cfg:        cfg,
		repo:       repo,
		client:     client,
		webBuilder: webBuilder,
		logger:     logger,
	}
}

func (s *PaymentService) CreateTransaction(ctx context.Context, userID uint, siteID uint) (*CreateResult, error) {
	if s.webBuilder == nil {
		return nil, ErrWebBuilderMissing
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
		Status:      statusPending,
		Provider:    providerMidtrans,
		RedirectURL: response.RedirectURL,
		SnapToken:   response.Token,
	}

	if err := s.repo.Create(ctx, payment); err != nil {
		return nil, err
	}

	return &CreateResult{
		OrderID:     orderID,
		RedirectURL: response.RedirectURL,
	}, nil
}

func (s *PaymentService) HandleWebhook(ctx context.Context, payload WebhookPayload) (string, error) {
	if payload.OrderID == "" || payload.SignatureKey == "" || payload.StatusCode == "" || payload.GrossAmount == "" {
		return "", ErrInvalidInput
	}

	if !s.verifySignature(payload) {
		return "", ErrInvalidSignature
	}

	payment, err := s.repo.GetByOrderID(ctx, payload.OrderID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return "", ErrPaymentNotFound
		}
		return "", err
	}

	newStatus := mapStatus(payload.TransactionStatus, payload.FraudStatus)
	updated := false

	if newStatus == statusPaid {
		if payment.Status != statusPaid {
			if s.webBuilder == nil {
				return "", ErrWebBuilderMissing
			}
			if _, err := s.webBuilder.PublishSite(ctx, payment.UserID, payment.SiteID); err != nil {
				return "", err
			}
			payment.Status = statusPaid
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
		return statusPaid
	case "capture":
		if strings.EqualFold(fraudStatus, "accept") || fraudStatus == "" {
			return statusPaid
		}
		return statusPending
	case "pending":
		return statusPending
	case "deny":
		return statusFailed
	case "expire":
		return statusExpired
	case "cancel":
		return statusCanceled
	default:
		return ""
	}
}

func (s *PaymentService) verifySignature(payload WebhookPayload) bool {
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
		return 0, ErrInvalidInput
	}
	if strings.Contains(trimmed, ".") {
		parts := strings.Split(trimmed, ".")
		trimmed = parts[0]
	}
	return strconv.ParseInt(trimmed, 10, 64)
}
