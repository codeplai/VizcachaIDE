package main

import (
	"context"
	"path/filepath"

	"github.com/codeplai/VizcachaIDE/wails/internal/adapters/cpp"
	"github.com/codeplai/VizcachaIDE/wails/internal/adapters/cpp/cmake"
	"github.com/codeplai/VizcachaIDE/wails/internal/adapters/cpp/runner"
	"github.com/codeplai/VizcachaIDE/wails/internal/adapters/cpp/vcpkg"
	"github.com/codeplai/VizcachaIDE/wails/internal/app"
	"github.com/codeplai/VizcachaIDE/wails/internal/domain"
	"github.com/codeplai/VizcachaIDE/wails/internal/protocol/process"
)

// The libraries are installed by the package manager while the supervisor is free, so the two
// share one slot by hand: a run or a build is refused while a library installs (cppSlot), and the
// manager refuses while a program runs (cppPackages).
var (
	_ cmake.Dependencies = (*vcpkg.Setup)(nil)
	_ vcpkg.Configurer   = (*cmake.Builder)(nil)
)

// newVcpkgLocator finds vcpkg with the compiler, CMake and Ninja the C++ locator found: vcpkg must
// use the IDE's cmake instead of downloading its own.
func newVcpkgLocator(store app.SettingsStore, base []string, locator *cpp.Locator) *vcpkg.Locator {
	folder := func(id string) string {
		if status := locator.Tool(context.Background(), id); status.Source != domain.ToolMissing {
			return filepath.Dir(status.Path)
		}
		return ""
	}
	return vcpkg.NewLocator(vcpkg.LocatorOptions{
		Settings: store, BaseEnvironment: base,
		CompilerBin: func() string {
			if compiler, err := locator.Compiler(context.Background()); err == nil {
				return filepath.Dir(compiler.Path)
			}
			return ""
		},
		ToolBins: func() []string { return []string{folder(cpp.ToolCMake), folder(cpp.ToolNinja)} },
	})
}

// lazyConfigurer configures with the CMake builder of the runner, which exists only when the
// tools are installed: a missing one is the error of the operation, not of the start-up.
type lazyConfigurer struct{ runner *runner.Runner }

func (c lazyConfigurer) Configure(ctx context.Context, root string) (string, error) {
	builder, err := c.runner.Builder(ctx)
	if err != nil {
		return "", err
	}
	return builder.Configure(ctx, root)
}

// cppSlot is the runner of C++ that refuses to start while a library installs, and whose Stop
// also stops that installation.
type cppSlot struct {
	*runner.Runner
	manager *vcpkg.Manager
}

func (s cppSlot) Run(ctx context.Context, config domain.RunConfiguration) error {
	if s.manager.Busy() {
		return app.ErrBusy
	}
	return s.Runner.Run(ctx, config)
}

func (s cppSlot) Build(ctx context.Context, config domain.RunConfiguration) error {
	if s.manager.Busy() {
		return app.ErrBusy
	}
	return s.Runner.Build(ctx, config)
}

func (s cppSlot) RunUntitled(ctx context.Context, path, source string, programArgs []string) (domain.RunConfiguration, error) {
	if s.manager.Busy() {
		return domain.RunConfiguration{}, app.ErrBusy
	}
	return s.Runner.RunUntitled(ctx, path, source, programArgs)
}

func (s cppSlot) Stop() error {
	s.manager.Stop()
	return s.Runner.Stop()
}

func (s cppSlot) IsRunning() bool { return s.Runner.IsRunning() || s.manager.Busy() }

// cppPackages is the package manager of C++ that refuses to add or remove a library while a
// program runs or builds.
type cppPackages struct {
	*vcpkg.Manager
	supervisor *process.Supervisor
}

func (p cppPackages) Add(ctx context.Context, dir, pkg string) error {
	if p.supervisor.IsRunning() {
		return app.ErrBusy
	}
	return p.Manager.Add(ctx, dir, pkg)
}

func (p cppPackages) Remove(ctx context.Context, dir, pkg string) error {
	if p.supervisor.IsRunning() {
		return app.ErrBusy
	}
	return p.Manager.Remove(ctx, dir, pkg)
}

// cppShell adds to the terminal what "cmake --preset debug" needs: the tools' folders (the C++
// locator) and VCPKG_ROOT.
type cppShell struct {
	*cpp.Locator
	libraries *vcpkg.Locator
}

// ShellVariables implements app.ShellVariables.
func (s cppShell) ShellVariables(context.Context) []string {
	if root := s.libraries.Root(); root != "" {
		return []string{"VCPKG_ROOT=" + root}
	}
	return nil
}
