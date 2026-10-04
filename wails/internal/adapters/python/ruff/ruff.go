// Package ruff is the Python formatter and checker: "python -m ruff", always --isolated so a
// ruff configuration of the student's machine never reaches them.
package ruff

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"
	"time"

	"github.com/codeplai/VizcachaIDE/wails/internal/adapters/python"
	"github.com/codeplai/VizcachaIDE/wails/internal/app"
	"github.com/codeplai/VizcachaIDE/wails/internal/domain"
	"github.com/codeplai/VizcachaIDE/wails/internal/protocol/process"
)

const (
	checkTimeout  = 60 * time.Second
	formatTimeout = 30 * time.Second
	// moduleMissing is what "python -m <module>" prints when the module is not installed.
	moduleMissing = "No module named"
)

// Config says where Python is.
type Config struct {
	Locator *python.Locator
	// BaseEnvironment is the "NAME=value" list the tools start from. Nil means os.Environ().
	BaseEnvironment []string
}

// Tool implements app.CodeChecker and app.CodeFormatter with ruff.
type Tool struct{ cfg Config }

var (
	_ app.CodeChecker   = (*Tool)(nil)
	_ app.CodeFormatter = (*Tool)(nil)
)

// New creates the tool.
func New(cfg Config) *Tool {
	if cfg.BaseEnvironment == nil {
		cfg.BaseEnvironment = os.Environ()
	}
	return &Tool{cfg: cfg}
}

// Check runs "ruff check" (syntax errors and pyflakes) on the target and returns its JSON.
// It returns "" when there are no findings or when python or ruff are missing. It emits no
// events and does not use the run slot.
func (t *Tool) Check(ctx context.Context, config domain.RunConfiguration) (string, error) {
	if info, err := os.Stat(config.WorkingDir); err != nil || !info.IsDir() {
		return "", nil // for example an untitled run, whose folder is already deleted
	}
	ctx, cancel := context.WithTimeout(ctx, checkTimeout)
	defer cancel()
	output, stderr, err := t.run(ctx, config.WorkingDir, "", "check", "--isolated", "--select", "E9,F",
		"--output-format", "json", "--no-cache", config.Target)
	if isMissing(err) || strings.Contains(stderr, moduleMissing) {
		return "", nil
	}
	var exit *exec.ExitError
	if err != nil && !errors.As(err, &exit) {
		return "", fmt.Errorf("ruff check: %w", err)
	}
	return output, nil
}

// Format formats text with "ruff format". A syntax error wraps app.ErrFormat and names the
// place as "main.py:5:2: message", which the frontend reads.
func (t *Tool) Format(path, text string) (string, error) {
	ctx, cancel := context.WithTimeout(context.Background(), formatTimeout)
	defer cancel()
	name := filepath.Base(path)
	dir := filepath.Dir(path)
	if info, err := os.Stat(dir); err != nil || !info.IsDir() {
		dir = "" // an untitled file has no folder
	}
	output, stderr, err := t.run(ctx, dir, text, "format", "--isolated", "--stdin-filename", name, "-")
	if isMissing(err) {
		return "", app.MissingTool(python.ToolPython)
	}
	if strings.Contains(stderr, moduleMissing) {
		return "", app.MissingTool(python.ToolRuff)
	}
	if err != nil {
		return "", formatError(name, stderr, err)
	}
	return output, nil
}

var location = regexp.MustCompile(`:(\d+):(\d+)`)

// formatError turns ruff's stderr ("error: Failed to parse main.py:1:7: Expected ...") into
// "main.py:1:7: Expected ..." wrapping app.ErrFormat.
func formatError(name, stderr string, cause error) error {
	message := strings.TrimSpace(stderr)
	if message == "" {
		return fmt.Errorf("%w: %w", app.ErrFormat, cause)
	}
	found := location.FindStringIndex(message)
	if found == nil {
		return fmt.Errorf("%w: %s", app.ErrFormat, message)
	}
	rest := strings.TrimSpace(strings.TrimPrefix(strings.TrimSpace(message[found[1]:]), ":"))
	return fmt.Errorf("%w: %s%s: %s", app.ErrFormat, name, message[found[0]:found[1]], rest)
}

// run starts "python -m ruff args..." in dir with stdin and returns stdout and stderr.
func (t *Tool) run(ctx context.Context, dir, stdin string, args ...string) (string, string, error) {
	interpreter, err := t.cfg.Locator.Find(ctx, dir)
	if err != nil {
		return "", "", err
	}
	cmd := exec.CommandContext(ctx, interpreter.Path, append([]string{"-m", "ruff"}, args...)...)
	cmd.Dir = dir
	cmd.Env = python.EnvironmentList(python.Environment(t.cfg.BaseEnvironment, interpreter))
	cmd.Stdin = strings.NewReader(stdin)
	process.HideConsole(cmd)
	var stdout, stderr bytes.Buffer
	cmd.Stdout, cmd.Stderr = &stdout, &stderr
	err = cmd.Run()
	return stdout.String(), stderr.String(), err
}

// isMissing reports a missing interpreter (app.ErrToolNotFound from the locator).
func isMissing(err error) bool { return errors.Is(err, app.ErrToolNotFound) }
