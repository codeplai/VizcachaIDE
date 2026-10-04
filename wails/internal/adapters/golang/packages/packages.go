// Package packages is the Go package manager: it implements app.PackageManager with the "go mod"
// and "go get" commands, run through the shared process supervisor so they emit run events.
package packages

import (
	"context"
	"fmt"

	"github.com/codeplai/VizcachaIDE/wails/internal/adapters/golang/runner"
	"github.com/codeplai/VizcachaIDE/wails/internal/app"
	"github.com/codeplai/VizcachaIDE/wails/internal/protocol/process"
)

// Manager is the Go adapter of app.PackageManager.
type Manager struct {
	supervisor *process.Supervisor
	runner     *runner.Runner
}

var _ app.PackageManager = (*Manager)(nil)

// New creates the package manager. It shares the supervisor with the runner, and takes from it
// the location of the go tool and its environment.
func New(supervisor *process.Supervisor, goRunner *runner.Runner) *Manager {
	return &Manager{supervisor: supervisor, runner: goRunner}
}

// Init implements app.PackageManager: "go mod init <name>".
func (m *Manager) Init(ctx context.Context, dir, name string) error {
	args, err := modInitArguments(name)
	if err != nil {
		return err
	}
	return m.run(ctx, dir, args)
}

// Add implements app.PackageManager: "go get <pkg>".
func (m *Manager) Add(ctx context.Context, dir, pkg string) error {
	args, err := getArguments(pkg)
	if err != nil {
		return err
	}
	return m.run(ctx, dir, args)
}

// Tidy implements app.PackageManager: "go mod tidy".
func (m *Manager) Tidy(ctx context.Context, dir string) error {
	return m.run(ctx, dir, modTidyArguments())
}

// Remove implements app.PackageManager. Go drops a dependency by tidying after deleting its import.
func (m *Manager) Remove(context.Context, string, string) error { return app.ErrUnsupported }

// List implements app.PackageManager.
func (m *Manager) List(context.Context, string) error { return app.ErrUnsupported }

func (m *Manager) run(ctx context.Context, dir string, args []string) error {
	job, err := m.runner.CommandJob(dir, args)
	if err != nil {
		return err
	}
	if err := m.supervisor.Start(ctx, job); err != nil {
		return fmt.Errorf("go %v: %w", args, err)
	}
	return nil
}
