package runner

import (
	"context"
	"io"

	"github.com/codeplai/VizcachaIDE/wails/internal/domain"
	"github.com/codeplai/VizcachaIDE/wails/internal/protocol/process"
	"github.com/codeplai/VizcachaIDE/wails/internal/protocol/pty"
)

// Run implements app.ProgramRunner: build the CMake project, then run the program.
func (r *Runner) Run(ctx context.Context, config domain.RunConfiguration) error {
	if project, ok := projectOf(config); ok {
		return r.runProject(ctx, config, project)
	}
	return r.start(ctx, config, nil)
}

// Build implements app.ProgramRunner: only the build stages, with the executable next to the
// code (nandu.exe, nandu), like "go build". A CMake project builds in Release. It emits the run
// events.
func (r *Runner) Build(ctx context.Context, config domain.RunConfiguration) error {
	if project, ok := projectOf(config); ok {
		return r.buildProject(ctx, config, project)
	}
	c, err := r.compile(ctx, config, besideSource)
	if err != nil {
		return err
	}
	return r.supervisor.Start(ctx, r.compileJob(c, config))
}

// start runs a configuration compiled directly (an untitled file) as one two-stage job. cleanup runs when the run ends or cannot
// start.
func (r *Runner) start(ctx context.Context, config domain.RunConfiguration, cleanup func()) error {
	c, err := r.compile(ctx, config, r.cacheOutput)
	if err != nil {
		finish(cleanup)
		return err
	}
	mode := process.Pipes
	if r.terminalWorks(ctx, c.compiler.Path) {
		mode = process.Terminal
	}
	config.Echo = mode == process.Terminal
	job := r.compileJob(c, config)
	job.Then = func(exitCode int) (*process.Job, bool) {
		if exitCode != 0 {
			finish(cleanup)
			return nil, false
		}
		next := programJob(c.output, c.folder, config, r.Environment(), mode)
		next.Cleanup = cleanup
		return &next, true
	}
	if err := r.supervisor.Start(ctx, job); err != nil {
		finish(cleanup)
		return err
	}
	return nil
}

// programJob is stage 2: the executable with the program's arguments. A terminal echoes what
// the user types, so Config.Echo follows the mode.
func programJob(executable, folder string, config domain.RunConfiguration, env map[string]string, mode process.Mode) process.Job {
	config.Echo = mode == process.Terminal
	return process.Job{
		Config: config, Command: executable, Args: config.ProgramArgs, Dir: folder, Env: env, Mode: mode,
		Finished: crashLine,
	}
}

func finish(cleanup func()) {
	if cleanup != nil {
		cleanup()
	}
}

// terminalWorks reports (once) whether a pseudoterminal can start on this machine. The program
// stage starts after the compiler, when the supervisor can no longer retry in another mode, so
// the choice is made before the run.
func (r *Runner) terminalWorks(ctx context.Context, compiler string) bool {
	r.probe.Do(func() { r.hasTerminal = r.terminal(ctx, compiler) })
	return r.hasTerminal
}

// terminalWorks starts "<compiler> --version" in a pseudoterminal and waits for it.
func terminalWorks(_ context.Context, compiler string) bool {
	terminal, err := pty.Start(pty.Program{Command: compiler, Args: []string{"--version"}})
	if err != nil {
		return false
	}
	go func() { _, _ = io.Copy(io.Discard, terminal) }()
	_ = terminal.Wait()
	_ = terminal.Close()
	return true
}
