package cpp

import (
	"context"
	"fmt"
	"os/exec"
	"strings"
	"time"

	"github.com/codeplai/VizcachaIDE/wails/internal/protocol/process"
)

const probeTimeout = 5 * time.Second

// runVersion runs "<compiler> --version" with a time limit.
func runVersion(ctx context.Context, path string) (string, error) {
	ctx, cancel := context.WithTimeout(ctx, probeTimeout)
	defer cancel()
	cmd := exec.CommandContext(ctx, path, "--version")
	process.HideConsole(cmd)
	output, err := cmd.Output()
	if err != nil {
		return "", fmt.Errorf("probing %s: %w", path, err)
	}
	return string(output), nil
}

// xcodeReady reports whether "xcode-select -p" answers (macOS Command Line Tools installed).
func xcodeReady() bool {
	ctx, cancel := context.WithTimeout(context.Background(), probeTimeout)
	defer cancel()
	return exec.CommandContext(ctx, "xcode-select", "-p").Run() == nil
}

// searchPath returns the PATH of a "NAME=value" list (Windows usually spells it "Path").
func searchPath(base []string) string {
	for _, entry := range base {
		name, value, ok := strings.Cut(entry, "=")
		if ok && strings.EqualFold(name, "PATH") {
			return value
		}
	}
	return ""
}
