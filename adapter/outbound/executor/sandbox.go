package executor

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"time"

	"github.com/AndreeJait/go-utility/v2/logw"
	portOutbound "github.com/AndreeJait/zora-mcp-server/port/outbound"
)

// Sandbox implements portOutbound.ScriptExecutor by spawning subprocesses.
type Sandbox struct {
	workDir string
	timeout time.Duration
}

var _ portOutbound.ScriptExecutor = (*Sandbox)(nil)

func NewSandbox(workDir string, timeout time.Duration) *Sandbox {
	return &Sandbox{
		workDir: workDir,
		timeout: timeout,
	}
}

func (s *Sandbox) Execute(ctx context.Context, language string, script []byte, args map[string]any, env map[string]string) (*portOutbound.ExecutionResult, error) {
	// Write script to temp file
	ext := scriptExt(language)
	tmpFile, err := os.CreateTemp(s.workDir, fmt.Sprintf("zora-*%s", ext))
	if err != nil {
		return nil, fmt.Errorf("create temp file: %w", err)
	}
	defer os.Remove(tmpFile.Name())

	if _, err := tmpFile.Write(script); err != nil {
		tmpFile.Close()
		return nil, fmt.Errorf("write temp file: %w", err)
	}
	tmpFile.Close()

	if err := os.Chmod(tmpFile.Name(), 0o700); err != nil {
		return nil, fmt.Errorf("chmod: %w", err)
	}

	// Serialize arguments as JSON for the script
	argsJSON, _ := json.Marshal(args)

	// Build command
	cmd := buildCmd(ctx, language, tmpFile.Name(), string(argsJSON))

	// Inject stored env vars into the subprocess.
	// When cmd.Env is nil, Go inherits parent env automatically.
	// When set, it replaces the entire environment, so we must
	// start with os.Environ() and append tool-specific vars.
	if len(env) > 0 {
		cmd.Env = os.Environ()
		for k, v := range env {
			cmd.Env = append(cmd.Env, k+"="+v)
		}
	}

	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	start := time.Now()
	err = cmd.Run()
	duration := time.Since(start)

	result := &portOutbound.ExecutionResult{
		Stdout:     stdout.String(),
		Stderr:     stderr.String(),
		DurationMs: int(duration.Milliseconds()),
	}

	if err != nil {
		if ctx.Err() == context.DeadlineExceeded {
			result.TimedOut = true
			logw.CtxWarningf(ctx, "Sandbox: execution timed out for %s after %dms", language, result.DurationMs)
			return result, nil
		}
		if exitErr, ok := err.(*exec.ExitError); ok {
			result.ExitCode = exitErr.ExitCode()
		}
		logw.CtxErrorf(ctx, "Sandbox: execution failed for %s: exit=%d, stderr=%s",
			language, result.ExitCode, result.Stderr)
		return result, nil // Return result even on error — caller decides
	}

	return result, nil
}

func scriptExt(language string) string {
	switch language {
	case "python":
		return ".py"
	case "go":
		return ".go"
	case "bash":
		return ".sh"
	default:
		return ".txt"
	}
}

func buildCmd(ctx context.Context, language, scriptPath, argsJSON string) *exec.Cmd {
	switch language {
	case "python":
		return exec.CommandContext(ctx, "python3", scriptPath, argsJSON)
	case "go":
		return exec.CommandContext(ctx, "go", "run", scriptPath, argsJSON)
	case "bash":
		return exec.CommandContext(ctx, "bash", scriptPath, argsJSON)
	default:
		return exec.CommandContext(ctx, "bash", scriptPath, argsJSON)
	}
}

func init() {
	// Ensure temp directory exists
	tmpDir := filepath.Join(os.TempDir(), "zora-executor")
	os.MkdirAll(tmpDir, 0o755)
	_ = logw.Info
}
