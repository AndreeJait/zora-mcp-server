package entity

import (
	"time"

	"github.com/pgvector/pgvector-go"
)

// Tool represents a registered tool script stored in MinIO.
type Tool struct {
	ID          string         `json:"id" gorm:"primaryKey;type:uuid;default:gen_random_uuid()"`
	Name        string         `json:"name" gorm:"uniqueIndex;not null"`
	Description string         `json:"description" gorm:"not null"`
	Language    string         `json:"language" gorm:"not null;check:language IN ('python','go','bash')"`
	ObjectKey   string         `json:"object_key" gorm:"not null"`
	Bucket      string         `json:"bucket" gorm:"default:'zora-scripts'"`
	Parameters  map[string]any    `json:"parameters" gorm:"serializer:json;default:'{}'"`
	Env         map[string]string `json:"env" gorm:"serializer:json;default:'{}'"`
	Tags        []Tag           `json:"tags" gorm:"-"` // populated by repository, not persisted directly
	Version     int               `json:"version" gorm:"default:1"`
	IsActive    bool           `json:"is_active" gorm:"default:true"`
	CreatedBy   string         `json:"created_by"`
	CreatedAt   time.Time      `json:"created_at" gorm:"autoCreateTime"`
	UpdatedAt   time.Time      `json:"updated_at" gorm:"autoUpdateTime"`
	Metadata    map[string]any `json:"metadata" gorm:"serializer:json;default:'{}'"`
	Embedding   pgvector.Vector `json:"-" gorm:"type:vector(768)"`
	Score       float64         `json:"score,omitempty" gorm:"-"` // populated by SearchByEmbedding, not persisted
}

func (Tool) TableName() string {
	return "zora_tools_registry"
}

// ToolExecution records the outcome of a tool invocation.
type ToolExecution struct {
	ID         int64          `json:"id" gorm:"primaryKey;autoIncrement"`
	ToolID     string         `json:"tool_id" gorm:"not null;index"`
	TraceID    string         `json:"trace_id" gorm:"not null;index"`
	Parameters map[string]any `json:"parameters" gorm:"serializer:json"`
	Result     map[string]any `json:"result" gorm:"serializer:json"`
	Status     string         `json:"status" gorm:"not null;check:status IN ('success','failed','timeout')"`
	ErrorMsg   string         `json:"error_msg"`
	DurationMs int            `json:"duration_ms"`
	CreatedAt  time.Time      `json:"created_at" gorm:"autoCreateTime"`
}

func (ToolExecution) TableName() string {
	return "zora_tool_executions"
}
