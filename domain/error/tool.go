package error

import "github.com/AndreeJait/go-utility/v2/statusw"

var (
	ErrToolNotFound       = statusw.NotFound.WithCustomMessage("tool not found")
	ErrToolNameExists     = statusw.Conflict.WithCustomMessage("tool with this name already exists")
	ErrInvalidLanguage    = statusw.InvalidReqParam.WithCustomMessage("language must be python, go, or bash")
	ErrScriptFetchFailed  = statusw.InternalServerError.WithCustomMessage("failed to fetch script from storage")
	ErrExecutionTimeout   = statusw.InternalServerError.WithCustomMessage("tool execution timed out")
	ErrExecutionFailed    = statusw.InternalServerError.WithCustomMessage("tool execution failed")
	ErrInvalidParameters  = statusw.InvalidReqParam.WithCustomMessage("invalid tool parameters")
	ErrInvalidEnvKey      = statusw.InvalidReqParam.WithCustomMessage("environment variable keys must start with MCP_")
)
