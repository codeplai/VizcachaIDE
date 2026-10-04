//go:build windows

package repl

import (
	"context"
	"os/exec"
	"strconv"
	"time"

	"github.com/codeplai/VizcachaIDE/wails/internal/protocol/process"
)

// killTree ends pid and its descendants (taskkill /T /F). A venv's python.exe is a launcher that
// starts the real interpreter as a child, so killing only the launcher would leave the code of
// the student running.
func killTree(pid int) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	cmd := exec.CommandContext(ctx, "taskkill", "/PID", strconv.Itoa(pid), "/T", "/F")
	process.HideConsole(cmd)
	_ = cmd.Run() // best effort: kill also kills the direct child
}
