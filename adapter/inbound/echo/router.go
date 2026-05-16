package echo

import (
	"github.com/AndreeJait/zora-mcp-server/domain/entity"
	"github.com/AndreeJait/zora-mcp-server/port/inbound/health"
	"github.com/AndreeJait/zora-mcp-server/port/inbound/mcp"
	"github.com/AndreeJait/zora-mcp-server/port/inbound/tool"
	"github.com/AndreeJait/zora-mcp-server/port/inbound/upload"
	httpw "github.com/AndreeJait/go-utility/v2/httpw/echow"
	"github.com/AndreeJait/go-utility/v2/responsew"
	"github.com/labstack/echo/v5"
)

// RouteRegistrar is implemented by both *echo.Echo and *echo.Group,
// allowing route registration functions to accept either.
type RouteRegistrar interface {
	GET(path string, h echo.HandlerFunc, m ...echo.MiddlewareFunc) echo.RouteInfo
	POST(path string, h echo.HandlerFunc, m ...echo.MiddlewareFunc) echo.RouteInfo
	PUT(path string, h echo.HandlerFunc, m ...echo.MiddlewareFunc) echo.RouteInfo
	DELETE(path string, h echo.HandlerFunc, m ...echo.MiddlewareFunc) echo.RouteInfo
}

var _ = entity.Health{}

func RegisterRoutes(e *echo.Echo, healthUC health.UseCase, toolUC tool.UseCase, mcpUC mcp.UseCase, uploadUC upload.UseCase, apiKey string) {
	e.GET("/health", httpw.Bind(checkHealth(healthUC)))

	// Management routes — protected by API key if configured
	mgmt := e.Group("", APIKeyMiddleware(apiKey))
	RegisterToolRoutes(mgmt, toolUC)
	RegisterUploadRoutes(mgmt, uploadUC)
	RegisterStorageRoutes(mgmt, uploadUC)

	// MCP routes — open for internal service calls
	RegisterMCPRoutes(e, mcpUC)
}

// @Summary      Health check
// @Description  Check if the service is healthy including DB and Redis connectivity
// @Tags         health
// @Produce      json
// @Success      200  {object}  entity.Health
// @Router       /health [get]
func checkHealth(healthUC health.UseCase) func(c *echo.Context) (any, error) {
	return func(c *echo.Context) (any, error) {
		h := healthUC.Check(c.Request().Context())
		return responsew.Success(h, "Service is healthy"), nil
	}
}
