package tag

import (
	"context"

	domainEntity "github.com/AndreeJait/zora-mcp-server/domain/entity"
	portOutbound "github.com/AndreeJait/zora-mcp-server/port/outbound"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

// Repository implements portOutbound.TagRepository using GORM.
type Repository struct {
	db *gorm.DB
}

var _ portOutbound.TagRepository = (*Repository)(nil)

func NewRepository(db *gorm.DB) *Repository {
	return &Repository{db: db}
}

func (r *Repository) FindOrCreateByNames(ctx context.Context, names []string, descriptions map[string]string) ([]domainEntity.Tag, error) {
	if len(names) == 0 {
		return nil, nil
	}

	// Build tag entities for upsert
	tags := make([]domainEntity.Tag, 0, len(names))
	for _, name := range names {
		desc := ""
		if d, ok := descriptions[name]; ok {
			desc = d
		}
		tags = append(tags, domainEntity.Tag{
			Name:        name,
			Description: desc,
		})
	}

	// Upsert: create tags that don't exist, ignore conflicts on name
	if err := r.db.WithContext(ctx).Clauses(clause.OnConflict{
		Columns:   []clause.Column{{Name: "name"}},
		DoNothing: true,
	}).Create(&tags).Error; err != nil {
		return nil, err
	}

	// Now fetch all tags by name (both existing and newly created)
	var result []domainEntity.Tag
	if err := r.db.WithContext(ctx).Where("name IN ?", names).Order("name").Find(&result).Error; err != nil {
		return nil, err
	}

	return result, nil
}

func (r *Repository) ListAll(ctx context.Context) ([]domainEntity.Tag, error) {
	var tags []domainEntity.Tag
	if err := r.db.WithContext(ctx).Order("name").Find(&tags).Error; err != nil {
		return nil, err
	}
	return tags, nil
}

func (r *Repository) FindByName(ctx context.Context, name string) (*domainEntity.Tag, error) {
	var tag domainEntity.Tag
	if err := r.db.WithContext(ctx).Where("name = ?", name).First(&tag).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil
		}
		return nil, err
	}
	return &tag, nil
}