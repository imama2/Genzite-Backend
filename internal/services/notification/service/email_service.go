package service

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"net/smtp"
	"strconv"
	"strings"
	"time"

	"github.com/imama2/Genzite-Backend/internal/core/broker"
	"github.com/imama2/Genzite-Backend/internal/core/config"
)

var (
	ErrInvalidInput  = errors.New("invalid input")
	ErrNotConfigured = errors.New("notification service not configured")
)

type EmailService struct {
	cfg    *config.Config
	broker *broker.Client
	logger *slog.Logger
}

type EmailRequest struct {
	To      string `json:"to"`
	Subject string `json:"subject"`
	Body    string `json:"body"`
}

func New(cfg *config.Config, brokerClient *broker.Client, logger *slog.Logger) *EmailService {
	return &EmailService{
		cfg:    cfg,
		broker: brokerClient,
		logger: logger,
	}
}

func (s *EmailService) SendEmail(ctx context.Context, req EmailRequest) error {
	if s.broker == nil {
		return ErrNotConfigured
	}
	if strings.TrimSpace(req.To) == "" || strings.TrimSpace(req.Subject) == "" {
		return ErrInvalidInput
	}

	body, err := json.Marshal(req)
	if err != nil {
		return ErrInvalidInput
	}

	return s.broker.Publish(ctx, body)
}

func (s *EmailService) StartConsumer(ctx context.Context) error {
	if s.broker == nil {
		return ErrNotConfigured
	}

	return s.broker.Consume(ctx, func(ctx context.Context, payload []byte) error {
		var req EmailRequest
		if err := json.Unmarshal(payload, &req); err != nil {
			return err
		}

		if err := s.sendSMTP(ctx, req); err != nil {
			s.logger.Error("failed to send email", "error", err, "to", req.To)
			return err
		}

		return nil
	})
}

func (s *EmailService) sendSMTP(_ context.Context, req EmailRequest) error {
	host := strings.TrimSpace(s.cfg.SMTPHost)
	port := strings.TrimSpace(s.cfg.SMTPPort)
	from := strings.TrimSpace(s.cfg.SMTPFrom)
	if host == "" || port == "" || from == "" {
		return ErrNotConfigured
	}

	portNumber, err := strconv.Atoi(port)
	if err != nil || portNumber <= 0 {
		return fmt.Errorf("invalid SMTP_PORT")
	}

	addr := fmt.Sprintf("%s:%d", host, portNumber)
	message := buildMessage(s.cfg, req)

	var auth smtp.Auth
	if s.cfg.SMTPUser != "" {
		auth = smtp.PlainAuth("", s.cfg.SMTPUser, s.cfg.SMTPPassword, host)
	}

	return smtp.SendMail(addr, auth, from, []string{req.To}, []byte(message))
}

func buildMessage(cfg *config.Config, req EmailRequest) string {
	fromName := strings.TrimSpace(cfg.SMTPFromName)
	from := cfg.SMTPFrom
	if fromName != "" {
		from = fmt.Sprintf("%s <%s>", fromName, cfg.SMTPFrom)
	}

	headers := map[string]string{
		"From":         from,
		"To":           req.To,
		"Subject":      req.Subject,
		"Date":         time.Now().UTC().Format(time.RFC1123Z),
		"MIME-Version": "1.0",
		"Content-Type": "text/plain; charset=UTF-8",
	}

	var builder strings.Builder
	for key, value := range headers {
		builder.WriteString(fmt.Sprintf("%s: %s\r\n", key, value))
	}
	builder.WriteString("\r\n")
	builder.WriteString(req.Body)

	return builder.String()
}
