package ai

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"backend/internal/config"

	"go.uber.org/fx"
	"go.uber.org/zap"
)

var Module = fx.Module("ai", fx.Provide(NewClient))

// Client wraps OpenAI-compatible API
type Client struct {
	apiKey  string
	baseURL string
	model   string
	http    *http.Client
	logger  *zap.Logger
	enabled bool
}

// ClientParams for fx
type ClientParams struct {
	fx.In
	Config config.Config
	Logger *zap.Logger `optional:"true"`
}

func NewClient(params ClientParams) *Client {
	logger := params.Logger
	if logger == nil {
		logger = zap.NewNop()
	}

	enabled := params.Config.AI.APIKey != ""
	baseURL := params.Config.AI.BaseURL
	if baseURL == "" {
		baseURL = "https://api.deepseek.com"
	}
	model := params.Config.AI.Model
	if model == "" {
		model = "deepseek-chat"
	}

	return &Client{
		apiKey:  params.Config.AI.APIKey,
		baseURL: baseURL,
		model:   model,
		enabled: enabled,
		http: &http.Client{
			Timeout: 30 * time.Second,
		},
		logger: logger.Named("ai"),
	}
}

// IsEnabled returns true if API key is configured
func (c *Client) IsEnabled() bool {
	return c.enabled
}

// ChatCompletionRequest for OpenAI API
type ChatCompletionRequest struct {
	Model    string          `json:"model"`
	Messages []ChatMessage   `json:"messages"`
	MaxTokens int            `json:"max_tokens,omitempty"`
	Temperature float64      `json:"temperature,omitempty"`
}

// ChatMessage for OpenAI API
type ChatMessage struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

// ChatCompletionResponse from OpenAI API
type ChatCompletionResponse struct {
	Choices []struct {
		Message struct {
			Content string `json:"content"`
		} `json:"message"`
	} `json:"choices"`
	Error *struct {
		Message string `json:"message"`
	} `json:"error,omitempty"`
}

// Analyze sends prompt to LLM and returns response
func (c *Client) Analyze(ctx context.Context, systemPrompt, userPrompt string) (string, error) {
	if !c.enabled {
		return "", fmt.Errorf("AI not configured: set AI_API_KEY environment variable")
	}

	reqBody := ChatCompletionRequest{
		Model: c.model,
		Messages: []ChatMessage{
			{Role: "system", Content: systemPrompt},
			{Role: "user", Content: userPrompt},
		},
		MaxTokens:  1000,
		Temperature: 0.7,
	}

	body, err := json.Marshal(reqBody)
	if err != nil {
		return "", fmt.Errorf("marshal request: %w", err)
	}

	url := fmt.Sprintf("%s/chat/completions", c.baseURL)
	req, err := http.NewRequestWithContext(ctx, "POST", url, bytes.NewReader(body))
	if err != nil {
		return "", fmt.Errorf("create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", c.apiKey))

	resp, err := c.http.Do(req)
	if err != nil {
		return "", fmt.Errorf("send request: %w", err)
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("read response: %w", err)
	}

	var result ChatCompletionResponse
	if err := json.Unmarshal(respBody, &result); err != nil {
		return "", fmt.Errorf("parse response: %w", err)
	}

	if result.Error != nil {
		return "", fmt.Errorf("API error: %s", result.Error.Message)
	}

	if len(result.Choices) == 0 {
		return "", fmt.Errorf("no choices in response")
	}

	return result.Choices[0].Message.Content, nil
}
