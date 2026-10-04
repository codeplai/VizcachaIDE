package runner

import (
	"context"
	"os/exec"
	"regexp"
	"strings"
	"sync"
	"time"

	"github.com/codeplai/VizcachaIDE/wails/internal/adapters/golang"
	"github.com/codeplai/VizcachaIDE/wails/internal/domain"
	"github.com/codeplai/VizcachaIDE/wails/internal/protocol/process"
	"github.com/codeplai/VizcachaIDE/wails/internal/protocol/toollocator"
)

const versionTimeout = 5 * time.Second

var versionNumber = regexp.MustCompile(`\d+\.\d+(?:\.\d+)?`)

// toolFor describes one tool of the Go profile for the locator. Go ships in the GOROOT of the
// bundled toolchain; Delve and gopls ship in toolchain/bin.
func toolFor(toolID string) toollocator.Tool {
	tool := toollocator.Tool{Language: domain.CodeLanguageGo, BundledDirectories: []string{bundledToolFolder}}
	for _, spec := range golang.Profile.Tools {
		if spec.ID == toolID {
			tool.Spec = spec
		}
	}
	if toolID == "go" {
		tool.BundledDirectories = []string{bundledGoBinary}
	}
	return tool
}

// Locate reports where one tool ("go", "dlv" or "gopls") comes from, without its version.
func (r *Runner) Locate(toolID string) domain.ToolStatus {
	return r.locator.Locate(toolFor(toolID))
}

// Tools implements app.ProgramRunner: where each Go tool was found and its version.
func (r *Runner) Tools(ctx context.Context) []domain.ToolStatus {
	statuses := make([]domain.ToolStatus, len(golang.Profile.Tools))
	var wg sync.WaitGroup
	for index, spec := range golang.Profile.Tools {
		wg.Add(1)
		go func() {
			defer wg.Done()
			status := r.Locate(spec.ID)
			status.Version = r.version(ctx, status)
			statuses[index] = status
		}()
	}
	wg.Wait()
	return statuses
}

// version runs "<tool> version" and returns the first version number it prints ("" if it fails).
func (r *Runner) version(ctx context.Context, status domain.ToolStatus) string {
	if status.Source == domain.ToolMissing {
		return ""
	}
	ctx, cancel := context.WithTimeout(ctx, versionTimeout)
	defer cancel()
	cmd := exec.CommandContext(ctx, status.Path, "version")
	cmd.Env = environmentList(r.Environment())
	process.HideConsole(cmd)
	output, err := cmd.Output()
	if err != nil {
		return ""
	}
	return parseVersion(string(output))
}

func parseVersion(output string) string {
	return strings.TrimSpace(versionNumber.FindString(output))
}
