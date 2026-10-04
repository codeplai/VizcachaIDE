// Package runner is the Go program runner: it implements app.ProgramRunner, app.CodeChecker and
// app.CodeFormatter. It locates the Go tools (configured path, bundled next to the executable,
// PATH), builds the "go run" and "go build" jobs and hands them to the shared process
// supervisor, which streams the output through the EventSink and owns the single run slot.
package runner

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/codeplai/VizcachaIDE/wails/internal/adapters/golang"
	"github.com/codeplai/VizcachaIDE/wails/internal/app"
	"github.com/codeplai/VizcachaIDE/wails/internal/domain"
	"github.com/codeplai/VizcachaIDE/wails/internal/protocol/process"
	"github.com/codeplai/VizcachaIDE/wails/internal/protocol/toollocator"
)

const defaultFirstBuildDelay = 5 * time.Second

// Options configures a Runner. All of them are optional.
type Options struct {
	// Settings supplies the tool paths chosen by the user.
	Settings app.SettingsStore
	// AppDir is the folder of the executable, where toolchain/ lives. Default: the running
	// executable's folder.
	AppDir string
	// BaseEnvironment is the "NAME=value" list the tools start from. Default: os.Environ().
	BaseEnvironment []string
	// FirstBuildNotice returns the already translated "preparing Go for the first time" text.
	// It is printed to stdout when a run is silent for FirstBuildDelay.
	FirstBuildNotice func() string
	// FirstBuildDelay defaults to 5 seconds.
	FirstBuildDelay time.Duration
}

// Runner is the Go adapter of the run ports.
type Runner struct {
	supervisor *process.Supervisor
	settings   app.SettingsStore
	base       []string
	locator    *toollocator.Locator
	notice     func() string
	delay      time.Duration
}

var (
	_ app.ProgramRunner = (*Runner)(nil)
	_ app.CodeChecker   = (*Runner)(nil)
	_ app.CodeFormatter = (*Runner)(nil)
)

// New creates the runner on the supervisor shared by the whole IDE.
func New(supervisor *process.Supervisor, options Options) *Runner {
	r := &Runner{
		supervisor: supervisor,
		settings:   options.Settings,
		base:       options.BaseEnvironment,
		notice:     options.FirstBuildNotice,
		delay:      options.FirstBuildDelay,
	}
	if r.base == nil {
		r.base = os.Environ()
	}
	if r.delay <= 0 {
		r.delay = defaultFirstBuildDelay
	}
	env := parseEnvironment(r.base)
	r.locator = toollocator.New(applicationDirectory(options.AppDir), env[pathVariableName(env)], r.configuredPath)
	return r
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
func (r *Runner) configuredPath(toolID string) string {
	if r.settings == nil {
		return ""
	}
	settings, err := r.settings.Load()
	if err != nil {
		return ""
	}
	return strings.TrimSpace(settings.ToolPaths[toolID])
}

// Configure runs the file alone, or its module when the file lives inside one.
func (r *Runner) Configure(path string, programArgs []string) domain.RunConfiguration {
	return golang.ConfigurationForFile(path, programArgs)
}

// Run implements app.ProgramRunner: "go run".
func (r *Runner) Run(ctx context.Context, config domain.RunConfiguration) error {
	return r.start(ctx, config, runArguments(config), nil)
}

// Build implements app.ProgramRunner: "go build".
func (r *Runner) Build(ctx context.Context, config domain.RunConfiguration) error {
	output := golang.GoExecutableName(config, isWindows())
	return r.start(ctx, config, []string{"build", "-o", output, golang.GoTargetArgument(config)}, nil)
}

// RunUntitled implements app.ProgramRunner: it runs unsaved source from a temporary folder
// that is deleted when the program ends. Go needs a main package, so the file is main.go.
func (r *Runner) RunUntitled(ctx context.Context, _, source string, programArgs []string) (domain.RunConfiguration, error) {
	if r.supervisor.IsRunning() {
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
	config := domain.NewFileRunConfiguration(domain.CodeLanguageGo, path, programArgs)
	return config, r.start(ctx, config, runArguments(config), cleanup)
}

// Stop implements app.ProgramRunner: it ends the program and everything it started.
func (r *Runner) Stop() error { return r.supervisor.Stop() }

// IsRunning implements app.ProgramRunner.
func (r *Runner) IsRunning() bool { return r.supervisor.IsRunning() }

// WriteInput implements app.ProgramRunner: it types text and Enter into the program.
func (r *Runner) WriteInput(text string) error { return r.supervisor.WriteInput(text) }

// Environment implements app.ProgramRunner.
func (r *Runner) Environment() map[string]string { return buildEnvironment(r.base, r.locator) }

func runArguments(config domain.RunConfiguration) []string {
	return append([]string{"run", golang.GoTargetArgument(config)}, config.ProgramArgs...)
}

// start runs a go command for the configuration. Cleanup runs when it ends or fails to start.
func (r *Runner) start(ctx context.Context, config domain.RunConfiguration, args []string, cleanup func()) error {
	job, err := r.job(config, args, cleanup)
	if err != nil {
		if cleanup != nil {
			cleanup()
		}
		return err
	}
	return r.supervisor.Start(ctx, job)
}

// job describes "go <args>" for the supervisor.
func (r *Runner) job(config domain.RunConfiguration, args []string, cleanup func()) (process.Job, error) {
	goTool := r.Locate("go")
	if goTool.Source == domain.ToolMissing {
		return process.Job{}, app.MissingTool("go")
	}
	job := process.Job{
		Config: config, Command: goTool.Path, Args: args, Dir: config.WorkingDir,
		Env: r.Environment(), Mode: process.Pipes, Cleanup: cleanup,
	}
	if r.notice != nil {
		job.Notice = &process.SilentNotice{Text: r.notice, Delay: r.delay}
	}
	return job, nil
}

// CommandJob describes "go <args>" run in workingDir as a project command ("go mod tidy").
// The package manager starts it on the shared supervisor.
func (r *Runner) CommandJob(workingDir string, args []string) (process.Job, error) {
	config := domain.RunConfiguration{
		CodeLanguage: domain.CodeLanguageGo,
		Target:       strings.Join(append([]string{"go"}, args...), " "),
		WorkingDir:   workingDir,
		Mode:         domain.RunProject,
		ProgramArgs:  []string{},
	}
	return r.job(config, args, nil)
}
