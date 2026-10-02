package bridge

import (
	"os"
	"path/filepath"
)

// existingFolder returns the folder itself or its closest existing parent, so a native dialog
// never starts in a folder that was deleted (Windows refuses to open the dialog then).
// Relative paths, such as the "untitled/" of new files, give "" (the system default).
func existingFolder(folder string) string {
	if folder == "" || !filepath.IsAbs(folder) {
		return ""
	}
	for current := filepath.Clean(folder); ; current = filepath.Dir(current) {
		if info, err := os.Stat(current); err == nil && info.IsDir() {
			return current
		}
		if filepath.Dir(current) == current {
			return ""
		}
	}
}

// withStartFolder opens a dialog that starts in folder (or its closest existing parent). If
// the system still refuses that folder, it opens the dialog once more without one.
func withStartFolder(folder string, open func(start string) (string, error)) (string, error) {
	start := existingFolder(folder)
	path, err := open(start)
	if err == nil || start == "" {
		return path, err
	}
	return open("")
}
