package mcp

import (
	"context"

	"github.com/AndreeJait/zora-mcp-server/domain/entity"
)

// UseCase defines the inbound port for MCP protocol operations.
type UseCase interface {
	// ListTools returns all active tools as MCP tool definitions.
	ListTools(ctx context.Context) ([]ToolDefinition, error)

	// CallTool executes a tool by name with the given arguments.
	CallTool(ctx context.Context, name string, arguments map[string]any) (*entity.ToolExecution, error)
}

// ToolDefinition is the MCP-compatible tool representation.
type ToolDefinition struct {
	Name        string         `json:"name"`
	Description string         `json:"description"`
	InputSchema map[string]any `json:"inputSchema"`
}
