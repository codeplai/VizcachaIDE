// Package clangformat is the C++ formatter: clang-format with the text on stdin.
//
// clang-format does not fail on syntax errors: it formats what it can and leaves the rest, so
// unlike ruff this formatter never reports a place in the code. Only a failure to run it (for
// example an unreadable .clang-format) is an app.ErrFormat.
package clangformat

import (
	"bytes"
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	"github.com/codeplai/VizcachaIDE/wails/internal/adapters/cpp"
	"github.com/codeplai/VizcachaIDE/wails/internal/app"
	"github.com/codeplai/VizcachaIDE/wails/internal/domain"
	"github.com/codeplai/VizcachaIDE/wails/internal/protocol/process"
)

const (
	formatTimeout = 30 * time.Second
	// defaultStyle applies when no .clang-format exists: the plan's 4-space indent. It goes in
	// --style because --fallback-style only accepts the names of predefined styles.
	defaultStyle = "{BasedOnStyle: LLVM, IndentWidth: 4}"
	styleFile    = "file"
)

// Tool implements app.CodeFormatter with clang-format.
type Tool struct{ locator *cpp.Locator }

var _ app.CodeFormatter = (*Tool)(nil)

// New creates the formatter.
func New(locator *cpp.Locator) *Tool { return &Tool{locator: locator} }

// Format formats text as the file at path. A .clang-format in the folder of path (or above)
// wins over the default style. A missing clang-format is app.MissingTool("clang-format").
func (t *Tool) Format(path, text string) (string, error) {
	ctx, cancel := context.WithTimeout(context.Background(), formatTimeout)
	defer cancel()
	status := t.locator.Tool(ctx, cpp.ToolClangFormat)
	if status.Source == domain.ToolMissing || status.Path == "" {
		return "", app.MissingTool(cpp.ToolClangFormat)
	}
	dir := folderOf(path)
	cmd := exec.CommandContext(ctx, status.Path, "--assume-filename="+filepath.Base(path), "--style="+styleFor(dir))
	cmd.Dir = dir
	cmd.Stdin = strings.NewReader(text)
	process.HideConsole(cmd)
	var stdout, stderr bytes.Buffer
	cmd.Stdout, cmd.Stderr = &stdout, &stderr
	if err := cmd.Run(); err != nil {
		message := strings.TrimSpace(stderr.String())
		if message == "" {
			message = err.Error()
		}
		return "", fmt.Errorf("%w: %s", app.ErrFormat, message)
	}
	return stdout.String(), nil
}

// styleFor is "file" when a .clang-format or _clang-format is in dir or a parent (clang-format
// searches upwards from the file), and the default style otherwise.
func styleFor(dir string) string {
	for dir != "" {
		for _, name := range []string{".clang-format", "_clang-format"} {
			if _, err := os.Stat(filepath.Join(dir, name)); err == nil {
				return styleFile
			}
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			break
		}
		dir = parent
	}
	return defaultStyle
}

// folderOf is the folder of path, or "" for an untitled file whose folder does not exist.
func folderOf(path string) string {
	dir := filepath.Dir(path)
	if info, err := os.Stat(dir); err != nil || !info.IsDir() {
		return ""
	}
	return dir
}
