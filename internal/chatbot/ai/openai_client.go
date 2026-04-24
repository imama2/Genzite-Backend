package ai

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"strings"
	"time"

	"github.com/imama2/Genzite-Backend/internal/core/config"
	webbuilder "github.com/imama2/Genzite-Backend/internal/webbuilder/service"
)

type OpenAIClient struct {
	apiKey     string
	model      string
	baseURL    string
	httpClient *http.Client
	logger     *slog.Logger
}

func NewOpenAIClient(cfg *config.Config, logger *slog.Logger) (*OpenAIClient, error) {
	if cfg.OpenAIAPIKey == "" {
		return nil, errors.New("OPENAI_API_KEY is required")
	}

	baseURL := strings.TrimRight(cfg.OpenAIBaseURL, "/")
	if baseURL == "" {
		baseURL = "https://api.openai.com/v1"
	}

	model := strings.TrimSpace(cfg.OpenAIModel)
	if model == "" {
		model = "gpt-4o-mini"
	}

	return &OpenAIClient{
		apiKey:     cfg.OpenAIAPIKey,
		model:      model,
		baseURL:    baseURL,
		httpClient: &http.Client{Timeout: 20 * time.Second},
		logger:     logger,
	}, nil
}

type chatRequest struct {
	Model          string        `json:"model"`
	Messages       []chatMessage `json:"messages"`
	Temperature    float64       `json:"temperature,omitempty"`
	ResponseFormat *format       `json:"response_format,omitempty"`
}

type chatMessage struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

type format struct {
	Type string `json:"type"`
}

type chatResponse struct {
	Choices []struct {
		Message struct {
			Content string `json:"content"`
		} `json:"message"`
	} `json:"choices"`
	Error *struct {
		Message string `json:"message"`
	} `json:"error,omitempty"`
}

type BrandingInput struct {
	Prompt    string
	Name      string
	Headline  string
	Bio       string
	AvatarURL string
	Links     []webbuilder.Link
}

func (c *OpenAIClient) GenerateSiteConfig(ctx context.Context, input BrandingInput) (webbuilder.SiteConfig, error) {
	systemPrompt := `You are a branding assistant. Produce a JSON object for a personal website.
Return only JSON with keys: "title", "name", "headline", "bio", "avatar_url", "links".
Each link must be an object with "label" and "url". Use empty arrays if needed.`

	userPrompt := buildUserPrompt(input)

	payload := chatRequest{
		Model: c.model,
		Messages: []chatMessage{
			{Role: "system", Content: systemPrompt},
			{Role: "user", Content: userPrompt},
		},
		Temperature:    0.7,
		ResponseFormat: &format{Type: "json_object"},
	}

	body, err := json.Marshal(payload)
	if err != nil {
		return webbuilder.SiteConfig{}, fmt.Errorf("marshal request: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, c.baseURL+"/chat/completions", bytes.NewReader(body))
	if err != nil {
		return webbuilder.SiteConfig{}, fmt.Errorf("build request: %w", err)
	}
	req.Header.Set("Authorization", "Bearer "+c.apiKey)
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return webbuilder.SiteConfig{}, fmt.Errorf("openai request failed: %w", err)
	}
	defer resp.Body.Close()

	responseBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return webbuilder.SiteConfig{}, fmt.Errorf("read response: %w", err)
	}

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return webbuilder.SiteConfig{}, fmt.Errorf("openai error: %s", strings.TrimSpace(string(responseBody)))
	}

	var parsed chatResponse
	if err := json.Unmarshal(responseBody, &parsed); err != nil {
		return webbuilder.SiteConfig{}, fmt.Errorf("parse response: %w", err)
	}

	if parsed.Error != nil {
		return webbuilder.SiteConfig{}, fmt.Errorf("openai error: %s", parsed.Error.Message)
	}

	if len(parsed.Choices) == 0 {
		return webbuilder.SiteConfig{}, errors.New("openai response empty")
	}

	content := strings.TrimSpace(parsed.Choices[0].Message.Content)
	if content == "" {
		return webbuilder.SiteConfig{}, errors.New("openai response empty")
	}

	var config webbuilder.SiteConfig
	if err := json.Unmarshal([]byte(content), &config); err != nil {
		c.logger.Error("failed to parse openai json", "error", err, "content", content)
		return webbuilder.SiteConfig{}, errors.New("invalid openai response")
	}

	return config, nil
}

func buildUserPrompt(input BrandingInput) string {
	var builder strings.Builder
	builder.WriteString("Branding brief:\n")
	builder.WriteString(input.Prompt)

	if strings.TrimSpace(input.Name) != "" {
		builder.WriteString("\nName: ")
		builder.WriteString(input.Name)
	}
	if strings.TrimSpace(input.Headline) != "" {
		builder.WriteString("\nHeadline: ")
		builder.WriteString(input.Headline)
	}
	if strings.TrimSpace(input.Bio) != "" {
		builder.WriteString("\nBio: ")
		builder.WriteString(input.Bio)
	}
	if strings.TrimSpace(input.AvatarURL) != "" {
		builder.WriteString("\nAvatar URL: ")
		builder.WriteString(input.AvatarURL)
	}
	if len(input.Links) > 0 {
		builder.WriteString("\nLinks:")
		for _, link := range input.Links {
			builder.WriteString("\n- ")
			builder.WriteString(link.Label)
			builder.WriteString(": ")
			builder.WriteString(link.URL)
		}
	}

	return builder.String()
}
