package process

import (
	"context"
	"errors"
	"fmt"
	"os/exec"
	"path/filepath"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	"github.com/codeplai/VizcachaIDE/wails/internal/app"
	"github.com/codeplai/VizcachaIDE/wails/internal/protocol/pty"
)

const (
	// partialLineDelay is how long a line without its end stays hidden. Prompts such as
	// "Name: " have no newline, so after this pause they are shown anyway.
	partialLineDelay = 50 * time.Millisecond
	// drainTime is how long the output of an ended program may still arrive.
	drainTime = 300 * time.Millisecond
	readSize  = 4096
)

// terminalSession is a program running in a pseudoterminal. Its output is cleaned of escape
// sequences line by line and sent as stdout.
type terminalSession struct {
	terminal *pty.Terminal
	sink     app.EventSink
	output   atomic.Bool
	ended    chan struct{}
	drained  chan struct{}

	mu      sync.Mutex
	pending string
	flusher *time.Timer
}

func startTerminal(ctx context.Context, job Job, sink app.EventSink) (session, error) {
	var env []string
	if job.Env != nil {
		env = environmentList(job.Env)
	}
	terminal, err := pty.Start(pty.Program{Command: job.Command, Args: job.Args, Dir: job.Dir, Env: env})
	if err != nil {
		if errors.Is(err, exec.ErrNotFound) {
			return nil, app.MissingTool(strings.TrimSuffix(filepath.Base(job.Command), filepath.Ext(job.Command)))
		}
		return nil, fmt.Errorf("start %s: %w", filepath.Base(job.Command), err)
	}
	session := &terminalSession{
		terminal: terminal, sink: sink, ended: make(chan struct{}), drained: make(chan struct{}),
	}
	go session.read()
	go session.killOnCancel(ctx)
	return session, nil
}

// killOnCancel ends the program at once when the context of Start is canceled.
func (t *terminalSession) killOnCancel(ctx context.Context) {
	select {
	case <-ctx.Done():
		killTree(t.terminal.Pid())
	case <-t.ended:
	}
}

// read turns the terminal's output into events until the terminal is closed.
func (t *terminalSession) read() {
	defer close(t.drained)
	buffer := make([]byte, readSize)
	for {
		count, err := t.terminal.Read(buffer)
		if count > 0 {
			t.accept(string(buffer[:count]))
		}
		if err != nil {
			return
		}
	}
}

// accept sends the complete lines of the new text and schedules the unfinished one.
func (t *terminalSession) accept(text string) {
	t.mu.Lock()
	defer t.mu.Unlock()
	t.pending += text
	if cut := strings.LastIndexByte(t.pending, '\n'); cut >= 0 {
		t.emit(t.pending[:cut+1])
		t.pending = t.pending[cut+1:]
	}
	if t.flusher != nil {
		t.flusher.Stop()
	}
	if t.pending != "" {
		t.flusher = time.AfterFunc(partialLineDelay, t.flush)
	}
}

// flush shows the unfinished line (a prompt).
func (t *terminalSession) flush() {
	t.mu.Lock()
	defer t.mu.Unlock()
	if t.pending == "" {
		return
	}
	t.emit(t.pending)
	t.pending = ""
}

// emit cleans the text and sends it. The caller holds the lock.
func (t *terminalSession) emit(text string) {
	cleaned := strings.ReplaceAll(pty.Clean(text), "\r\n", "\n")
	if cleaned == "" {
		return
	}
	t.output.Store(true)
	t.sink.RunOutput("stdout", cleaned)
}

// write types a line; the terminal sends it to the program and echoes it back.
func (t *terminalSession) write(text string) error {
	line := strings.TrimRight(text, "\r\n") + "\r"
	if _, err := t.terminal.Write([]byte(line)); err != nil {
		return fmt.Errorf("write to the program: %w", err)
	}
	return nil
}

// stop sends Ctrl+C and kills the tree if the program stays alive: ConPTY ignores the ETX byte.
func (t *terminalSession) stop() {
	_ = t.terminal.Interrupt()
	pid := t.terminal.Pid()
	time.AfterFunc(killGrace, func() {
		select {
		case <-t.ended:
		default:
			killTree(pid)
		}
	})
}

func (t *terminalSession) seen() bool { return t.output.Load() }

func (t *terminalSession) wait() int {
	err := t.terminal.Wait()
	select {
	case <-t.drained:
	case <-time.After(drainTime):
	}
	_ = t.terminal.Close()
	<-t.drained
	t.mu.Lock()
	if t.flusher != nil {
		t.flusher.Stop()
	}
	t.mu.Unlock()
	t.flush()
	close(t.ended)
	return exitCode(err)
}

// exitCode reads the exit status from the error of Wait: 0 for nil, 1 when it is not an exit error.
func exitCode(err error) int {
	if err == nil {
		return 0
	}
	var exit *exec.ExitError
	if errors.As(err, &exit) {
		return exit.ExitCode()
	}
	return 1
}
