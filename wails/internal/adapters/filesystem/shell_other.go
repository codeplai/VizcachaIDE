//go:build !windows && !darwin

package filesystem

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/codeplai/VizcachaIDE/wails/internal/app"
)

func trash(path string) error {
	if err := run(true, "gio", "trash", path); err != nil {
		return fmt.Errorf("%w: %v", app.ErrTrashUnavailable, err)
	}
	return nil
}

func reveal(path string) error {
	folder := path
	if info, err := os.Stat(path); err != nil || !info.IsDir() {
		folder = filepath.Dir(path)
	}
	return run(false, "xdg-open", folder)
}
