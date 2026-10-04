package process

import (
	"context"
	"errors"
	"fmt"
	"io"
	"os/exec"
	"path/filepath"
	"sort"
	"strings"
	"sync/atomic"

	"github.com/codeplai/VizcachaIDE/wails/internal/app"
)

// pipeSession is a program connected through stdin, stdout and stderr pipes.
type pipeSession struct {
	cmd    *exec.Cmd
	stdin  io.WriteCloser
	stdout *streamWriter
	stderr *streamWriter
	output *atomic.Bool
	ended  chan struct{}
}

func startPipes(ctx context.Context, job Job, sink JobEvents) (session, error) {
	output := &atomic.Bool{}
	pipes := &pipeSession{
		stdout: newStreamWriter(sink, "stdout", output),
		stderr: newStreamWriter(sink, "stderr", output),
		output: output,
		ended:  make(chan struct{}),
	}
	cmd := exec.CommandContext(ctx, job.Command, job.Args...)
	cmd.Dir = job.Dir
	if job.Env != nil {
		cmd.Env = environmentList(job.Env)
	}
	cmd.Stdout, cmd.Stderr = pipes.stdout, pipes.stderr
	cmd.WaitDelay = waitDelay
	cmd.Cancel = func() error { killTree(cmd.Process.Pid); return nil }
	prepareTree(cmd)
	stdin, err := cmd.StdinPipe()
	if err != nil {
		return nil, fmt.Errorf("prepare stdin: %w", err)
	}
	if err := cmd.Start(); err != nil {
		if errors.Is(err, exec.ErrNotFound) {
			return nil, app.MissingTool(strings.TrimSuffix(filepath.Base(job.Command), filepath.Ext(job.Command)))
		}
		return nil, fmt.Errorf("start %s: %w", filepath.Base(job.Command), err)
	}
	pipes.cmd, pipes.stdin = cmd, stdin
	return pipes, nil
}

// write types a line into the program's stdin.
func (p *pipeSession) write(text string) error {
	line := strings.TrimRight(text, "\r\n") + "\n"
	if _, err := io.WriteString(p.stdin, line); err != nil {
		return fmt.Errorf("write to the program: %w", err)
	}
	return nil
}

func (p *pipeSession) stop() { terminateTree(p.cmd.Process.Pid, p.ended) }

func (p *pipeSession) seen() bool { return p.output.Load() }

func (p *pipeSession) wait() int {
	_ = p.cmd.Wait() // the exit code is read from ProcessState
	p.stdout.Flush()
	p.stderr.Flush()
	close(p.ended)
	return exitCodeOf(p.cmd.ProcessState)
}

// environmentList is env in the "NAME=value" form os/exec wants.
func environmentList(env map[string]string) []string {
	list := make([]string, 0, len(env))
	for name, value := range env {
		list = append(list, name+"="+value)
	}
	sort.Strings(list)
	return list
}
