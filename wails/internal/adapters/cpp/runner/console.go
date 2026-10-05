package runner

import (
	"bytes"
	_ "embed" // the console helper is embedded data
	"os"
	"path/filepath"
	"runtime"
)

// utf8Console sets the Windows console to UTF-8 before main runs (data/vizcacha_console_utf8.cpp).
// std::cout writes UTF-8 bytes and the console read them with its OEM code page: "¿Cómo"
// showed as "┐C├│mo" (found by the New project QA). Go, Python and Rust write UTF-16 to the
// console and need nothing.
//
//go:embed data/vizcacha_console_utf8.cpp
var utf8Console []byte

// consoleSource is the helper to compile with the student's sources: its path in the IDE's build
// cache on Windows (written once), "" elsewhere or when it cannot be written.
func (r *Runner) consoleSource() string {
	if runtime.GOOS != "windows" {
		return ""
	}
	path := filepath.Join(filepath.Dir(r.buildFolder("")), "vizcacha_console_utf8.cpp")
	if current, err := os.ReadFile(path); err == nil && bytes.Equal(current, utf8Console) {
		return path
	}
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return ""
	}
	if err := os.WriteFile(path, utf8Console, 0o600); err != nil {
		return ""
	}
	return path
}
