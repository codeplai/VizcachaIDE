package runner

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"os/exec"

	"github.com/codeplai/VizcachaIDE/wails/internal/domain"
	"github.com/codeplai/VizcachaIDE/wails/internal/protocol/process"
)

// ErrCompileFailed is wrapped by CompileForDebug when the compiler rejected the code; the
// compiler's rendered text is returned beside it.
var ErrCompileFailed = errors.New("the compiler rejected the code")

// Build implements app.ProgramRunner: only the compiler stage. A loose file leaves
// <name>(.exe) next to the source; a crate builds into its target/. It emits the run events.
func (r *Runner) Build(ctx context.Context, config domain.RunConfiguration) error {
	c, err := r.compile(ctx, config, besideSource)
	if err != nil {
		return err
	}
	return r.supervisor.Start(ctx, r.compileJob(c, config))
}

// CompileForDebug is stage 1 for the debugger (track R2): it compiles the configuration
// synchronously, without run events, into the same place Run uses (the build cache for a loose
// file, target/debug for a crate) and returns the executable. output is the compiler's rendered
// text (warnings too). When the compiler rejects the code, exePath is "" and err wraps
// ErrCompileFailed; the output is what the Assistant explains. Other errors are
// app.MissingTool("rustc" or "cargo"), ErrNoBinary, ErrChooseMember or a failure to start.
func (r *Runner) CompileForDebug(ctx context.Context, config domain.RunConfiguration) (exePath, output string, err error) {
	c, err := r.compile(ctx, config, r.cacheOutput)
	if err != nil {
		return "", "", err
	}
	text, err := capture(ctx, c)
	if err != nil {
		return "", text, err
	}
	return c.program(), text, nil
}

// capture runs the compiler call outside the supervisor and returns its output as the user
// would see it. A non-zero exit wraps ErrCompileFailed.
func capture(ctx context.Context, c compilation) (string, error) {
	cmd := exec.CommandContext(ctx, c.command, c.args...)
	cmd.Dir, cmd.Env = c.dir, environmentList(c.env)
	process.HideConsole(cmd)
	var output bytes.Buffer
	cmd.Stdout, cmd.Stderr = &output, &output
	err := cmd.Run()
	text := c.filter.text(output.String())
	var exit *exec.ExitError
	if errors.As(err, &exit) {
		return text, fmt.Errorf("%w (exit code %d)", ErrCompileFailed, exit.ExitCode())
	}
	if err != nil {
		return text, fmt.Errorf("start the compiler: %w", err)
	}
	return text, nil
}

func environmentList(env map[string]string) []string {
	list := make([]string, 0, len(env))
	for name, value := range env {
		list = append(list, name+"="+value)
	}
	return list
}
