package main

import (
	"fmt"
	"net/http"

	echoAdapter "github.com/AndreeJait/zora-mcp-server/adapter/inbound/echo"
	"github.com/AndreeJait/zora-mcp-server/config"
	"github.com/AndreeJait/zora-mcp-server/port/inbound/health"
	"github.com/AndreeJait/zora-mcp-server/port/inbound/mcp"
	"github.com/AndreeJait/zora-mcp-server/port/inbound/tool"
	"github.com/AndreeJait/zora-mcp-server/port/inbound/upload"
	httpwEcho "github.com/AndreeJait/go-utility/v2/httpw/echow"
	"go.uber.org/dig"
)

func provideRouter(c *dig.Container) {
	c.Provide(newRouter)
}

func newRouter(
	// kyan:param:start
	cfg *config.AppConfig,
	healthUC health.UseCase,
	toolUC tool.UseCase,
	mcpUC mcp.UseCase,
	uploadUC upload.UseCase,
	// kyan:param:end
) (http.Handler, error) {
	switch cfg.HTTP.Engine {
	case "echo":
		e := httpwEcho.New(&httpwEcho.Config{
			DebugMode:     cfg.HTTP.DebugMode,
			EnableSwagger: cfg.HTTP.EnableSwagger,
		})
		echoAdapter.RegisterRoutes(e, healthUC, toolUC, mcpUC, uploadUC, cfg.HTTP.APIKey)
		return e, nil
	default:
		return nil, fmt.Errorf("unsupported engine: %s (zora-mcp-server only supports echo)", cfg.HTTP.Engine)
	}
}
