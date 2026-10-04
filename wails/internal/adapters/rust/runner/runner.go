// Package runner is the Rust program runner: it implements app.ProgramRunner. Run is one
// two-stage job on the shared process supervisor: rustc or cargo (pipes, JSON diagnostics turned
// into their rendered text) builds the executable and the program then runs in a pseudoterminal
// so read_line and println! behave like in a console. The checker (clippy), the formatter
// (rustfmt) and the debugger (lldbdap, which compiles with CompileForDebug) are other packages.
package runner

import (
	"context"
	"os"
	"path/filepath"
	"sync"
	"time"

	"github.com/codeplai/VizcachaIDE/wails/internal/adapters/rust"
	"github.com/codeplai/VizcachaIDE/wails/internal/app"
	"github.com/codeplai/VizcachaIDE/wails/internal/protocol/process"
)

const defaultNoticeDelay = 2 * time.Second

// Options configures a Runner. All of them are optional.
type Options struct {
	// Settings supplies the tool paths chosen by the user.
	Settings app.SettingsStore
	// AppDir is the folder of the executable, where toolchain/ lives. Default: the running one.
	AppDir string
	// BaseEnvironment is the "NAME=value" list the tools and programs start from. Default:
	// os.Environ().
	BaseEnvironment []string
	// Locator is shared with the other Rust adapters so they agree on the toolchain and probe it
	// once. Default: a new one from Settings, AppDir and BaseEnvironment.
	Locator *rust.Locator
	// CompilingNotice returns the already translated "Compiling..." text (key run.compiling). It
	// is printed to stdout when the compiler is silent for NoticeDelay.
	CompilingNotice func() string
	// NoticeDelay defaults to 2 seconds.
	NoticeDelay time.Duration
	// CacheDir is the folder that holds VizcachaIDE/build. Default: os.UserCacheDir().
	CacheDir string
}

// Runner is the Rust adapter of app.ProgramRunner.
type Runner struct {
	supervisor *process.Supervisor
	locator    *rust.Locator
	base       []string
	options    Options
	// terminal reports whether a pseudoterminal can start (replaced in tests).
	terminal    func(ctx context.Context, program string) bool
	probe       sync.Once
	hasTerminal bool
}

var _ app.ProgramRunner = (*Runner)(nil)

// New creates the runner on the supervisor shared by the whole IDE.
func New(supervisor *process.Supervisor, options Options) *Runner {
	base := options.BaseEnvironment
	if base == nil {
		base = os.Environ()
	}
	if options.NoticeDelay <= 0 {
		options.NoticeDelay = defaultNoticeDelay
	}
	if options.AppDir == "" {
		options.AppDir = executableFolder()
	}
	locator := options.Locator
	if locator == nil {
		locator = rust.NewLocator(rust.Options{Settings: options.Settings, AppDir: options.AppDir, BaseEnvironment: base})
	}
	return &Runner{supervisor: supervisor, locator: locator, base: base, options: options, terminal: terminalWorks}
}

// Stop implements app.ProgramRunner: it ends the program and everything it started.
func (r *Runner) Stop() error { return r.supervisor.Stop() }

// IsRunning implements app.ProgramRunner.
func (r *Runner) IsRunning() bool { return r.supervisor.IsRunning() }

// WriteInput implements app.ProgramRunner: it types text and Enter into the program.
func (r *Runner) WriteInput(text string) error { return r.supervisor.WriteInput(text) }

// Environment implements app.ProgramRunner: what every Rust process runs with (plain output,
// RUST_BACKTRACE=1, cargo's folder first in PATH). The debugger launches the program with it.
func (r *Runner) Environment() map[string]string { return environmentMap(r.locator.Environment()) }

// Toolchain reads the rustc in use (app.MissingTool("rustc") when there is none).
func (r *Runner) Toolchain(ctx context.Context) (rust.Toolchain, error) {
	return r.locator.Toolchain(ctx)
}

func environmentMap(list []string) map[string]string {
	env := make(map[string]string, len(list))
	for _, entry := range list {
		for i := 1; i < len(entry); i++ { // from 1: Windows has hidden "=C:=..." entries
			if entry[i] == '=' {
				env[entry[:i]] = entry[i+1:]
				break
			}
		}
	}
	return env
}

func executableFolder() string {
	executable, err := os.Executable()
	if err != nil {
		return "."
	}
	return filepath.Dir(executable)
}
