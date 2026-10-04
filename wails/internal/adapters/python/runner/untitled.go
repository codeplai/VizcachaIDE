package runner

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/codeplai/VizcachaIDE/wails/internal/app"
	"github.com/codeplai/VizcachaIDE/wails/internal/domain"
)

// RunUntitled implements app.ProgramRunner: it runs unsaved source from a temporary folder
// that is deleted when the program ends.
func (r *Runner) RunUntitled(ctx context.Context, path, source string, programArgs []string) (domain.RunConfiguration, error) {
	if r.supervisor.IsRunning() {
		return domain.RunConfiguration{}, app.ErrBusy
	}
	dir := filepath.Join(os.TempDir(), fmt.Sprintf("vizcacha_py_%d", os.Getpid()))
	if err := os.MkdirAll(dir, 0o700); err != nil {
		return domain.RunConfiguration{}, fmt.Errorf("create temporary folder: %w", err)
	}
	cleanup := func() { _ = os.RemoveAll(dir) }
	file := filepath.Join(dir, untitledFileName(path))
	if err := os.WriteFile(file, []byte(source), 0o600); err != nil {
		cleanup()
		return domain.RunConfiguration{}, fmt.Errorf("save the untitled file: %w", err)
	}
	config := domain.NewFileRunConfiguration(domain.CodeLanguagePython, file, programArgs)
	config.Project = &domain.ProjectContext{Root: dir, Kind: domain.ProjectFolder}
	return config, r.start(ctx, config, cleanup)
}

// untitledFileName keeps the name of the untitled tab ("untitled-1.py") when it is a Python
// file name, and uses main.py otherwise.
func untitledFileName(path string) string {
	name := filepath.Base(strings.ReplaceAll(path, "\\", "/"))
	extension := strings.ToLower(filepath.Ext(name))
	if path == "" || (extension != ".py" && extension != ".pyw") || name == extension {
		return "main.py"
	}
	return name
}
