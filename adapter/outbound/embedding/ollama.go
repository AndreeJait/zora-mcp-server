package embedding

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
	httpClient *http.Client
}

var _ portOutbound.EmbeddingProvider = (*ollamaProvider)(nil)

func NewOllamaProvider(cfg *config.AppConfig) portOutbound.EmbeddingProvider {
	return &ollamaProvider{
		baseURL: cfg.Embedding.BaseURL,
		model:   cfg.Embedding.Model,
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
}

type embedRequest struct {
	Model string `json:"model"`
	Input string `json:"input"`
}

type embedResponse struct {
	Model      string      `json:"model"`
	Embeddings [][]float64 `json:"embeddings"`
}

func (p *ollamaProvider) Embed(ctx context.Context, text string) ([]float64, error) {
	body, err := json.Marshal(embedRequest{
		Model: p.model,
		Input: text,
	})
	if err != nil {
		return nil, fmt.Errorf("marshal embed request: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, p.baseURL+"/api/embed", bytes.NewReader(body))
	if err != nil {
		return nil, fmt.Errorf("create embed request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := p.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("call ollama embed api: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		respBody, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("ollama embed api returned status %d: %s", resp.StatusCode, string(respBody))
	}

	var result embedResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("decode embed response: %w", err)
	}

	if len(result.Embeddings) == 0 {
		return nil, fmt.Errorf("ollama embed api returned no embeddings")
	}

	return result.Embeddings[0], nil
}