package cpp

import (
	"context"
	"path/filepath"

	"github.com/codeplai/VizcachaIDE/wails/internal/domain"
)

// ShellPaths implements app.ShellPaths: the bin folder of the bundled LLVM (clang, clang++, g++)
// and the folder of a compiler the user configured in Settings.
func (l *Locator) ShellPaths(context.Context) []string {
	paths := l.fixed.ExistingDirectories(bundledBin)
	if status := l.fixed.Locate(toolFor(ToolCompiler, "clang++")); status.Source == domain.ToolConfigured {
		paths = append([]string{filepath.Dir(status.Path)}, paths...)
	}
	return paths
}
