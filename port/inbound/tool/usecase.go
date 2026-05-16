package tool

import (
	"context"

	"github.com/AndreeJait/zora-mcp-server/domain/entity"
)

// UseCase defines the inbound port for tool lifecycle management.
type UseCase interface {
	Register(ctx context.Context, input RegisterInput) (*entity.Tool, error)
	CreateWithPrompt(ctx context.Context, input CreateWithPromptInput) (*entity.Tool, error)
	Update(ctx context.Context, id string, input UpdateInput) (*entity.Tool, error)
	Delete(ctx context.Context, id string) error
	GetByID(ctx context.Context, id string) (*entity.Tool, error)
	List(ctx context.Context, page, perPage int) ([]entity.Tool, int64, error)
	SearchTools(ctx context.Context, input SearchInput) ([]ToolSearchResult, error)
	SetEnv(ctx context.Context, id string, input SetEnvInput) (*entity.Tool, error)
	DeleteEnv(ctx context.Context, id, key string) (*entity.Tool, error)
	ListTags(ctx context.Context) ([]TagOutput, error)
}