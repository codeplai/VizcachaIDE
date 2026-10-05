package runner

import (
	"path/filepath"

	"github.com/codeplai/VizcachaIDE/wails/internal/adapters/cpp/cmake"
)

// consoleSource is the UTF-8 console helper to compile with the sources of a direct compilation
// (an untitled file; CMake projects get it from vizcacha.cmake): its path in the IDE's build
// cache on Windows, "" elsewhere or when it cannot be written (see cmake.ConsoleSource).
func (r *Runner) consoleSource() string {
	return cmake.ConsoleSource(filepath.Dir(r.buildFolder("")))
}
