// Package packages is the Python package manager: it implements app.PackageManager with pip
// ("python -m pip"), run through the shared process supervisor so the commands emit run events.
// Python has no init or tidy: those verbs return app.ErrUnsupported.
package packages

import (
	"context"
	"fmt"

	"github.com/codeplai/VizcachaIDE/wails/internal/adapters/python/runner"
	"github.com/codeplai/VizcachaIDE/wails/internal/app"
	"github.com/codeplai/VizcachaIDE/wails/internal/protocol/process"
)

// Manager is the Python adapter of app.PackageManager.
type Manager struct {
	supervisor *process.Supervisor
	runner     *runner.Runner
}

var _ app.PackageManager = (*Manager)(nil)

// New creates the package manager. It shares the supervisor with the runner and takes from it
// the interpreter and its environment.
func New(supervisor *process.Supervisor, pythonRunner *runner.Runner) *Manager {
	return &Manager{supervisor: supervisor, runner: pythonRunner}
}

// Init implements app.PackageManager.
func (m *Manager) Init(context.Context, string, string) error { return app.ErrUnsupported }

// Tidy implements app.PackageManager.
func (m *Manager) Tidy(context.Context, string) error { return app.ErrUnsupported }

// Add implements app.PackageManager: "python -m pip install <pkg>".
func (m *Manager) Add(ctx context.Context, dir, pkg string) error {
	name, err := app.SingleWordArgument(pkg, "package")
	if err != nil {
		return err
	}
	return m.run(ctx, dir, []string{"-m", "pip", "install", name})
}

// Remove implements app.PackageManager: "python -m pip uninstall -y <pkg>".
func (m *Manager) Remove(ctx context.Context, dir, pkg string) error {
	name, err := app.SingleWordArgument(pkg, "package")
	if err != nil {
		return err
	}
	return m.run(ctx, dir, []string{"-m", "pip", "uninstall", "-y", name})
}

// List implements app.PackageManager: "python -m pip list".
func (m *Manager) List(ctx context.Context, dir string) error {
	return m.run(ctx, dir, []string{"-m", "pip", "list"})
}

func (m *Manager) run(ctx context.Context, dir string, args []string) error {
	job, err := m.runner.CommandJob(ctx, dir, args)
	if err != nil {
		return err
	}
	if err := m.supervisor.Start(ctx, job); err != nil {
		return fmt.Errorf("pip %v: %w", args[2:], err)
	}
	return nil
}
