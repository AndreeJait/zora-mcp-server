package echo

import (
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/AndreeJait/go-utility/v2/httpw/echow"
	"github.com/AndreeJait/go-utility/v2/responsew"
	"github.com/AndreeJait/go-utility/v2/statusw"
	"github.com/AndreeJait/zora-mcp-server/domain/entity"
	"github.com/AndreeJait/zora-mcp-server/port/inbound/tool"
	"github.com/labstack/echo/v5"
)

// swagger type hint — ensures swag can resolve entity.Tool
var _ entity.Tool

// RegisterToolRoutes registers REST routes for tool management.
func RegisterToolRoutes(r RouteRegistrar, toolUC tool.UseCase) {
	r.POST("/api/v1/tools", echow.Bind(registerTool(toolUC)))
	r.POST("/api/v1/tools/create-with-prompt", echow.Bind(createWithPrompt(toolUC)))
	r.GET("/api/v1/tools/:id", echow.Bind(getTool(toolUC)))
	r.PUT("/api/v1/tools/:id", echow.Bind(updateTool(toolUC)))
	r.DELETE("/api/v1/tools/:id", echow.Bind(deleteTool(toolUC)))
	r.GET("/api/v1/tools", echow.Bind(listTools(toolUC)))
	r.POST("/api/v1/tools/search", echow.Bind(searchTools(toolUC)))
	r.GET("/api/v1/tags", echow.Bind(listTags(toolUC)))

	// Env management
	r.PUT("/api/v1/tools/:id/env", echow.Bind(setToolEnv(toolUC)))
	r.DELETE("/api/v1/tools/:id/env/:key", echow.Bind(deleteToolEnv(toolUC)))
}

// @Summary      Register a new tool
// @Description  Create a new tool entry with script metadata and configuration
// @Tags         tools
// @Accept       json
// @Produce      json
// @Param        body  body  tool.RegisterInput  true  "Tool registration data"
// @Success      201  {object}  entity.Tool
// @Failure      400  {object}  responsew.BaseResponse
// @Failure      409  {object}  responsew.BaseResponse
// @Security     ApiKeyAuth
// @Router       /api/v1/tools [post]
func createWithPrompt(toolUC tool.UseCase) func(c *echo.Context) (any, error) {
	return func(c *echo.Context) (any, error) {
		var input tool.CreateWithPromptInput
		if err := json.NewDecoder((*c).Request().Body).Decode(&input); err != nil {
			return nil, err
		}
		if input.Prompt == "" {
			return nil, statusw.InvalidReqParam.WithCustomMessage("prompt is required")
		}
		result, err := toolUC.CreateWithPrompt(c.Request().Context(), input)
		if err != nil {
			return nil, err
		}
		return responsew.Success(result, "Tool created from prompt"), nil
	}
}

// @Summary      Register a new tool
// @Description  Create a new tool entry with script metadata and configuration
// @Tags         tools
// @Accept       json
// @Produce      json
// @Param        body  body  tool.RegisterInput  true  "Tool registration data"
// @Success      201  {object}  entity.Tool
// @Failure      400  {object}  responsew.BaseResponse
// @Failure      409  {object}  responsew.BaseResponse
// @Security     ApiKeyAuth
// @Router       /api/v1/tools [post]
func registerTool(toolUC tool.UseCase) func(c *echo.Context) (any, error) {
	return func(c *echo.Context) (any, error) {
		var input tool.RegisterInput
		if err := json.NewDecoder((*c).Request().Body).Decode(&input); err != nil {
			return nil, err
		}
		result, err := toolUC.Register(c.Request().Context(), input)
		if err != nil {
			return nil, err
		}
		return responsew.Success(result, "Tool registered"), nil
	}
}

// @Summary      Get a tool by ID
// @Description  Retrieve a single tool by its UUID
// @Tags         tools
// @Produce      json
// @Param        id  path  string  true  "Tool ID"
// @Success      200  {object}  entity.Tool
// @Failure      404  {object}  responsew.BaseResponse
// @Security     ApiKeyAuth
// @Router       /api/v1/tools/{id} [get]
func getTool(toolUC tool.UseCase) func(c *echo.Context) (any, error) {
	return func(c *echo.Context) (any, error) {
		id := c.Param("id")
		result, err := toolUC.GetByID(c.Request().Context(), id)
		if err != nil {
			return nil, err
		}
		return responsew.Success(result, "Tool found"), nil
	}
}

// @Summary      Update a tool
// @Description  Partially update a tool by ID
// @Tags         tools
// @Accept       json
// @Produce      json
// @Param        id    path  string             true  "Tool ID"
// @Param        body  body  tool.UpdateInput   true  "Tool update data"
// @Success      200   {object}  entity.Tool
// @Failure      400   {object}  responsew.BaseResponse
// @Failure      404   {object}  responsew.BaseResponse
// @Security     ApiKeyAuth
// @Router       /api/v1/tools/{id} [put]
func updateTool(toolUC tool.UseCase) func(c *echo.Context) (any, error) {
	return func(c *echo.Context) (any, error) {
		id := c.Param("id")
		var input tool.UpdateInput
		if err := json.NewDecoder((*c).Request().Body).Decode(&input); err != nil {
			return nil, err
		}
		result, err := toolUC.Update(c.Request().Context(), id, input)
		if err != nil {
			return nil, err
		}
		return responsew.Success(result, "Tool updated"), nil
	}
}

