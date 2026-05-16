package tool

// RegisterInput carries the data to register a new tool.
type RegisterInput struct {
	Name        string            `json:"name" validate:"required"`
	Description string            `json:"description" validate:"required"`
	Language    string            `json:"language" validate:"required,oneof=python go bash"`
	ObjectKey   string            `json:"object_key" validate:"required"`
	Bucket      string            `json:"bucket"`
	Parameters  map[string]any    `json:"parameters"`
	Env         map[string]string `json:"env"`
	Tags        []string          `json:"tags"`
	CreatedBy   string            `json:"created_by"`
	Metadata    map[string]any    `json:"metadata"`
}

// UpdateInput carries partial update data for a tool.
type UpdateInput struct {
	Description *string            `json:"description"`
	ObjectKey   *string            `json:"object_key"`
	Bucket      *string            `json:"bucket"`
	Parameters  map[string]any     `json:"parameters"`
	Env         map[string]string  `json:"env"`
	Tags        *[]string          `json:"tags"`
	IsActive    *bool              `json:"is_active"`
	Metadata    map[string]any    `json:"metadata"`
}

// SetEnvInput carries env vars to add/update on a tool.
// Existing keys are overwritten; keys not present are kept.
type SetEnvInput struct {
	Env map[string]string `json:"env" validate:"required"`
}

// SearchInput carries the semantic search query.
type SearchInput struct {
	Embedding []float64 `json:"embedding"`
	Query     string    `json:"query"`
	Tags      []string  `json:"tags"`
	UserID    string    `json:"user_id"`
	Limit     int       `json:"limit"`
}

// CreateWithPromptInput carries a natural language prompt to auto-generate a tool.
type CreateWithPromptInput struct {
	Prompt    string `json:"prompt" validate:"required"`
	CreatedBy string `json:"created_by"`
}

// TagDTO represents a tag in API responses (name + description).
type TagDTO struct {
	Name        string `json:"name"`
	Description string `json:"description"`
}

// TagOutput represents a tag in the GET /api/v1/tags response.
type TagOutput struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	Description string `json:"description"`
}

// ToolSearchResult is a tool matched by semantic search.
type ToolSearchResult struct {
	ID          string         `json:"id"`
	Name        string         `json:"name"`
	Description string         `json:"description"`
	Language    string         `json:"language"`
	Parameters  map[string]any `json:"parameters"`
	Score       float64        `json:"score"`
	Tags        []TagDTO       `json:"tags"`
}