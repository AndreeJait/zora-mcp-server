package outbound

import (
	"context"

	"github.com/AndreeJait/zora-mcp-server/domain/entity"
)

// ToolSearchFilter carries optional filters for semantic tool search.
type ToolSearchFilter struct {
	Tags   []string // if provided, only return tools whose tags overlap
	UserID string   // if provided, only return tools owned by this user or global (created_by is empty)
}

// ToolRepository defines the outbound port for tool persistence.
type ToolRepository interface {
	Create(ctx context.Context, tool *entity.Tool) error
	Update(ctx context.Context, tool *entity.Tool) error
	Delete(ctx context.Context, id string) error
	GetByID(ctx context.Context, id string) (*entity.Tool, error)
	GetByName(ctx context.Context, name string) (*entity.Tool, error)
	List(ctx context.Context, page, perPage int) ([]entity.Tool, int64, error)
	SearchByEmbedding(ctx context.Context, embedding []float64, limit int, filter ToolSearchFilter) ([]entity.Tool, error)
	SaveExecution(ctx context.Context, exec *entity.ToolExecution) error
	ReplaceToolTags(ctx context.Context, toolID string, tagIDs []string) error
	LoadToolTags(ctx context.Context, tools []entity.Tool) error
}
