package rust

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	"github.com/codeplai/VizcachaIDE/wails/internal/protocol/process"
)

// probeTimeout bounds "rustc -vV": a rustup proxy may first sync a rust-toolchain.toml.
const probeTimeout = 20 * time.Second

// runTool runs a tool with a time limit in env and returns stdout and stderr together.
func runTool(ctx context.Context, env []string, path string, args ...string) (string, error) {
	ctx, cancel := context.WithTimeout(ctx, probeTimeout)
	defer cancel()
	cmd := exec.CommandContext(ctx, path, args...)
	cmd.Env = env
	process.HideConsole(cmd)
	output, err := cmd.CombinedOutput()
	if err != nil {
		return string(output), fmt.Errorf("%s %s: %w", filepath.Base(path), strings.Join(args, " "), err)
	}
	return string(output), nil
}

// cargoBin is where rustup puts its proxies: CARGO_HOME/bin, or ~/.cargo/bin.
func cargoBin(env []string) string {
	if home := variable(env, "CARGO_HOME"); home != "" {
		return filepath.Join(home, "bin")
	}
	home := variable(env, "USERPROFILE")
	if home == "" {
		home = variable(env, "HOME")
	}
	if home == "" {
		return ""
	}
	return filepath.Join(home, ".cargo", "bin")
}

// variable returns the value of a "NAME=value" entry (names compared without case, as Windows
// spells PATH "Path"), or "".
func variable(env []string, name string) string {
	for _, entry := range env {
		key, value, ok := strings.Cut(entry, "=")
		if ok && strings.EqualFold(key, name) {
			return value
		}
	}
	return ""
}

// withPathFirst puts dir in front of PATH (adding PATH when there is none).
func withPathFirst(env []string, dir string) []string {
	result := make([]string, 0, len(env)+1)
	found := false
	for _, entry := range env {
		key, value, _ := strings.Cut(entry, "=")
		if strings.EqualFold(key, "PATH") {
			entry, found = key+"="+dir+string(os.PathListSeparator)+value, true
		}
		result = append(result, entry)
	}
	if !found {
		result = append(result, "PATH="+dir)
	}
	return result
}

func applicationDirectory(given string) string {
	if given != "" {
		return given
	}
	executable, err := os.Executable()
	if err != nil {
		return "."
	}
	return filepath.Dir(executable)
}
