package runner

import (
	"context"

	"github.com/codeplai/VizcachaIDE/wails/internal/app"
	"github.com/codeplai/VizcachaIDE/wails/internal/domain"
)

// Transitional (M0): the bridge still talks to the 2.0 port app.Toolchain. ToolchainCompat
// satisfies it on top of the Runner; it is a type of its own because Toolchain.RunUntitled has
// another signature than ProgramRunner.RunUntitled. The orchestrator deletes this file when the
// bridge (track N5) uses the language ports.

// ToolchainCompat is the app.Toolchain view of a Runner.
type ToolchainCompat struct {
	*Runner
}

var _ app.Toolchain = ToolchainCompat{}

// Compat returns the app.Toolchain view of the runner.
func (r *Runner) Compat() ToolchainCompat { return ToolchainCompat{Runner: r} }

// Info implements app.Toolchain.
func (c ToolchainCompat) Info(ctx context.Context) domain.ToolchainInfo {
	var info domain.ToolchainInfo
	for _, status := range c.Tools(ctx) {
		switch status.ID {
		case "go":
			info.GoVersion, info.GoSource = status.Version, status.Source
		case "dlv":
			info.DelveVersion, info.DelveSource = status.Version, status.Source
		case "gopls":
			info.GoplsVersion, info.GoplsSource = status.Version, status.Source
		}
	}
	return info
}

// RunUntitled implements app.Toolchain.
func (c ToolchainCompat) RunUntitled(ctx context.Context, source string, programArgs []string) (domain.RunConfiguration, error) {
	return c.Runner.RunUntitled(ctx, "", source, programArgs)
}

// RunGoCommand implements app.Toolchain.
func (c ToolchainCompat) RunGoCommand(ctx context.Context, workingDir string, args []string) error {
	job, err := c.CommandJob(workingDir, args)
	if err != nil {
		return err
	}
	return c.supervisor.Start(ctx, job)
}

// Vet implements app.Toolchain.
func (c ToolchainCompat) Vet(ctx context.Context, config domain.RunConfiguration) (string, error) {
	return c.Check(ctx, config)
}

// FormatSource implements app.Toolchain.
func (c ToolchainCompat) FormatSource(text string) (string, error) { return c.Format("", text) }
