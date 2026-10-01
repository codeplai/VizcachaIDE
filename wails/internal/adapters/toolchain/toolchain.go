// Package toolchain implements app.Toolchain: it locates Go (configured path,
// bundled next to the executable, PATH), runs the user's program with os/exec,
// streams its output through the EventSink and formats code with go/format.
package toolchain

import (
	"context"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/codeplai/VizcachaIDE/wails/internal/app"
	"github.com/codeplai/VizcachaIDE/wails/internal/domain"
)

const defaultFirstBuildDelay = 5 * time.Second

// Options configures a Toolchain. Only Sink is required.
type Options struct {
	Sink app.EventSink
	// Settings supplies the tool paths chosen by the user. May be nil.
	Settings app.SettingsStore
	// AppDir is the folder of the executable, where toolchain/ lives. Default: the running executable's folder.
	AppDir string
	// BaseEnvironment is the "NAME=value" list the tools start from. Default: os.Environ().
	BaseEnvironment []string
	// FirstBuildNotice returns the already translated "preparing Go for the first
	// time" text. It is printed to stdout when a run is silent for FirstBuildDelay.
	FirstBuildNotice func() string
	// FirstBuildDelay defaults to 5 seconds.
	FirstBuildDelay time.Duration
}

// Toolchain is the app.Toolchain adapter.
type Toolchain struct {
	sink     app.EventSink
	settings app.SettingsStore
	base     []string
	locator  *Locator
	notice   func() string
	delay    time.Duration

	mu      sync.Mutex
	current *process
}

var _ app.Toolchain = (*Toolchain)(nil)

// New creates the adapter.
func New(options Options) *Toolchain {
	t := &Toolchain{
		sink:     options.Sink,
		settings: options.Settings,
		base:     options.BaseEnvironment,
		notice:   options.FirstBuildNotice,
		delay:    options.FirstBuildDelay,
	}
	if t.base == nil {
		t.base = os.Environ()
	}
	if t.delay <= 0 {
		t.delay = defaultFirstBuildDelay
	}
	searchPath := parseEnvironment(t.base)[pathVariableName(parseEnvironment(t.base))]
	t.locator = NewLocator(applicationDirectory(options.AppDir), searchPath, t.configuredPath)
	return t
}

// applicationDirectory is the folder where toolchain/ is expected.
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

// configuredPath is the path the user chose in the settings for a tool, or "".
func (t *Toolchain) configuredPath(tool string) string {
	if t.settings == nil {
		return ""
	}
	settings, err := t.settings.Load()
	if err != nil {
		return ""
	}
	switch tool {
	case ToolGo:
		return strings.TrimSpace(settings.GoPath)
	case ToolDelve:
		return strings.TrimSpace(settings.DelvePath)
	case ToolGopls:
		return strings.TrimSpace(settings.GoplsPath)
	}
	return ""
}

// Locate reports where a tool ("go", "dlv" or "gopls") comes from.
func (t *Toolchain) Locate(tool string) Location { return t.locator.Locate(tool) }

// Environment implements app.Toolchain.
func (t *Toolchain) Environment() map[string]string {
	return buildEnvironment(t.base, t.locator)
}

// IsRunning implements app.Toolchain.
func (t *Toolchain) IsRunning() bool {
	t.mu.Lock()
	defer t.mu.Unlock()
	return t.current != nil
}

// Stop implements app.Toolchain: it ends the program and everything it started.
func (t *Toolchain) Stop() error {
	t.mu.Lock()
	proc := t.current
	t.mu.Unlock()
	if proc == nil {
		return nil
	}
	terminateTree(proc.cmd.Process.Pid, proc.done)
	return nil
}

// WriteInput implements app.Toolchain: it types text and Enter into the program.
func (t *Toolchain) WriteInput(text string) error {
	t.mu.Lock()
	proc := t.current
	t.mu.Unlock()
	if proc == nil {
		return nil
	}
	line := strings.TrimRight(text, "\r\n") + "\n"
	if _, err := io.WriteString(proc.stdin, line); err != nil {
		return fmt.Errorf("write to the program: %w", err)
	}
	return nil
}

// RunUntitled implements app.Toolchain: it runs unsaved source from a temporary
// folder that is deleted when the program ends.
func (t *Toolchain) RunUntitled(ctx context.Context, source string, programArgs []string) (domain.RunConfiguration, error) {
	if t.IsRunning() {
		return domain.RunConfiguration{}, app.ErrBusy
	}
	dir, err := os.MkdirTemp("", "vizcacha-untitled-")
	if err != nil {
		return domain.RunConfiguration{}, fmt.Errorf("create temporary folder: %w", err)
	}
	cleanup := func() { _ = os.RemoveAll(dir) }
	path := filepath.Join(dir, "main.go")
	if err := os.WriteFile(path, []byte(source), 0o600); err != nil {
		cleanup()
		return domain.RunConfiguration{}, fmt.Errorf("save the untitled file: %w", err)
	}
	config := domain.NewFileRunConfiguration(path, programArgs)
	err = t.launch(ctx, launch{config: config, args: runArguments(config), cleanup: cleanup})
	return config, err
}
