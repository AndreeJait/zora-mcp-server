package usecase

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/AndreeJait/go-utility/v2/logw"
	"github.com/AndreeJait/zora-mcp-server/config"
	domainEntity "github.com/AndreeJait/zora-mcp-server/domain/entity"
	domainError "github.com/AndreeJait/zora-mcp-server/domain/error"
	"github.com/AndreeJait/zora-mcp-server/port/inbound/mcp"
	"github.com/AndreeJait/zora-mcp-server/port/outbound"
	"github.com/google/uuid"
)

type mcpUseCase struct {
	cfg      *config.AppConfig
	repo     outbound.ToolRepository
	storage  outbound.Storage
	executor outbound.ScriptExecutor
}

var _ mcp.UseCase = (*mcpUseCase)(nil)

func NewMCPUseCase(
	cfg *config.AppConfig,
	repo outbound.ToolRepository,
	storage outbound.Storage,
	executor outbound.ScriptExecutor,
) mcp.UseCase {
	return &mcpUseCase{
		cfg:      cfg,
		repo:     repo,
		storage:  storage,
		executor: executor,
	}
}

func (uc *mcpUseCase) ListTools(ctx context.Context) ([]mcp.ToolDefinition, error) {
	tools, _, err := uc.repo.List(ctx, 1, 1000)
	if err != nil {
		return nil, fmt.Errorf("list tools: %w", err)
	}

	defs := make([]mcp.ToolDefinition, 0, len(tools))
	for _, t := range tools {
		if !t.IsActive {
			continue
		}
		defs = append(defs, mcp.ToolDefinition{
			Name:        t.Name,
			Description: t.Description,
			InputSchema: t.Parameters,
		})
	}

	return defs, nil
}

func (uc *mcpUseCase) CallTool(ctx context.Context, name string, arguments map[string]any) (*domainEntity.ToolExecution, error) {
	tool, err := uc.repo.GetByName(ctx, name)
	if err != nil || tool == nil {
		return nil, domainError.ErrToolNotFound
	}
	if !tool.IsActive {
		return nil, domainError.ErrToolNotFound.WithCustomMessage("tool is inactive: " + name)
	}

	// Fetch script from MinIO
	script, err := uc.storage.GetObject(ctx, tool.Bucket, tool.ObjectKey)
	if err != nil {
		logw.CtxErrorf(ctx, "mcpUseCase: fetch script %s failed: %v", name, err)
		return nil, domainError.ErrScriptFetchFailed.WithError(err)
	}

	// Execute with timeout context
	execCtx, cancel := context.WithTimeout(ctx, 60*time.Second)
	defer cancel()

	start := time.Now()
	result, err := uc.executor.Execute(execCtx, tool.Language, script, arguments, tool.Env)
	durationMs := int(time.Since(start).Milliseconds())

	exec := &domainEntity.ToolExecution{
		ToolID:     tool.ID,
		TraceID:    logw.GetLogID(ctx),
		Parameters: arguments,
		DurationMs: durationMs,
	}

	if err != nil {
		exec.Status = "failed"
		exec.ErrorMsg = err.Error()
		exec.DurationMs = durationMs
		_ = uc.repo.SaveExecution(ctx, exec)
		return exec, nil
	}

	exec.Status = "success"
	if result.TimedOut {
		exec.Status = "timeout"
		exec.ErrorMsg = "execution timed out"
	}

	// Try to parse stdout as JSON result
	var parsedResult map[string]any
	if err := json.Unmarshal([]byte(result.Stdout), &parsedResult); err != nil {
		parsedResult = map[string]any{
			"stdout": result.Stdout,
			"stderr": result.Stderr,
		}
	}

	// Handle _files convention: upload file entries to MinIO
	if filesRaw, ok := parsedResult["_files"]; ok {
		if filesList, ok := filesRaw.([]any); ok {
			var savedFiles []map[string]any
			for _, f := range filesList {
				fm, ok := f.(map[string]any)
				if !ok {
					continue
				}
				filename, _ := fm["filename"].(string)
				content, _ := fm["content"].(string)
				contentType, _ := fm["content_type"].(string)
				if filename == "" || content == "" {
					continue
				}
				if contentType == "" {
					contentType = "text/plain"
				}

				objectKey := fmt.Sprintf("tools/%s-%s", uuid.New().String(), filename)
				data := []byte(content)
				reader := bytes.NewReader(data)

				if err := uc.storage.Upload(ctx, "zora-scripts", objectKey, reader, int64(len(data)), contentType); err != nil {
					logw.CtxWarningf(ctx, "mcpUseCase: failed to upload file %s: %v", filename, err)
					continue
				}

				url, _ := uc.storage.GetPresignedURL(ctx, "zora-scripts", objectKey, 24*time.Hour)
				savedFiles = append(savedFiles, map[string]any{
					"object_key": objectKey,
					"bucket":     "zora-scripts",
					"filename":   filename,
					"url":        url,
				})
			}
			if len(savedFiles) > 0 {
				parsedResult["saved_files"] = savedFiles
			}
		}
		delete(parsedResult, "_files")
	}

	exec.Result = parsedResult

	if err := uc.repo.SaveExecution(ctx, exec); err != nil {
		logw.CtxErrorf(ctx, "mcpUseCase: save execution failed: %v", err)
	}

	logw.CtxInfof(ctx, "mcpUseCase: executed tool %s in %dms (status=%s)", name, durationMs, exec.Status)
	return exec, nil
}
