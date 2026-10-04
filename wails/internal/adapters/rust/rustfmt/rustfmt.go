// Package rustfmt is the Rust formatter: rustfmt with the text on stdin.
//
// rustfmt refuses code it cannot parse, so a syntax error is an app.ErrFormat that names the
// place, as ruff's does ("main.rs:2:13: expected expression, found `;`").
package rustfmt

import (
	"bytes"
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"
	"time"

	"github.com/codeplai/VizcachaIDE/wails/internal/adapters/rust"
	"github.com/codeplai/VizcachaIDE/wails/internal/app"
	"github.com/codeplai/VizcachaIDE/wails/internal/domain"
	"github.com/codeplai/VizcachaIDE/wails/internal/protocol/process"
)

const formatTimeout = 30 * time.Second

// Tool implements app.CodeFormatter with rustfmt.
type Tool struct{ locator *rust.Locator }

var _ app.CodeFormatter = (*Tool)(nil)

// New creates the formatter.
func New(locator *rust.Locator) *Tool { return &Tool{locator: locator} }

// Format formats text as the file at path. rustfmt runs in the folder of the file, so a
// rustfmt.toml or .rustfmt.toml there or above (a teacher's) wins over its defaults. The edition
// is the one of the crate's Cargo.toml, or the one a loose file compiles with. A missing
// rustfmt is app.MissingTool("rustfmt").
func (t *Tool) Format(path, text string) (string, error) {
	ctx, cancel := context.WithTimeout(context.Background(), formatTimeout)
	defer cancel()
	status := t.locator.Tool(rust.ToolRustfmt)
	if status.Source == domain.ToolMissing || status.Path == "" {
		return "", app.MissingTool(rust.ToolRustfmt)
	}
	cmd := exec.CommandContext(ctx, status.Path, "--emit", "stdout", "--edition", t.editionOf(ctx, path))
	cmd.Dir = folderOf(path)
	cmd.Env = t.locator.Environment()
	cmd.Stdin = strings.NewReader(text)
	process.HideConsole(cmd)
	var stdout, stderr bytes.Buffer
	cmd.Stdout, cmd.Stderr = &stdout, &stderr
	if err := cmd.Run(); err != nil {
		return "", formatError(filepath.Base(path), stderr.String(), err)
	}
	return stdout.String(), nil
}

// editionOf is the edition of the crate around path, or of the toolchain for a loose file.
func (t *Tool) editionOf(ctx context.Context, path string) string {
	if project, found, err := rust.FindProject(path); err == nil && found && project.Edition != "" {
		return project.Edition
	}
	if toolchain, err := t.locator.Toolchain(ctx); err == nil {
		return toolchain.Edition()
	}
	return "2021"
}

// failure is what rustfmt prints for code it cannot parse: the message, then " --> <stdin>:2:13".
var failure = regexp.MustCompile(`(?m)^error[^:\n]*: (.+)\n\s*--> [^\n]*?:(\d+):(\d+)`)

// formatError turns the stderr of rustfmt into "main.rs:2:13: message" wrapping app.ErrFormat.
func formatError(name, stderr string, cause error) error {
	if found := failure.FindStringSubmatch(strings.ReplaceAll(stderr, "\r\n", "\n")); found != nil {
		return fmt.Errorf("%w: %s:%s:%s: %s", app.ErrFormat, name, found[2], found[3], strings.TrimSpace(found[1]))
	}
	message := strings.TrimSpace(stderr)
	if message == "" {
		return fmt.Errorf("%w: %w", app.ErrFormat, cause)
	}
	return fmt.Errorf("%w: %s", app.ErrFormat, message)
}

// folderOf is the folder of path, or "" for an untitled file whose folder does not exist.
func folderOf(path string) string {
	dir := filepath.Dir(path)
	if info, err := os.Stat(dir); err != nil || !info.IsDir() {
		return ""
	}
	return dir
}
