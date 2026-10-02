// Package filesystem talks to the operating system's shell: the Recycle Bin and the
// file manager ("Show in Explorer"). The OS-specific parts live in the shell_windows.go,
// shell_darwin.go and shell_other.go files.
package filesystem

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/codeplai/VizcachaIDE/wails/internal/app"
)

// Shell implements app.SystemShell.
type Shell struct{}

var _ app.SystemShell = Shell{}

// New creates the adapter.
func New() Shell { return Shell{} }

// MoveToTrash sends the path to the Recycle Bin. A path that does not exist is an error.
// Nothing is ever deleted permanently.
func (Shell) MoveToTrash(path string) error {
	abs, err := filepath.Abs(path)
	if err != nil {
		return fmt.Errorf("trash %s: %w", path, err)
	}
	if _, err := os.Lstat(abs); err != nil {
		return fmt.Errorf("trash %s: %w", path, err)
	}
	return trash(abs)
}

// Reveal shows the path in the system's file manager.
func (Shell) Reveal(path string) error {
	abs, err := filepath.Abs(path)
	if err != nil {
		return fmt.Errorf("reveal %s: %w", path, err)
	}
	return reveal(abs)
}
