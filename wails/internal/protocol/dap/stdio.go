package dap

import (
	"context"
	"errors"
	"fmt"
	"io"
	"os"
	"os/exec"
	"sync"

	"github.com/codeplai/VizcachaIDE/wails/internal/protocol/process"
)

// StdioTransport starts a debug adapter that speaks DAP over its stdin and stdout
// (python -m debugpy.adapter, lldb-dap).
type StdioTransport struct {
	Command string
	Args    []string
	Dir     string
	// Env is the full environment of the adapter; nil means the IDE's own environment.
	Env []string
	// Stderr receives what the adapter writes to stderr (its own log); nil discards it.
	Stderr io.Writer

	mu  sync.Mutex
	cmd *exec.Cmd
}

var _ Transport = (*StdioTransport)(nil)

// Open starts the adapter process and returns its stdin and stdout as one connection.
func (t *StdioTransport) Open(ctx context.Context) (io.ReadWriteCloser, error) {
	t.mu.Lock()
	defer t.mu.Unlock()
	if t.cmd != nil {
		return nil, errors.New("the debug adapter is already running")
	}
	cmd := exec.CommandContext(ctx, t.Command, t.Args...)
	cmd.Dir, cmd.Env, cmd.Stderr = t.Dir, t.Env, t.Stderr
	if cmd.Env == nil {
		cmd.Env = os.Environ()
	}
	process.HideConsole(cmd)
	stdin, err := cmd.StdinPipe()
	if err != nil {
		return nil, err
	}
	stdout, err := cmd.StdoutPipe()
	if err != nil {
		return nil, err
	}
	if err := cmd.Start(); err != nil {
		return nil, fmt.Errorf("starting the debug adapter %s: %w", t.Command, err)
	}
	t.cmd = cmd
	return stdioConn{Reader: stdout, writer: stdin}, nil
}

// Close ends the adapter process; it is safe to call more than once.
func (t *StdioTransport) Close() error {
	t.mu.Lock()
	cmd := t.cmd
	t.cmd = nil
	t.mu.Unlock()
	if cmd == nil || cmd.Process == nil {
		return nil
	}
	_ = cmd.Process.Kill() // the adapter may already be gone
	_ = cmd.Wait()
	return nil
}

// stdioConn reads the adapter's stdout and writes to its stdin.
type stdioConn struct {
	io.Reader
	writer io.WriteCloser
}

func (c stdioConn) Write(p []byte) (int, error) { return c.writer.Write(p) }

// Close closes the adapter's stdin; the process ends through Transport.Close.
func (c stdioConn) Close() error { return c.writer.Close() }
