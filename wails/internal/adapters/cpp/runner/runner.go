// Package runner is the C++ program runner: it implements app.ProgramRunner and app.CodeChecker.
// Run is one two-stage job on the shared process supervisor: the compiler (pipes) writes the
// executable into the build cache and the program then runs in a pseudoterminal so std::cin and
// std::cout behave like in a console. Formatting (clang-format) belongs to the clangformat
// package and the debugger to lldbdap, which compiles with CompileForDebug.
package runner

import (
	"context"
	"os"
	"sync"
	"time"

	"github.com/codeplai/VizcachaIDE/wails/internal/adapters/cpp"
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
	// Locator is shared with the other C++ adapters so they agree on the compiler and probe it
	// once. Default: a new one from Settings, AppDir and BaseEnvironment.
	Locator *cpp.Locator
	// CompilingNotice returns the already translated "Compiling..." text (key run.compiling). It
	// is printed to stdout when the compiler is silent for NoticeDelay.
	CompilingNotice func() string
	// NoticeDelay defaults to 2 seconds.
	NoticeDelay time.Duration
	// CacheDir is the folder that holds VizcachaIDE/build. Default: os.UserCacheDir().
	CacheDir string
}

// Runner is the C++ adapter of the run ports.
type Runner struct {
	supervisor *process.Supervisor
	locator    *cpp.Locator
	base       []string
	options    Options
	// terminal reports whether a pseudoterminal can start (replaced in tests).
	terminal    func(ctx context.Context, compiler string) bool
	probe       sync.Once
	hasTerminal bool
}

var (
	_ app.ProgramRunner = (*Runner)(nil)
	_ app.CodeChecker   = (*Runner)(nil)
)

// New creates the runner on the supervisor shared by the whole IDE.
func New(supervisor *process.Supervisor, options Options) *Runner {
	base := options.BaseEnvironment
	if base == nil {
		base = os.Environ()
	}
	if options.NoticeDelay <= 0 {
		options.NoticeDelay = defaultNoticeDelay
	}
	locator := options.Locator
	if locator == nil {
		locator = cpp.NewLocator(cpp.Options{Settings: options.Settings, AppDir: options.AppDir, BaseEnvironment: base})
	}
	return &Runner{supervisor: supervisor, locator: locator, base: base, options: options, terminal: terminalWorks}
}

// Stop implements app.ProgramRunner: it ends the program and everything it started.
func (r *Runner) Stop() error { return r.supervisor.Stop() }

// IsRunning implements app.ProgramRunner.
func (r *Runner) IsRunning() bool { return r.supervisor.IsRunning() }

// WriteInput implements app.ProgramRunner: it types text and Enter into the program.
func (r *Runner) WriteInput(text string) error { return r.supervisor.WriteInput(text) }

// Environment implements app.ProgramRunner: the base environment. Windows executables are linked
// statically (cpp.CompileFlags), so the compiler's folder is not needed on PATH.
func (r *Runner) Environment() map[string]string { return environmentMap(r.base) }

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

// Compiler finds the C++ compiler (app.MissingTool("cxx") when there is none).
func (r *Runner) Compiler(ctx context.Context) (cpp.Compiler, error) {
	return r.locator.Compiler(ctx)
}
