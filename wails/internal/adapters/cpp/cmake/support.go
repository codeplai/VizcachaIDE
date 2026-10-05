// Package cmake is how VizcachaIDE builds C++ projects: every project is a CMake project
// (docs/PLAN_CPP_CMAKE.md). It finds and generates the project files, configures and builds with
// CMake and Ninja, finds the program to run through the CMake File API and filters the build's
// noise. Libraries come from vcpkg through the Dependencies interface, which this package does
// not know the implementation of.
package cmake

import (
	"bytes"
	_ "embed" // the console helper and the CMake script are embedded data
	"os"
	"path/filepath"
	"runtime"
	"strings"
)

// utf8Console sets the Windows console to UTF-8 before main runs. std::cout writes UTF-8 bytes
// and the console read them with its OEM code page: "¿Cómo" showed as "┐C├│mo". Go, Python and
// Rust write UTF-16 to the console and need nothing.
//
//go:embed data/vizcacha_console_utf8.cpp
var utf8Console []byte

// includeTemplate is vizcacha.cmake with the helper's path still to fill in.
//
//go:embed data/vizcacha.cmake
var includeTemplate string

const (
	consoleFile = "vizcacha_console_utf8.cpp"
	includeFile = "vizcacha.cmake"
)

// ConsoleSource writes the UTF-8 console helper into dir (once) and returns its path. It is ""
// outside Windows, where nothing is needed, or when the file cannot be written. Direct
// compilations (an untitled file) and the CMake include share the same copy.
func ConsoleSource(dir string) string {
	if runtime.GOOS != "windows" {
		return ""
	}
	return writeIfChanged(filepath.Join(dir, consoleFile), utf8Console)
}

// Include writes vizcacha.cmake (and the helper it adds) into dir and returns the path to give
// to CMAKE_PROJECT_TOP_LEVEL_INCLUDES, or "" when it is not needed (not Windows) or cannot be
// written.
func Include(dir string) string {
	console := ConsoleSource(dir)
	if console == "" {
		return ""
	}
	script := strings.ReplaceAll(includeTemplate, "@CONSOLE_SOURCE@", filepath.ToSlash(console))
	return writeIfChanged(filepath.Join(dir, includeFile), []byte(script))
}

// writeIfChanged writes content unless the file already holds it; it returns the path, or "".
func writeIfChanged(path string, content []byte) string {
	if current, err := os.ReadFile(path); err == nil && bytes.Equal(current, content) {
		return path
	}
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return ""
	}
	if err := os.WriteFile(path, content, 0o600); err != nil {
		return ""
	}
	return path
}
