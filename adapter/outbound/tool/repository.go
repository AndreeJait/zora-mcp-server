package tool

import (
	"context"
	"fmt"

	domainEntity "github.com/AndreeJait/zora-mcp-server/domain/entity"
	portOutbound "github.com/AndreeJait/zora-mcp-server/port/outbound"
	"github.com/pgvector/pgvector-go"
	"gorm.io/gorm"
)

// Repository implements portOutbound.ToolRepository using GORM + pgvector.
type Repository struct {
	db *gorm.DB
}

var _ portOutbound.ToolRepository = (*Repository)(nil)

func NewRepository(db *gorm.DB) *Repository {
	return &Repository{db: db}
}

func (r *Repository) Create(ctx context.Context, tool *domainEntity.Tool) error {
	return r.db.WithContext(ctx).Create(tool).Error
}

func (r *Repository) Update(ctx context.Context, tool *domainEntity.Tool) error {
	return r.db.WithContext(ctx).Save(tool).Error
}

func (r *Repository) Delete(ctx context.Context, id string) error {
	return r.db.WithContext(ctx).Where("id = ?", id).Delete(&domainEntity.Tool{}).Error
}

func (r *Repository) GetByID(ctx context.Context, id string) (*domainEntity.Tool, error) {
	var tool domainEntity.Tool
	if err := r.db.WithContext(ctx).Where("id = ?", id).First(&tool).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil
		}
		return nil, err
	}
	if err := r.LoadToolTags(ctx, []domainEntity.Tool{tool}); err != nil {
		return nil, err
	}
	return &tool, nil
}

func (r *Repository) GetByName(ctx context.Context, name string) (*domainEntity.Tool, error) {
	var tool domainEntity.Tool
	if err := r.db.WithContext(ctx).Where("name = ?", name).First(&tool).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil
		}
		return nil, err
	}
	if err := r.LoadToolTags(ctx, []domainEntity.Tool{tool}); err != nil {
		return nil, err
	}
	return &tool, nil
}

func (r *Repository) List(ctx context.Context, page, perPage int) ([]domainEntity.Tool, int64, error) {
	var tools []domainEntity.Tool
	var total int64

	offset := (page - 1) * perPage

	if err := r.db.WithContext(ctx).Model(&domainEntity.Tool{}).Count(&total).Error; err != nil {
		return nil, 0, err
	}

	if err := r.db.WithContext(ctx).
		Order("created_at DESC").
		Offset(offset).
		Limit(perPage).
		Find(&tools).Error; err != nil {
		return nil, 0, err
	}

	if err := r.LoadToolTags(ctx, tools); err != nil {
		return nil, 0, err
	}

	return tools, total, nil
}

// toolWithScore holds a tool row plus its cosine similarity score.
type toolWithScore struct {
	domainEntity.Tool
	Score float64 `gorm:"column:score"`
}

// SearchByEmbedding uses pgvector cosine similarity to find relevant tools.
// Returns tools with their similarity scores (0 to 1, higher is more similar).
// Supports optional filtering by tags and user_id.
func (r *Repository) SearchByEmbedding(ctx context.Context, embedding []float64, limit int, filter portOutbound.ToolSearchFilter) ([]domainEntity.Tool, error) {
	if limit <= 0 {
		limit = 15
	}

	vec32 := make([]float32, len(embedding))
	for i, v := range embedding {
		vec32[i] = float32(v)
	}
	vec := pgvector.NewVector(vec32)

	query := r.db.WithContext(ctx).
		Model(&domainEntity.Tool{}).
		Select(fmt.Sprintf("*, 1 - (embedding <=> '%s') AS score", vec.String())).
		Where("is_active = ?", true).
		Where("embedding IS NOT NULL")

	// Filter by tags: only return tools whose tags overlap with the provided tag names
	if len(filter.Tags) > 0 {
		subQuery := r.db.Table("zora_tool_tags tt").
			Select("DISTINCT tt.tool_id").
			Joins("JOIN zora_tags t ON t.id = tt.tag_id").
			Where("t.name IN ?", filter.Tags)
		query = query.Where("zora_tools_registry.id IN (?)", subQuery)
	}

	// Filter by user_id: return tools owned by this user or global (created_by is empty)
	if filter.UserID != "" {
		query = query.Where("created_by = ? OR created_by = '' OR created_by IS NULL", filter.UserID)
	}

	var results []toolWithScore
	err := query.Order("score DESC").Limit(limit).Find(&results).Error
	if err != nil {
		return nil, err
	}

	tools := make([]domainEntity.Tool, len(results))
	for i, r := range results {
		tools[i] = r.Tool
		tools[i].Score = r.Score
	}

	if err := r.LoadToolTags(ctx, tools); err != nil {
		return nil, err
	}

	return tools, nil
}

func (r *Repository) SaveExecution(ctx context.Context, exec *domainEntity.ToolExecution) error {
	return r.db.WithContext(ctx).Create(exec).Error
}

// ReplaceToolTags replaces all tag associations for a tool.
func (r *Repository) ReplaceToolTags(ctx context.Context, toolID string, tagIDs []string) error {
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		// Delete existing associations
		if err := tx.Where("tool_id = ?", toolID).Delete(&domainEntity.ToolTag{}).Error; err != nil {
			return err
		}
		// Insert new associations
		if len(tagIDs) > 0 {
			associations := make([]domainEntity.ToolTag, 0, len(tagIDs))
			for _, tagID := range tagIDs {
				associations = append(associations, domainEntity.ToolTag{
					ToolID: toolID,
					TagID:  tagID,
				})
			}
			if err := tx.Create(&associations).Error; err != nil {
				return err
			}
		}
		return nil
	})
}

// LoadToolTags populates the Tags field for the given tools by querying
// zora_tool_tags JOIN zora_tags.
func (r *Repository) LoadToolTags(ctx context.Context, tools []domainEntity.Tool) error {
	if len(tools) == 0 {
		return nil
	}

	toolIDs := make([]string, len(tools))
	toolMap := make(map[string]*domainEntity.Tool, len(tools))
	for i := range tools {
		toolIDs[i] = tools[i].ID
		toolMap[tools[i].ID] = &tools[i]
		tools[i].Tags = nil // reset
	}

	type toolTagRow struct {
		ToolID      string
		TagID       string
		TagName     string
		TagDesc     string
		TagCreatedAt string
	}

	var rows []toolTagRow
	if err := r.db.WithContext(ctx).
		Table("zora_tool_tags tt").
		Select("tt.tool_id, t.id as tag_id, t.name as tag_name, t.description as tag_desc, t.created_at as tag_created_at").
		Joins("JOIN zora_tags t ON t.id = tt.tag_id").
		Where("tt.tool_id IN ?", toolIDs).
		Find(&rows).Error; err != nil {
		return err
	}

	for _, row := range rows {
		if tool, ok := toolMap[row.ToolID]; ok {
			tool.Tags = append(tool.Tags, domainEntity.Tag{
				ID:          row.TagID,
				Name:        row.TagName,
				Description: row.TagDesc,
			})
		}
	}

	return nil
}