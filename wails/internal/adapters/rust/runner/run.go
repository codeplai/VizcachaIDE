package runner

import (
	"context"
	"io"

	"github.com/codeplai/VizcachaIDE/wails/internal/domain"
	"github.com/codeplai/VizcachaIDE/wails/internal/protocol/process"
	"github.com/codeplai/VizcachaIDE/wails/internal/protocol/pty"
)

// Run implements app.ProgramRunner: compile, then run the program. A crate runs its built
// executable directly, not "cargo run", so the terminal gets the program and not cargo.
func (r *Runner) Run(ctx context.Context, config domain.RunConfiguration) error {
	return r.start(ctx, config, nil)
}

// start runs the configuration as one two-stage job. cleanup runs when the run ends or cannot
// start.
func (r *Runner) start(ctx context.Context, config domain.RunConfiguration, cleanup func()) error {
	c, err := r.compile(ctx, config, r.cacheOutput)
	if err != nil {
		finish(cleanup)
		return err
	}
	mode := process.Pipes
	if r.terminalWorks(ctx, c.command) {
		mode = process.Terminal
	}
	config.Echo = mode == process.Terminal
	job := r.compileJob(c, config)
	job.Then = func(exitCode int) (*process.Job, bool) {
		if exitCode != 0 {
			finish(cleanup)
			return nil, false
		}
		next := r.programJob(c, config, mode)
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
func (r *Runner) programJob(c compilation, config domain.RunConfiguration, mode process.Mode) process.Job {
	config.Echo = mode == process.Terminal
	return process.Job{
		Config: config, Command: c.program(), Args: config.ProgramArgs, Dir: config.WorkingDir,
		Env: r.Environment(), Mode: mode, Finished: crashLine,
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
func (r *Runner) terminalWorks(ctx context.Context, program string) bool {
	r.probe.Do(func() { r.hasTerminal = r.terminal(ctx, program) })
	return r.hasTerminal
}

// terminalWorks starts "<program> --version" in a pseudoterminal and waits for it.
func terminalWorks(_ context.Context, program string) bool {
	terminal, err := pty.Start(pty.Program{Command: program, Args: []string{"--version"}})
	if err != nil {
		return false
	}
	go func() { _, _ = io.Copy(io.Discard, terminal) }()
	_ = terminal.Wait()
	_ = terminal.Close()
	return true
}
