// Package runner is the Python program runner: it implements app.ProgramRunner. It finds the
// interpreter with python.Locator and hands "python -X utf8 file.py" to the shared process
// supervisor, in a pseudoterminal so input() and print() behave like in a console. Checking and
// formatting (ruff) belong to the ruff package.
package runner

import (
	"context"
	"os"

	"github.com/codeplai/VizcachaIDE/wails/internal/adapters/python"
	"github.com/codeplai/VizcachaIDE/wails/internal/app"
	"github.com/codeplai/VizcachaIDE/wails/internal/domain"
	"github.com/codeplai/VizcachaIDE/wails/internal/protocol/process"
)

// Options configures a Runner. All of them are optional.
type Options struct {
	// Settings supplies the python path chosen by the user.
	Settings app.SettingsStore
	// AppDir is the folder of the executable, where toolchain/ lives. Default: the running one.
	AppDir string
	// BaseEnvironment is the "NAME=value" list Python starts from. Default: os.Environ().
	BaseEnvironment []string
	// Locator is shared with the other Python adapters so they agree on the interpreter and
	// probe it once. Default: a new one from Settings, AppDir and BaseEnvironment.
	Locator *python.Locator
}

// Runner is the Python adapter of the run ports.
type Runner struct {
	supervisor *process.Supervisor
	locator    *python.Locator
	base       []string
}

var _ app.ProgramRunner = (*Runner)(nil)

// New creates the runner on the supervisor shared by the whole IDE.
func New(supervisor *process.Supervisor, options Options) *Runner {
	base := options.BaseEnvironment
	if base == nil {
		base = os.Environ()
	}
	locator := options.Locator
	if locator == nil {
		locator = python.NewLocator(python.Options{
			Settings: options.Settings, AppDir: options.AppDir, BaseEnvironment: base,
		})
	}
	return &Runner{supervisor: supervisor, locator: locator, base: base}
}

// Configure runs the file alone from its folder; Python has no project mode.
func (r *Runner) Configure(path string, programArgs []string) domain.RunConfiguration {
	config := domain.NewFileRunConfiguration(domain.CodeLanguagePython, path, programArgs)
	project := &domain.ProjectContext{Root: config.WorkingDir, Kind: domain.ProjectFolder}
	if interpreter, err := r.locator.Find(context.Background(), config.WorkingDir); err == nil && interpreter.Venv {
		project.Name = ".venv"
	}
	config.Project = project
	return config
}

// Run implements app.ProgramRunner.
func (r *Runner) Run(ctx context.Context, config domain.RunConfiguration) error {
	return r.start(ctx, config, nil)
}

// Build implements app.ProgramRunner: Python has no build step.
func (r *Runner) Build(context.Context, domain.RunConfiguration) error { return app.ErrUnsupported }

// Stop implements app.ProgramRunner: it ends the program and everything it started.
func (r *Runner) Stop() error { return r.supervisor.Stop() }

// IsRunning implements app.ProgramRunner.
func (r *Runner) IsRunning() bool { return r.supervisor.IsRunning() }

// WriteInput implements app.ProgramRunner: it types text and Enter into the program.
func (r *Runner) WriteInput(text string) error { return r.supervisor.WriteInput(text) }

// start runs the configuration. cleanup runs when the program ends or cannot start.
func (r *Runner) start(ctx context.Context, config domain.RunConfiguration, cleanup func()) error {
	folder := config.WorkingDir
	if config.Project != nil && config.Project.Root != "" {
		folder = config.Project.Root
	}
	interpreter, err := r.locator.Find(ctx, folder)
	if err != nil {
		if cleanup != nil {
			cleanup()
		}
		return err
	}
	return r.launch(ctx, interpreter, config, cleanup)
}
