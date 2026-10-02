package toolchain

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"time"

	"github.com/codeplai/VizcachaIDE/wails/internal/app"
	"github.com/codeplai/VizcachaIDE/wails/internal/domain"
)

const vetTimeout = 60 * time.Second

// Vet implements app.Toolchain: "go vet <target>" in the background. It does not use
// the run slot and emits no events. Findings (exit code 1 or 2 with text) are
// returned as output, not as an error.
func (t *Toolchain) Vet(ctx context.Context, config domain.RunConfiguration) (string, error) {
	goPath := t.locator.Locate(ToolGo)
	if goPath.Origin == OriginMissing {
		return "", fmt.Errorf("go: %w", app.ErrToolNotFound)
	}
	if info, err := os.Stat(config.WorkingDir); err != nil || !info.IsDir() {
		return "", nil // for example an untitled run, whose folder is already deleted
	}
	ctx, cancel := context.WithTimeout(ctx, vetTimeout)
	defer cancel()
	cmd := exec.CommandContext(ctx, goPath.Path, "vet", config.GoTargetArgument())
	cmd.Dir = config.WorkingDir
	cmd.Env = environmentList(t.Environment())
	hideConsole(cmd)
	var output bytes.Buffer
	cmd.Stdout, cmd.Stderr = &output, &output
	err := cmd.Run()
	var exit *exec.ExitError
	if err != nil && !errors.As(err, &exit) {
		return "", fmt.Errorf("go vet: %w", err)
	}
	return output.String(), nil
}
