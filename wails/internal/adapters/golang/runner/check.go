package runner

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"go/format"
	"go/scanner"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"time"

	"github.com/codeplai/VizcachaIDE/wails/internal/adapters/golang"
	"github.com/codeplai/VizcachaIDE/wails/internal/app"
	"github.com/codeplai/VizcachaIDE/wails/internal/domain"
	"github.com/codeplai/VizcachaIDE/wails/internal/protocol/process"
)

const vetTimeout = 60 * time.Second

func isWindows() bool { return runtime.GOOS == "windows" }

// Check implements app.CodeChecker: "go vet <target>" in the background. It does not use the
// run slot and emits no events. Findings (exit code 1 or 2 with text) are returned as output,
// not as an error.
func (r *Runner) Check(ctx context.Context, config domain.RunConfiguration) (string, error) {
	goTool := r.Locate("go")
	if goTool.Source == domain.ToolMissing {
		return "", app.MissingTool("go")
	}
	if info, err := os.Stat(config.WorkingDir); err != nil || !info.IsDir() {
		return "", nil // for example an untitled run, whose folder is already deleted
	}
	ctx, cancel := context.WithTimeout(ctx, vetTimeout)
	defer cancel()
	cmd := exec.CommandContext(ctx, goTool.Path, "vet", golang.GoTargetArgument(config))
	cmd.Dir = config.WorkingDir
	cmd.Env = environmentList(r.Environment())
	process.HideConsole(cmd)
	var output bytes.Buffer
	cmd.Stdout, cmd.Stderr = &output, &output
	err := cmd.Run()
	var exit *exec.ExitError
	if err != nil && !errors.As(err, &exit) {
		return "", fmt.Errorf("go vet: %w", err)
	}
	return output.String(), nil
}

// Format implements app.CodeFormatter with the standard go/format package. A syntax error
// wraps app.ErrFormat as "<file>:<line>:<column>: <message>", the form the frontend reads to name
// the line (errors.formatRejected).
func (r *Runner) Format(path, text string) (string, error) {
	formatted, err := format.Source([]byte(text))
	if err == nil {
		return string(formatted), nil
	}
	var list scanner.ErrorList
	if errors.As(err, &list) && len(list) > 0 {
		position := list[0].Pos
		return "", fmt.Errorf("%w: %s:%d:%d: %s", app.ErrFormat, filepath.Base(path), position.Line, position.Column, list[0].Msg)
	}
	return "", fmt.Errorf("%w: %w", app.ErrFormat, err)
}
