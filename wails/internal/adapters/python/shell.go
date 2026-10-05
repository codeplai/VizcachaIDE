package python

import (
	"context"
	"os"
	"path/filepath"
	"runtime"
)

// ShellPaths implements app.ShellPaths: the folder of the interpreter the IDE uses (configured,
// else bundled) and, on Windows, its Scripts folder, where pip and the installed commands live.
// Without an interpreter there is nothing to add.
func (l *Locator) ShellPaths(ctx context.Context) []string {
	interpreter, err := l.Find(ctx, "")
	if err != nil {
		return nil
	}
	dir := filepath.Dir(interpreter.Path)
	paths := []string{dir}
	if runtime.GOOS == "windows" {
		if scripts := filepath.Join(dir, "Scripts"); isFolder(scripts) {
			paths = append(paths, scripts)
		}
	}
	return paths
}

func isFolder(path string) bool {
	info, err := os.Stat(path)
	return err == nil && info.IsDir()
}
