package outbound

import (
	"context"

	"github.com/AndreeJait/zora-mcp-server/domain/entity"
)

// TagRepository defines the outbound port for tag persistence.
type TagRepository interface {
	// FindOrCreateByNames resolves tag names to Tag entities.
	// For names that don't exist yet, it creates them with the provided descriptions.
	// The descriptions map maps tag name -> description (used only for new tags).
	FindOrCreateByNames(ctx context.Context, names []string, descriptions map[string]string) ([]entity.Tag, error)

	// ListAll returns all tags ordered by name.
	ListAll(ctx context.Context) ([]entity.Tag, error)

	// FindByName returns a tag by its name, or nil if not found.
	FindByName(ctx context.Context, name string) (*entity.Tag, error)
}