package llm

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/AndreeJait/zora-mcp-server/config"
	portOutbound "github.com/AndreeJait/zora-mcp-server/port/outbound"
)

type ollamaProvider struct {
	baseURL    string
	model      string
	apiKey     string
	httpClient *http.Client
}

var _ portOutbound.LLMProvider = (*ollamaProvider)(nil)

func NewOllamaProvider(cfg *config.AppConfig) portOutbound.LLMProvider {
	return &ollamaProvider{
		baseURL: cfg.LLM.BaseURL,
		model:   cfg.LLM.Model,
		apiKey:  cfg.LLM.APIKey,
		httpClient: &http.Client{
			Timeout: 120 * time.Second,
		},
	}
}

type chatRequest struct {
	Model       string           `json:"model"`
	Messages    []chatMessage    `json:"messages"`
	Temperature float64          `json:"temperature"`
	MaxTokens   int              `json:"max_tokens"`
}

type chatMessage struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

type chatResponse struct {
	Choices []struct {
		Message struct {
			Content string `json:"content"`
		} `json:"message"`
	} `json:"choices"`
}

func (p *ollamaProvider) Generate(ctx context.Context, systemPrompt string, userPrompt string) (string, error) {
	messages := []chatMessage{
		{Role: "system", Content: systemPrompt},
		{Role: "user", Content: userPrompt},
	}

	body, err := json.Marshal(chatRequest{
		Model:       p.model,
		Messages:    messages,
		Temperature: 0.2,
		MaxTokens:   4096,
	})
	if err != nil {
		return "", fmt.Errorf("marshal chat request: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, p.baseURL+"/chat/completions", bytes.NewReader(body))
	if err != nil {
		return "", fmt.Errorf("create chat request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")
	if p.apiKey != "" {
		req.Header.Set("Authorization", "Bearer "+p.apiKey)
	}

	resp, err := p.httpClient.Do(req)
	if err != nil {
		return "", fmt.Errorf("call llm api: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		respBody, _ := io.ReadAll(resp.Body)
		return "", fmt.Errorf("llm api returned status %d: %s", resp.StatusCode, string(respBody))
	}

	var result chatResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return "", fmt.Errorf("decode chat response: %w", err)
	}

	if len(result.Choices) == 0 {
		return "", fmt.Errorf("llm api returned no choices")
	}

	return result.Choices[0].Message.Content, nil
}