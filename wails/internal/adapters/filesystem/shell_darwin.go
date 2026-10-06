//go:build darwin

package filesystem

import (
	"fmt"
	"strings"

	"github.com/codeplai/VizcachaIDE/wails/internal/app"
)

func trash(path string) error {
	if err := run(true, "osascript", "-e", trashScript(path)); err != nil {
		return fmt.Errorf("%w: %w", app.ErrTrashUnavailable, err)
	}
	return nil
}

func trashScript(path string) string {
	escaped := strings.NewReplacer(`\`, `\\`, `"`, `\"`).Replace(path)
	return `tell application "Finder" to delete POSIX file "` + escaped + `"`
}

func reveal(path string) error { return run(false, "open", "-R", path) }
