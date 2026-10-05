package runner

import (
	"context"
	"path/filepath"

	"github.com/codeplai/VizcachaIDE/wails/internal/domain"
)

// ShellPaths implements app.ShellPaths: the folders of the Go the IDE uses, for the PATH of the
// integrated terminal. The bundled toolchain gives go and the tools shipped beside it; a Go the
// user configured in Settings gives its own folder. A Go found on PATH needs nothing.
func (r *Runner) ShellPaths(context.Context) []string {
	status := r.Locate("go")
	switch status.Source {
	case domain.ToolBundled:
		return r.locator.ExistingDirectories(bundledGoBinary, bundledToolFolder)
	case domain.ToolConfigured:
		return []string{filepath.Dir(status.Path)}
	default:
		return nil
	}
}
