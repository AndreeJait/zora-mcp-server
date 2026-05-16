package echo

import (
	"encoding/json"
	"net/http"

	"github.com/AndreeJait/go-utility/v2/httpw/echow"
	"github.com/AndreeJait/go-utility/v2/responsew"
	"github.com/AndreeJait/zora-mcp-server/port/inbound/mcp"
	"github.com/labstack/echo/v5"
)

// RegisterMCPRoutes registers MCP protocol endpoints.
func RegisterMCPRoutes(r RouteRegistrar, mcpUC mcp.UseCase) {
	r.GET("/api/v1/mcp/tools", echow.Bind(listMCPTools(mcpUC)))
	r.POST("/api/v1/mcp/tools/call", echow.Bind(callMCPTool(mcpUC)))
}

// @Summary      List available MCP tools
// @Description  Retrieve all active tools in MCP tool-definition format
// @Tags         mcp
// @Produce      json
// @Success      200  {array}  mcp.ToolDefinition
// @Router       /api/v1/mcp/tools [get]
func listMCPTools(mcpUC mcp.UseCase) func(c *echo.Context) (any, error) {
	return func(c *echo.Context) (any, error) {
		tools, err := mcpUC.ListTools(c.Request().Context())
		if err != nil {
			return nil, err
		}
		return responsew.Success(tools, "MCP tools listed"), nil
	}
}

type callToolRequest struct {
	Name      string         `json:"name"`
	Arguments map[string]any `json:"arguments"`
}

// @Summary      Execute an MCP tool
// @Description  Call a registered tool by name with the given arguments
// @Tags         mcp
// @Accept       json
// @Produce      json
// @Param        body  body  callToolRequest  true  "Tool call request"
// @Success      200   {object}  responsew.BaseResponse
// @Failure      404   {object}  responsew.BaseResponse
// @Failure      500   {object}  responsew.BaseResponse
// @Router       /api/v1/mcp/tools/call [post]
func callMCPTool(mcpUC mcp.UseCase) func(c *echo.Context) (any, error) {
	return func(c *echo.Context) (any, error) {
		var req callToolRequest
		if err := json.NewDecoder((*c).Request().Body).Decode(&req); err != nil {
			return nil, err
		}

		result, err := mcpUC.CallTool(c.Request().Context(), req.Name, req.Arguments)
		if err != nil {
			return nil, err
		}

		data := map[string]any{
			"content":  "",
			"is_error": result.Status != "success",
		}
		if result.Result != nil {
			if out, ok := result.Result["stdout"]; ok {
				data["content"] = out
			} else if out, ok := result.Result["output"]; ok {
				data["content"] = out
			} else {
				resultJSON, _ := json.Marshal(result.Result)
				data["content"] = string(resultJSON)
			}
			if sf, ok := result.Result["saved_files"]; ok {
				data["saved_files"] = sf
			}
		}

		return responsew.Success(data, "Tool executed"), nil
	}
}

// Ensure unused imports for swagger
var _ = http.StatusOK
