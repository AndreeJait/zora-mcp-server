package entity

import "time"

// Tag represents a normalized tag with a description for tool classification.
type Tag struct {
	ID          string    `json:"id" gorm:"primaryKey;type:uuid;default:gen_random_uuid()"`
	Name        string    `json:"name" gorm:"uniqueIndex;not null"`
	Description string    `json:"description" gorm:"not null;default:''"`
	CreatedAt   time.Time `json:"created_at" gorm:"autoCreateTime"`
}

func (Tag) TableName() string {
	return "zora_tags"
}

// ToolTag is the join entity linking tools to tags.
type ToolTag struct {
	ToolID string `gorm:"primaryKey;type:uuid"`
	TagID  string `gorm:"primaryKey;type:uuid"`
}

func (ToolTag) TableName() string {
	return "zora_tool_tags"
}