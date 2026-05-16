package outbound

import "context"

// EmbeddingProvider generates vector embeddings from text.
type EmbeddingProvider interface {
	Embed(ctx context.Context, text string) ([]float64, error)
}