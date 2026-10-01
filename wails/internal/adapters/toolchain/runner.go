package toolchain

import (
	"context"
	"errors"
	"fmt"
	"io"
	"os/exec"
	"strings"
	"sync/atomic"
	"time"

	"github.com/codeplai/VizcachaIDE/wails/internal/app"
	"github.com/codeplai/VizcachaIDE/wails/internal/domain"
)

const waitDelay = 3 * time.Second

// process is the one program (or go command) the IDE is running.
type process struct {
	cmd     *exec.Cmd
	stdin   io.WriteCloser
	stdout  *streamWriter
	stderr  *streamWriter
	seen    *atomic.Bool
	started time.Time
	done    chan struct{}
}

// launch describes one go command to run.
type launch struct {
	config  domain.RunConfiguration
	args    []string
	cleanup func()
}

func runArguments(config domain.RunConfiguration) []string {
	return append([]string{"run", config.GoTargetArgument()}, config.ProgramArgs...)
}

// Run implements app.Toolchain.
func (t *Toolchain) Run(ctx context.Context, config domain.RunConfiguration) error {
	return t.launch(ctx, launch{config: config, args: runArguments(config)})
}

// Build implements app.Toolchain.
func (t *Toolchain) Build(ctx context.Context, config domain.RunConfiguration) error {
	output := config.ExecutableName(isWindows())
	args := []string{"build", "-o", output, config.GoTargetArgument()}
	return t.launch(ctx, launch{config: config, args: args})
}

// RunGoCommand implements app.Toolchain.
func (t *Toolchain) RunGoCommand(ctx context.Context, workingDir string, args []string) error {
	config := domain.RunConfiguration{
		Target:      strings.Join(append([]string{"go"}, args...), " "),
		WorkingDir:  workingDir,
		Mode:        domain.RunPackage,
		ProgramArgs: []string{},
	}
	return t.launch(ctx, launch{config: config, args: args})
}

// launch starts the go command and watches it in the background.
func (t *Toolchain) launch(ctx context.Context, job launch) error {
	if job.cleanup == nil {
		job.cleanup = func() {}
	}
	proc, err := t.begin(ctx, job)
	if err != nil {
		job.cleanup()
		return err
	}
	t.sink.RunStarted(job.config)
	go t.supervise(proc, job)
	return nil
}

// begin starts the process unless another one is alive.
func (t *Toolchain) begin(ctx context.Context, job launch) (*process, error) {
	goPath := t.locator.Locate(ToolGo)
	if goPath.Origin == OriginMissing {
		return nil, fmt.Errorf("go: %w", app.ErrToolNotFound)
	}
	t.mu.Lock()
	defer t.mu.Unlock()
	if t.current != nil {
		return nil, app.ErrBusy
	}
	proc, err := t.startProcess(ctx, goPath.Path, job)
	if err != nil {
		return nil, err
	}
	t.current = proc
	return proc, nil
}

func (t *Toolchain) startProcess(ctx context.Context, goPath string, job launch) (*process, error) {
	seen := &atomic.Bool{}
	proc := &process{
		stdout:  newStreamWriter(t.sink, "stdout", seen),
		stderr:  newStreamWriter(t.sink, "stderr", seen),
		seen:    seen,
		started: time.Now(),
		done:    make(chan struct{}),
	}
	cmd := exec.CommandContext(ctx, goPath, job.args...)
	cmd.Dir = job.config.WorkingDir
	cmd.Env = environmentList(t.Environment())
	cmd.Stdout, cmd.Stderr = proc.stdout, proc.stderr
	cmd.WaitDelay = waitDelay
	cmd.Cancel = func() error { killTree(cmd.Process.Pid); return nil }
	prepareTree(cmd)
	stdin, err := cmd.StdinPipe()
	if err != nil {
		return nil, fmt.Errorf("prepare stdin: %w", err)
	}
	if err := cmd.Start(); err != nil {
		if errors.Is(err, exec.ErrNotFound) {
			return nil, fmt.Errorf("go: %w", app.ErrToolNotFound)
		}
		return nil, fmt.Errorf("start go: %w", err)
	}
	proc.cmd, proc.stdin = cmd, stdin
	return proc, nil
}

// supervise waits for the process, reports the end and frees the toolchain.
func (t *Toolchain) supervise(proc *process, job launch) {
	notice := time.AfterFunc(t.delay, func() { t.announceFirstBuild(proc) })
	_ = proc.cmd.Wait() // the exit code is read from ProcessState
	notice.Stop()
	proc.stdout.Flush()
	proc.stderr.Flush()
	exitCode := proc.cmd.ProcessState.ExitCode()
	job.cleanup()
	t.mu.Lock()
	t.current = nil
	t.mu.Unlock()
	close(proc.done)
	t.sink.RunFinished(exitCode, time.Since(proc.started).Milliseconds())
}

// announceFirstBuild tells the user why nothing has been printed yet. The text
// travels as ordinary stdout output: the event contract has no separate channel.
func (t *Toolchain) announceFirstBuild(proc *process) {
	if t.notice == nil || proc.seen.Load() {
		return
	}
	select {
	case <-proc.done:
		return
	default:
	}
	t.sink.RunOutput("stdout", t.notice()+"\n")
}
