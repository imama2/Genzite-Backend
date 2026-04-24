package midtrans

import (
	"bytes"
	"context"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/imama2/Genzite-Backend/internal/core/config"
)

type Client struct {
	serverKey string
	baseURL   string
	http      *http.Client
}

func New(cfg *config.Config) (*Client, error) {
	if strings.TrimSpace(cfg.MidtransServerKey) == "" {
		return nil, errors.New("MIDTRANS_SERVER_KEY is required")
	}

	baseURL := strings.TrimRight(cfg.MidtransBaseURL, "/")
	if baseURL == "" {
		baseURL = "https://app.sandbox.midtrans.com"
	}

	return &Client{
		serverKey: cfg.MidtransServerKey,
		baseURL:   baseURL,
		http:      &http.Client{Timeout: 15 * time.Second},
	}, nil
}

type TransactionRequest struct {
	TransactionDetails TransactionDetails `json:"transaction_details"`
	ItemDetails        []ItemDetail        `json:"item_details,omitempty"`
	CustomerDetails    *CustomerDetails    `json:"customer_details,omitempty"`
}

type TransactionDetails struct {
	OrderID     string `json:"order_id"`
	GrossAmount int64  `json:"gross_amount"`
}

type ItemDetail struct {
	ID       string `json:"id"`
	Price    int64  `json:"price"`
	Quantity int    `json:"quantity"`
	Name     string `json:"name"`
}

type CustomerDetails struct {
	Email string `json:"email,omitempty"`
	Name  string `json:"first_name,omitempty"`
}

type TransactionResponse struct {
	Token       string `json:"token"`
	RedirectURL string `json:"redirect_url"`
}

func (c *Client) CreateTransaction(ctx context.Context, request TransactionRequest) (*TransactionResponse, error) {
	body, err := json.Marshal(request)
	if err != nil {
		return nil, fmt.Errorf("marshal request: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, c.baseURL+"/snap/v1/transactions", bytes.NewReader(body))
	if err != nil {
		return nil, fmt.Errorf("build request: %w", err)
	}

	req.Header.Set("Authorization", "Basic "+basicAuth(c.serverKey))
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.http.Do(req)
	if err != nil {
		return nil, fmt.Errorf("midtrans request failed: %w", err)
	}
	defer resp.Body.Close()

	responseBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("read response: %w", err)
	}

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return nil, fmt.Errorf("midtrans error: %s", strings.TrimSpace(string(responseBody)))
	}

	var parsed TransactionResponse
	if err := json.Unmarshal(responseBody, &parsed); err != nil {
		return nil, fmt.Errorf("parse response: %w", err)
	}

	if parsed.RedirectURL == "" || parsed.Token == "" {
		return nil, errors.New("midtrans response missing token")
	}

	return &parsed, nil
}

func basicAuth(serverKey string) string {
	return base64.StdEncoding.EncodeToString([]byte(serverKey + ":"))
}
