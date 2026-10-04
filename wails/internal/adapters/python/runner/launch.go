package runner

import (
	"context"
	"errors"
	"sync"

	"github.com/codeplai/VizcachaIDE/wails/internal/adapters/python"
	"github.com/codeplai/VizcachaIDE/wails/internal/app"
	"github.com/codeplai/VizcachaIDE/wails/internal/domain"
	"github.com/codeplai/VizcachaIDE/wails/internal/protocol/process"
)

// launch starts the program in a pseudoterminal. If the terminal cannot start it falls back to
// pipes with unbuffered output (-u) and no echo, so the program still runs.
func (r *Runner) launch(ctx context.Context, interpreter python.Interpreter, config domain.RunConfiguration, cleanup func()) error {
	env := python.Environment(r.base, interpreter)
	terminal := programJob(interpreter, config, env, process.Terminal)
	guard := newStartupCleanup(cleanup)
	terminal.Cleanup = guard.run
	err := r.supervisor.Start(ctx, terminal)
	if err == nil {
		guard.started()
		return nil
	}
	if errors.Is(err, app.ErrBusy) || errors.Is(err, app.ErrToolNotFound) {
		if cleanup != nil {
			cleanup()
		}
		return err
	}
	pipes := programJob(interpreter, config, env, process.Pipes)
	pipes.Cleanup = cleanup
	return r.supervisor.Start(ctx, pipes)
}

// programJob describes "python -X utf8 file.py args" for the supervisor.
func programJob(interpreter python.Interpreter, config domain.RunConfiguration, env map[string]string, mode process.Mode) process.Job {
	args := []string{"-X", "utf8"}
	config.Echo = mode == process.Terminal
	if !config.Echo {
		args = []string{"-u", "-X", "utf8"}
	}
	args = append(append(args, config.Target), config.ProgramArgs...)
	return process.Job{
		Config: config, Command: interpreter.Path, Args: args, Dir: config.WorkingDir, Env: env, Mode: mode,
	}
}

// startupCleanup holds back the cleanup of the first attempt: the supervisor runs a job's
// cleanup when Start fails, but a failed terminal must not delete what the pipes retry needs.
type startupCleanup struct {
	mu      sync.Mutex
	real    func()
	began   bool
	pending bool
}

func newStartupCleanup(real func()) *startupCleanup { return &startupCleanup{real: real} }

// run is the job's Cleanup. Before the start is confirmed it only takes note.
func (c *startupCleanup) run() {
	c.mu.Lock()
	if !c.began {
		c.pending = true
		c.mu.Unlock()
		return
	}
	c.mu.Unlock()
	if c.real != nil {
		c.real()
	}
}

// started confirms the program is running; a cleanup that already fired (it ended at once) runs now.
func (c *startupCleanup) started() {
	c.mu.Lock()
	c.began = true
	pending := c.pending
	c.mu.Unlock()
	if pending && c.real != nil {
		c.real()
	}
}
