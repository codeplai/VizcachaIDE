package runner

import (
	"context"
	"fmt"
	"os"
	"path/filepath"

	"github.com/codeplai/VizcachaIDE/wails/internal/app"
	"github.com/codeplai/VizcachaIDE/wails/internal/domain"
)

// RunUntitled implements app.ProgramRunner: it writes the unsaved source as main.cpp in a
// temporary folder, runs it and deletes the folder and its build when the program ends. Stopping
// the run while the compiler is still working leaves the folder until the next untitled run
// (same name) overwrites it.
func (r *Runner) RunUntitled(ctx context.Context, _, source string, programArgs []string) (domain.RunConfiguration, error) {
	if r.supervisor.IsRunning() {
		return domain.RunConfiguration{}, app.ErrBusy
	}
	dir := filepath.Join(os.TempDir(), fmt.Sprintf("vizcacha_cpp_%d", os.Getpid()))
	if err := os.MkdirAll(dir, 0o700); err != nil {
		return domain.RunConfiguration{}, fmt.Errorf("create temporary folder: %w", err)
	}
	build := r.buildFolder(dir)
	cleanup := func() {
		_ = os.RemoveAll(dir)
		_ = os.RemoveAll(build)
	}
	file := filepath.Join(dir, "main.cpp")
	if err := os.WriteFile(file, []byte(source), 0o600); err != nil {
		cleanup()
		return domain.RunConfiguration{}, fmt.Errorf("save the untitled file: %w", err)
	}
	config := r.Configure(file, programArgs)
	return config, r.start(ctx, config, cleanup)
}
