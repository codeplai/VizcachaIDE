package cpp

import (
	"context"
	"path/filepath"

	"github.com/codeplai/VizcachaIDE/wails/internal/domain"
)

// ShellPaths implements app.ShellPaths: the bin folder of the bundled LLVM (clang, clang++, g++),
// the one of CMake, and the folders of a compiler or CMake the user configured in Settings.
func (l *Locator) ShellPaths(context.Context) []string {
	paths := l.fixed.ExistingDirectories(bundledBin, bundledCMake)
	if status := l.fixed.Locate(toolFor(ToolCMake, ToolCMake)); status.Source == domain.ToolConfigured {
		paths = append([]string{filepath.Dir(status.Path)}, paths...)
	}
	if status := l.fixed.Locate(toolFor(ToolCompiler, "clang++")); status.Source == domain.ToolConfigured {
		paths = append([]string{filepath.Dir(status.Path)}, paths...)
	}
	return paths
}
