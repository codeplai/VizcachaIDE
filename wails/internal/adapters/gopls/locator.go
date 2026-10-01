package gopls

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"

	"github.com/codeplai/VizcachaIDE/wails/internal/app"
)

// locateGopls finds the gopls executable: the configured path, then the toolchain's
// PATH, GOBIN, GOPATH/bin, and finally the current PATH and ~/go/bin.
func locateGopls(configured string, env map[string]string) (string, error) {
	if configured != "" {
		if isFile(configured) {
			return configured, nil
		}
		return "", fmt.Errorf("gopls at %q: %w", configured, app.ErrToolNotFound)
	}
	name := "gopls"
	if runtime.GOOS == "windows" {
		name = "gopls.exe"
	}
	for _, dir := range searchDirs(env) {
		if candidate := filepath.Join(dir, name); isFile(candidate) {
			return candidate, nil
		}
	}
	if found, err := exec.LookPath("gopls"); err == nil {
		return found, nil
	}
	return "", fmt.Errorf("gopls: %w", app.ErrToolNotFound)
}

func searchDirs(env map[string]string) []string {
	var dirs []string
	dirs = append(dirs, filepath.SplitList(valueOf(env, "PATH"))...)
	dirs = append(dirs, valueOf(env, "GOBIN"))
	for _, gopath := range filepath.SplitList(valueOf(env, "GOPATH")) {
		dirs = append(dirs, filepath.Join(gopath, "bin"))
	}
	if home, err := os.UserHomeDir(); err == nil {
		dirs = append(dirs, filepath.Join(home, "go", "bin"))
	}
	return dirs
}

// valueOf reads a variable from env, ignoring the case of the key (Windows "Path").
func valueOf(env map[string]string, key string) string {
	for name, value := range env {
		if strings.EqualFold(name, key) {
			return value
		}
	}
	return ""
}

func isFile(path string) bool {
	info, err := os.Stat(path)
	return err == nil && !info.IsDir()
}
