package outbound

import (
	"context"
)

// ExecutionResult holds the output of a script execution.
type ExecutionResult struct {
	Stdout     string `json:"stdout"`
	Stderr     string `json:"stderr"`
	ExitCode   int    `json:"exit_code"`
	DurationMs int    `json:"duration_ms"`
	TimedOut   bool   `json:"timed_out"`
}

// ScriptExecutor defines the outbound port for executing scripts.
type ScriptExecutor interface {
	// Execute runs a script with the given language, content, arguments, and environment variables.
	Execute(ctx context.Context, language string, script []byte, args map[string]any, env map[string]string) (*ExecutionResult, error)
}
