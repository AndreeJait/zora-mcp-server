package main

import (
	"path/filepath"
	"time"

	"github.com/AndreeJait/zora-mcp-server/adapter/outbound"
	embedAdp "github.com/AndreeJait/zora-mcp-server/adapter/outbound/embedding"
	execAdp "github.com/AndreeJait/zora-mcp-server/adapter/outbound/executor"
	llmAdp "github.com/AndreeJait/zora-mcp-server/adapter/outbound/llm"
	tagAdp "github.com/AndreeJait/zora-mcp-server/adapter/outbound/tag"
	toolAdp "github.com/AndreeJait/zora-mcp-server/adapter/outbound/tool"
	"github.com/AndreeJait/zora-mcp-server/config"
	"github.com/AndreeJait/zora-mcp-server/port/inbound/health"
	"github.com/AndreeJait/zora-mcp-server/port/inbound/upload"
	"github.com/AndreeJait/zora-mcp-server/port/inbound/mcp"
	"github.com/AndreeJait/zora-mcp-server/port/inbound/tool"
	portOutbound "github.com/AndreeJait/zora-mcp-server/port/outbound"
	"github.com/AndreeJait/zora-mcp-server/usecase"
	"go.uber.org/dig"
	"gorm.io/gorm"
)

func provideServices(c *dig.Container) {
	// kyan:provider:start
	c.Provide(newHealthRepository)
	c.Provide(newHealthUseCase)

	c.Provide(newToolRepository)
	c.Provide(newTagRepository)
	c.Provide(newEmbeddingProvider)
	c.Provide(newLLMProvider)
	c.Provide(newToolUseCase)
	c.Provide(newScriptExecutor)
	c.Provide(newMCPUseCase)
	c.Provide(newUploadUseCase)
	// kyan:provider:end
}

func newHealthRepository(db *outbound.DB, redisConn *outbound.RedisConn) portOutbound.HealthRepository {
	return outbound.NewHealthRepository(db, redisConn)
}

func newHealthUseCase(cfg *config.AppConfig, repo portOutbound.HealthRepository) health.UseCase {
	return usecase.NewHealthUseCase(cfg.App.Name, repo)
}

func newToolRepository(db *gorm.DB) portOutbound.ToolRepository {
	return toolAdp.NewRepository(db)
}

func newTagRepository(db *gorm.DB) portOutbound.TagRepository {
	return tagAdp.NewRepository(db)
}

func newEmbeddingProvider(cfg *config.AppConfig) portOutbound.EmbeddingProvider {
	return embedAdp.NewOllamaProvider(cfg)
}

func newLLMProvider(cfg *config.AppConfig) portOutbound.LLMProvider {
	return llmAdp.NewOllamaProvider(cfg)
}

func newToolUseCase(cfg *config.AppConfig, repo portOutbound.ToolRepository, tagRepo portOutbound.TagRepository, embedProvider portOutbound.EmbeddingProvider, llmProvider portOutbound.LLMProvider, storage portOutbound.Storage) tool.UseCase {
	return usecase.NewToolUseCase(cfg, repo, tagRepo, embedProvider, llmProvider, storage)
}

func newScriptExecutor(cfg *config.AppConfig) portOutbound.ScriptExecutor {
	workDir := filepath.Join("/tmp", "zora-executor")
	return execAdp.NewSandbox(workDir, 60*time.Second)
}

func newMCPUseCase(
	cfg *config.AppConfig,
	repo portOutbound.ToolRepository,
	storage portOutbound.Storage,
	executor portOutbound.ScriptExecutor,
) mcp.UseCase {
	return usecase.NewMCPUseCase(cfg, repo, storage, executor)
}

func newUploadUseCase(storage portOutbound.Storage) upload.UseCase {
	return usecase.NewUploadUseCase(storage)
}