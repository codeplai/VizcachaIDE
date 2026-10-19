package runner

import (
	"context"
	"fmt"

	"github.com/codeplai/VizcachaIDE/wails/internal/adapters/rust"
	"github.com/codeplai/VizcachaIDE/wails/internal/app"
	"github.com/codeplai/VizcachaIDE/wails/internal/domain"
)

var _ app.ProjectTester = (*Runner)(nil)

// ErrNoCargoProject means the file is not inside a Cargo crate, so there is nothing for "cargo
// test" to run. It wraps app.ErrUnsupported and carries the key errors.rustNoProject.
var ErrNoCargoProject = fmt.Errorf("%w: errors.rustNoProject", app.ErrUnsupported)

// Test implements app.ProjectTester: "cargo test" of the crate around path, or of the whole
// workspace when path is in the root of a virtual one. It runs on the shared supervisor, so it
// emits the run events and its output is the run's.
func (r *Runner) Test(ctx context.Context, path string) (domain.RunConfiguration, error) {
	project, ok, err := rust.FindProject(anchor(path))
	if err != nil {
		return domain.RunConfiguration{}, fmt.Errorf("read Cargo.toml: %w", err)
	}
	if !ok {
		return domain.RunConfiguration{}, ErrNoCargoProject
	}
	toolchain, err := r.locator.Toolchain(ctx)
	if err != nil {
		return domain.RunConfiguration{}, err
	}
	args := []string{"test", "--manifest-path", project.Manifest}
	if project.Virtual {
		args = append(args, "--workspace")
	}
	job, err := r.CommandJob(project.Root, args)
	if err != nil {
		return domain.RunConfiguration{}, err
	}
	job.Config.Target = "cargo test" // what the console header shows, not the manifest path
	if linker := r.linkerFor(toolchain); linker != "" {
		job.Env[cargoLinkerVariable(toolchain.Host)] = linker
	}
	if err := r.supervisor.Start(ctx, job); err != nil {
		return job.Config, err
	}
	return job.Config, nil
}
