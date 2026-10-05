package runner

import (
	"context"
	"fmt"
	"io"
	"os"
	"path/filepath"

	"github.com/codeplai/VizcachaIDE/wails/internal/adapters/cpp/cmake"
	"github.com/codeplai/VizcachaIDE/wails/internal/domain"
	"github.com/codeplai/VizcachaIDE/wails/internal/protocol/process"
)

// noProgram is what the run prints when the project builds but declares no executable. The
// program stage never starts, so the text is not translated by the frontend: both languages.
const noProgram = "The project has no program to run: add an add_executable(...) to CMakeLists.txt.\n" +
	"El proyecto no tiene un programa para ejecutar: agrega un add_executable(...) en CMakeLists.txt."

// cmakeBuild is one run of a CMake project on the supervisor: configure (when the build folder
// is stale), build, then what follows the build.
type cmakeBuild struct {
	runner  *Runner
	tools   cmakeTools
	plan    cmake.Plan
	project cmake.Project
	config  domain.RunConfiguration
	filter  *cmake.Filter
}

func (r *Runner) newBuild(ctx context.Context, config domain.RunConfiguration, project cmake.Project, release bool) (cmakeBuild, error) {
	tools, err := r.cmakeTools(ctx)
	if err != nil {
		return cmakeBuild{}, err
	}
	plan, err := tools.builder.Plan(ctx, project.Root, release)
	if err != nil {
		return cmakeBuild{}, err
	}
	return cmakeBuild{runner: r, tools: tools, plan: plan, project: project, config: config, filter: r.filter()}, nil
}

// runProject implements Run for a CMake project: configure, build, run the program.
func (r *Runner) runProject(ctx context.Context, config domain.RunConfiguration, project cmake.Project) error {
	build, err := r.newBuild(ctx, config, project, false)
	if err != nil {
		return err
	}
	mode := process.Pipes
	if r.terminalWorks(ctx, build.tools.compiler.Path) {
		mode = process.Terminal
	}
	config.Echo = mode == process.Terminal
	build.config = config
	return r.supervisor.Start(ctx, build.first(func() (*process.Job, bool) { return build.program(mode) }, nil))
}

// buildProject implements Build for a CMake project: a Release build, then the program is
// copied beside the code.
func (r *Runner) buildProject(ctx context.Context, config domain.RunConfiguration, project cmake.Project) error {
	build, err := r.newBuild(ctx, config, project, true)
	if err != nil {
		return err
	}
	return r.supervisor.Start(ctx, build.first(nil, build.copyBesideCode))
}

// first is the first stage: the configure command when the build folder needs it, else the build.
// afterBuild is the stage that follows a successful build (nil for none); finished is the text
// the last build stage prints when it ends (nil for none).
func (b cmakeBuild) first(afterBuild func() (*process.Job, bool), finished func(int) string) process.Job {
	built := b.job(b.plan.Build, b.filter.BuildLine)
	built.Finished = finished
	if afterBuild != nil {
		built.Then = func(exitCode int) (*process.Job, bool) {
			if exitCode != 0 {
				return nil, false
			}
			return afterBuild()
		}
	}
	if b.plan.Configure == nil {
		return built
	}
	configure := b.job(*b.plan.Configure, b.filter.Line)
	configure.Then = func(exitCode int) (*process.Job, bool) {
		return &built, exitCode == 0
	}
	return configure
}

// job is a CMake or Ninja command on the supervisor, its output filtered.
func (b cmakeBuild) job(command cmake.Command, filter process.OutputFilter) process.Job {
	return process.Job{
		Config: b.config, Command: command.Path, Args: command.Args, Dir: command.Dir,
		Env: environmentMap(command.Env), Mode: process.Pipes, OutputFilter: filter,
	}
}

// program is the last stage of a run: the executable CMake built, with the program's arguments.
func (b cmakeBuild) program(mode process.Mode) (*process.Job, bool) {
	executable, err := b.executable()
	if err != nil {
		return b.failure(noProgram), true
	}
	job := programJob(executable, b.project.Root, b.config, b.runner.Environment(), mode)
	return &job, true
}

// executable is the program of the build folder: the one with the active file, else the
// project's own, else the first.
func (b cmakeBuild) executable() (string, error) {
	executables, err := cmake.Executables(b.plan.BuildDir)
	if err != nil {
		return "", err
	}
	chosen, ok := cmake.Choose(executables, b.runner.activeFile(b.project.Root), b.project.Target)
	if !ok {
		return "", cmake.ErrNoExecutable
	}
	return chosen.Path, nil
}

// failure is a stage that only fails with a text: "cmake -E false" with the text as closing line.
func (b cmakeBuild) failure(text string) *process.Job {
	return &process.Job{
		Config: b.config, Command: b.plan.Build.Path, Args: []string{"-E", "false"}, Dir: b.project.Root,
		Env: environmentMap(b.plan.Build.Env), Mode: process.Pipes,
		Finished: func(int) string { return text },
	}
}

// copyBesideCode is the closing line of Build: it copies the Release program beside the code, like
// "go build". The text is an error, or "" when the copy worked.
func (b cmakeBuild) copyBesideCode(exitCode int) string {
	if exitCode != 0 {
		return ""
	}
	source, err := b.executable()
	if err != nil {
		return noProgram
	}
	if err := copyFile(source, filepath.Join(b.project.Root, filepath.Base(source))); err != nil {
		return err.Error()
	}
	return ""
}

func copyFile(from, to string) error {
	in, err := os.Open(from)
	if err != nil {
		return fmt.Errorf("copy the program: %w", err)
	}
	defer func() { _ = in.Close() }()
	out, err := os.OpenFile(to, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0o755)
	if err != nil {
		return fmt.Errorf("copy the program beside the code: %w", err)
	}
	if _, err := io.Copy(out, in); err != nil {
		_ = out.Close()
		return fmt.Errorf("copy the program beside the code: %w", err)
	}
	return out.Close()
}
