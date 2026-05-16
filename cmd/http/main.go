package main

import (
	"flag"
	"fmt"
	"net/http"
	"os"

	"github.com/AndreeJait/go-utility/v2/gracefulw"
	"github.com/AndreeJait/go-utility/v2/logw"
	"github.com/AndreeJait/zora-mcp-server/config"
	_ "github.com/AndreeJait/zora-mcp-server/docs"
	docs "github.com/AndreeJait/zora-mcp-server/docs"
)

// @title Zora MCP Server API
// @version 1.0
// @description Tool registry and execution service for the Zora agent platform.
// @securityDefinitions.apikey ApiKeyAuth
// @in header
// @name X-API-Key

func main() {
	engineFlag := flag.String("engine", "", "HTTP engine: echo|gin|mux (overrides config file)")
	configFlag := flag.String("config", "files/config/app.yaml", "Path to config file")
	flag.Parse()

	// Load configuration
	cfg, err := config.Load(*configFlag)
	if err != nil {
		fmt.Fprintf(os.Stderr, "failed to load config: %v\n", err)
		os.Exit(1)
	}

	// CLI flag overrides config file
	if *engineFlag != "" {
		cfg.HTTP.Engine = *engineFlag
	}

	// Override swagger host from config port
	docs.SwaggerInfo.Host = fmt.Sprintf("localhost:%d", cfg.App.HTTPPort)

	// Initialize logger
	if err := logw.Init(&logw.LogConfig{
		Level:       cfg.Log.Level,
		Format:      cfg.Log.Format,
		WriteToFile: cfg.Log.WriteToFile,
		FilePath:    cfg.Log.FilePath,
	}); err != nil {
		fmt.Fprintf(os.Stderr, "failed to init logger: %v\n", err)
		os.Exit(1)
	}

	logw.Infof("Starting %s with engine: %s", cfg.App.Name, cfg.HTTP.Engine)

	// Wire all dependencies
	handler, cleanup, err := wire(cfg)
	if err != nil {
		logw.Errorf("failed to wire dependencies: %v", err)
		os.Exit(1)
	}

	// Start server with graceful shutdown
	addr := fmt.Sprintf(":%d", cfg.App.HTTPPort)
	srv := &http.Server{Addr: addr, Handler: handler}

	gracefulw.Register("http-server", srv.Shutdown)
	gracefulw.Register("dependencies", cleanup)

	logw.Infof("HTTP server listening on %s", addr)
	gracefulw.Start(srv.ListenAndServe, cfg.Graceful.ShutdownTimeout)
}
