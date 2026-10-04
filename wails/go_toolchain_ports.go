package main

import (
	"context"
	"fmt"

	"github.com/codeplai/VizcachaIDE/wails/internal/adapters/toolchain"
	"github.com/codeplai/VizcachaIDE/wails/internal/app"
	"github.com/codeplai/VizcachaIDE/wails/internal/domain"
)

// goToolchainPorts exposes the 2.0 Go toolchain adapter (app.Toolchain) through the ports of
// contract v3: ProgramRunner, CodeFormatter, CodeChecker and PackageManager.
//
// Transitional (M0): track N3 replaces the adapter with a runner built on protocol/process that
// implements these ports itself; the integration deletes this file and the adapters/toolchain
// folder, and newGoSupport uses that runner.
type goToolchainPorts struct {
	toolchain *toolchain.Toolchain
}

var (
	_ app.ProgramRunner  = (*goToolchainPorts)(nil)
	_ app.CodeFormatter  = (*goToolchainPorts)(nil)
	_ app.CodeChecker    = (*goToolchainPorts)(nil)
	_ app.PackageManager = (*goToolchainPorts)(nil)
)

func newGoToolchainPorts(tc *toolchain.Toolchain) *goToolchainPorts {
	return &goToolchainPorts{toolchain: tc}
}

// Configure runs "go run ." inside a module and "go run file.go" outside.
func (p *goToolchainPorts) Configure(path string, programArgs []string) domain.RunConfiguration {
	return app.ConfigurationForFile(path, programArgs)
}

func (p *goToolchainPorts) Run(ctx context.Context, config domain.RunConfiguration) error {
	return p.toolchain.Run(ctx, config)
}

func (p *goToolchainPorts) Build(ctx context.Context, config domain.RunConfiguration) error {
	return p.toolchain.Build(ctx, config)
}

// RunUntitled ignores the extension of path: this adapter only knows Go.
func (p *goToolchainPorts) RunUntitled(ctx context.Context, _, source string, programArgs []string) (domain.RunConfiguration, error) {
	return p.toolchain.RunUntitled(ctx, source, programArgs)
}

func (p *goToolchainPorts) Stop() error                  { return p.toolchain.Stop() }
func (p *goToolchainPorts) IsRunning() bool              { return p.toolchain.IsRunning() }
func (p *goToolchainPorts) WriteInput(text string) error { return p.toolchain.WriteInput(text) }

func (p *goToolchainPorts) Environment() map[string]string { return p.toolchain.Environment() }

// Tools reports go, dlv and gopls the way contract v3 wants them.
func (p *goToolchainPorts) Tools(ctx context.Context) []domain.ToolStatus {
	info := p.toolchain.Info(ctx)
	status := func(id string, role domain.ToolRole, version string, source domain.ToolSource) domain.ToolStatus {
		return domain.ToolStatus{
			ID: id, CodeLanguage: domain.CodeLanguageGo, Role: role,
			Version: version, Source: source, Path: p.toolchain.Locate(id).Path,
		}
	}
	return []domain.ToolStatus{
		status(toolchain.ToolGo, domain.RoleRuntime, info.GoVersion, info.GoSource),
		status(toolchain.ToolDelve, domain.RoleDebugAdapter, info.DelveVersion, info.DelveSource),
		status(toolchain.ToolGopls, domain.RoleLanguageServer, info.GoplsVersion, info.GoplsSource),
	}
}

func (p *goToolchainPorts) Format(_, text string) (string, error) {
	return p.toolchain.FormatSource(text)
}

func (p *goToolchainPorts) Check(ctx context.Context, config domain.RunConfiguration) (string, error) {
	return p.toolchain.Vet(ctx, config)
}

func (p *goToolchainPorts) Init(ctx context.Context, dir, name string) error {
	args, err := app.ModInitArguments(name)
	if err != nil {
		return err
	}
	return p.goCommand(ctx, dir, args)
}

func (p *goToolchainPorts) Add(ctx context.Context, dir, pkg string) error {
	args, err := app.GetArguments(pkg)
	if err != nil {
		return err
	}
	return p.goCommand(ctx, dir, args)
}

func (p *goToolchainPorts) Tidy(ctx context.Context, dir string) error {
	return p.goCommand(ctx, dir, app.ModTidyArguments())
}

func (p *goToolchainPorts) Remove(context.Context, string, string) error { return app.ErrUnsupported }
func (p *goToolchainPorts) List(context.Context, string) error           { return app.ErrUnsupported }

func (p *goToolchainPorts) goCommand(ctx context.Context, dir string, args []string) error {
	if err := p.toolchain.RunGoCommand(ctx, dir, args); err != nil {
		return fmt.Errorf("go %v: %w", args, err)
	}
	return nil
}