// @Summary      Delete a tool
// @Description  Remove a tool by ID
// @Tags         tools
// @Produce      json
// @Param        id  path  string  true  "Tool ID"
// @Success      200  {object}  responsew.BaseResponse
// @Failure      404  {object}  responsew.BaseResponse
// @Security     ApiKeyAuth
// @Router       /api/v1/tools/{id} [delete]
func deleteTool(toolUC tool.UseCase) func(c *echo.Context) (any, error) {
	return func(c *echo.Context) (any, error) {
		id := c.Param("id")
		if err := toolUC.Delete(c.Request().Context(), id); err != nil {
			return nil, err
		}
		return responsew.Success(nil, "Tool deleted"), nil
	}
}

// @Summary      List tools
// @Description  Retrieve a paginated list of registered tools
// @Tags         tools
// @Produce      json
// @Param        page      query  int  false  "Page number (default 1)"
// @Param        per_page  query  int  false  "Items per page (default 10)"
// @Success      200  {object}  responsew.BaseResponse
// @Security     ApiKeyAuth
// @Router       /api/v1/tools [get]
func listTools(toolUC tool.UseCase) func(c *echo.Context) (any, error) {
	return func(c *echo.Context) (any, error) {
		page, _ := strconv.Atoi(c.QueryParam("page"))
		perPage, _ := strconv.Atoi(c.QueryParam("per_page"))
		if page < 1 {
			page = 1
		}
		if perPage < 1 {
			perPage = 10
		}

		items, total, err := toolUC.List(c.Request().Context(), page, perPage)
		if err != nil {
			return nil, err
		}
		return responsew.SuccessPaginated(items, total, page, perPage, "Tools retrieved"), nil
	}
}

// @Summary      Search tools by embedding
// @Description  Semantic search for tools using a text query or an embedding vector
// @Tags         tools
// @Accept       json
// @Produce      json
// @Param        body  body  tool.SearchInput  true  "Search query (text or embedding vector)"
// @Success      200  {array}  tool.ToolSearchResult
// @Failure      400  {object}  responsew.BaseResponse
// @Security     ApiKeyAuth
// @Router       /api/v1/tools/search [post]
func searchTools(toolUC tool.UseCase) func(c *echo.Context) (any, error) {
	return func(c *echo.Context) (any, error) {
		var input tool.SearchInput
		if err := json.NewDecoder((*c).Request().Body).Decode(&input); err != nil {
			return nil, err
		}

		// Support raw body parsing for responsew envelope
		// echow.Bind handles the wrapping

		results, err := toolUC.SearchTools(c.Request().Context(), input)
		if err != nil {
			return nil, err
		}
		return responsew.Success(results, "Tools retrieved"), nil
	}
}

// @Summary      Set or update environment variables
// @Description  Add or update environment variables on a tool. Keys must start with MCP_. Existing keys are overwritten; keys not present are kept.
// @Tags         tools
// @Accept       json
// @Produce      json
// @Param        id    path  string              true  "Tool ID"
// @Param        body  body  tool.SetEnvInput    true  "Environment variables to set"
// @Success      200   {object}  entity.Tool
// @Failure      400   {object}  responsew.BaseResponse
// @Failure      404   {object}  responsew.BaseResponse
// @Security     ApiKeyAuth
// @Router       /api/v1/tools/{id}/env [put]
func setToolEnv(toolUC tool.UseCase) func(c *echo.Context) (any, error) {
	return func(c *echo.Context) (any, error) {
		id := c.Param("id")
		var input tool.SetEnvInput
		if err := json.NewDecoder((*c).Request().Body).Decode(&input); err != nil {
			return nil, err
		}
		result, err := toolUC.SetEnv(c.Request().Context(), id, input)
		if err != nil {
			return nil, err
		}
		return responsew.Success(result, "Env vars updated"), nil
	}
}

// @Summary      Delete an environment variable
// @Description  Remove a specific environment variable from a tool by key
// @Tags         tools
// @Produce      json
// @Param        id   path  string  true  "Tool ID"
// @Param        key  path  string  true  "Environment variable key"
// @Success      200  {object}  entity.Tool
// @Failure      404  {object}  responsew.BaseResponse
// @Security     ApiKeyAuth
// @Router       /api/v1/tools/{id}/env/{key} [delete]
func deleteToolEnv(toolUC tool.UseCase) func(c *echo.Context) (any, error) {
	return func(c *echo.Context) (any, error) {
		id := c.Param("id")
		key := c.Param("key")
		result, err := toolUC.DeleteEnv(c.Request().Context(), id, key)
		if err != nil {
			return nil, err
		}
		return responsew.Success(result, "Env var deleted"), nil
	}
}

// Ensure imports used
var _ = http.StatusOK

// @Summary      List all tags
// @Description  Retrieve all tags with descriptions for task classification
// @Tags         tags
// @Produce      json
// @Success      200  {object}  responsew.BaseResponse
// @Security     ApiKeyAuth
// @Router       /api/v1/tags [get]
func listTags(toolUC tool.UseCase) func(c *echo.Context) (any, error) {
	return func(c *echo.Context) (any, error) {
		tags, err := toolUC.ListTags(c.Request().Context())
		if err != nil {
			return nil, err
		}
		return responsew.Success(tags, "Tags retrieved"), nil
	}
}
