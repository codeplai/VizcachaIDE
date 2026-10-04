// Package repl is the interactive console of Python (docs/PLAN_PYTHON.md section 4.7): an
// app.Console that talks JSON lines to a helper script (repl.py) running in the student's
// interpreter. The session lives in that process, so variables persist between snippets.
package repl

import (
	"bytes"
	_ "embed" // repl.py is embedded and copied to the cache folder
	"fmt"
	"os"
	"path/filepath"
)

//go:embed repl.py
var script []byte

// scriptFolder is where repl.py is kept, under the user cache folder.
const scriptFolder = "VizcachaIDE/repl"

// installScript copies the embedded repl.py to <cacheDir>/VizcachaIDE/repl/repl.py, rewriting
// it when the embedded content changed (a new IDE version), and returns its path.
func installScript(cacheDir string) (string, error) {
	if cacheDir == "" {
		userCache, err := os.UserCacheDir()
		if err != nil {
			return "", fmt.Errorf("finding the cache folder: %w", err)
		}
		cacheDir = userCache
	}
	folder := filepath.Join(cacheDir, filepath.FromSlash(scriptFolder))
	path := filepath.Join(folder, "repl.py")
	if current, err := os.ReadFile(path); err == nil && bytes.Equal(current, script) {
		return path, nil
	}
	if err := os.MkdirAll(folder, 0o755); err != nil {
		return "", fmt.Errorf("creating %s: %w", folder, err)
	}
	if err := os.WriteFile(path, script, 0o644); err != nil {
		return "", fmt.Errorf("writing %s: %w", path, err)
	}
	return path, nil
}
