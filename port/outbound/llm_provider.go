package outbound

import "context"

// LLMProvider generates text completions from prompts.
type LLMProvider interface {
	Generate(ctx context.Context, systemPrompt string, userPrompt string) (string, error)
}