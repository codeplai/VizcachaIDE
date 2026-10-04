package lsp

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"os/exec"
	"time"

	"go.lsp.dev/jsonrpc2"

	"github.com/codeplai/VizcachaIDE/wails/internal/app"
)

const exitGrace = 2 * time.Second

// notificationHandler receives the notifications the server sends (publishDiagnostics...).
type notificationHandler func(method string, params json.RawMessage)

// connection is a running language server process and its JSON-RPC link.
type connection struct {
	conn jsonrpc2.Conn
	cmd  *exec.Cmd
}

// stdio joins the process pipes into one io.ReadWriteCloser for jsonrpc2.
type stdio struct {
	io.Reader
	io.WriteCloser
}

func (s stdio) Close() error { return s.WriteCloser.Close() }

// startConnection launches the server and starts reading its messages.
func startConnection(executable string, args []string, env map[string]string, dir string, handle notificationHandler) (*connection, error) {
	cmd := exec.Command(executable, args...)
	cmd.Dir = dir
	cmd.Env = mergedEnvironment(env)
	hideWindow(cmd)
	stdin, err := cmd.StdinPipe()
	if err != nil {
		return nil, fmt.Errorf("language server stdin: %w", err)
	}
	stdout, err := cmd.StdoutPipe()
	if err != nil {
		return nil, fmt.Errorf("language server stdout: %w", err)
	}
	if err := cmd.Start(); err != nil {
		return nil, fmt.Errorf("start language server: %w: %w", app.ErrToolNotFound, err)
	}
	conn := jsonrpc2.NewConn(jsonrpc2.NewStream(stdio{Reader: stdout, WriteCloser: stdin}))
	conn.Go(context.Background(), func(ctx context.Context, reply jsonrpc2.Replier, req jsonrpc2.Request) error {
		if _, isCall := req.(*jsonrpc2.Call); isCall {
			return reply(ctx, nil, nil) // server-to-client requests: nothing to offer
		}
		handle(req.Method(), req.Params())
		return nil
	})
	return &connection{conn: conn, cmd: cmd}, nil
}

// mergedEnvironment is the process environment overridden by the toolchain's variables.
func mergedEnvironment(overrides map[string]string) []string {
	if len(overrides) == 0 {
		return os.Environ()
	}
	merged := map[string]string{}
	for _, entry := range os.Environ() {
		for i := 1; i < len(entry); i++ { // i starts at 1: Windows has "=C:=..." entries
			if entry[i] == '=' {
				merged[entry[:i]] = entry[i+1:]
				break
			}
		}
	}
	for key, value := range overrides {
		merged[key] = value
	}
	result := make([]string, 0, len(merged))
	for key, value := range merged {
		result = append(result, key+"="+value)
	}
	return result
}

// call sends a request and decodes the raw result.
func (c *connection) call(ctx context.Context, method string, params any) (json.RawMessage, error) {
	var result json.RawMessage
	if _, err := c.conn.Call(ctx, method, params, &result); err != nil {
		return nil, fmt.Errorf("language server %s: %w", method, err)
	}
	return result, nil
}

func (c *connection) notify(ctx context.Context, method string, params any) error {
	if err := c.conn.Notify(ctx, method, params); err != nil {
		return fmt.Errorf("language server %s: %w", method, err)
	}
	return nil
}

// close asks the server to exit and kills the process if it does not.
func (c *connection) close(ctx context.Context) {
	shutdownCtx, cancel := context.WithTimeout(ctx, exitGrace)
	defer cancel()
	_, _ = c.conn.Call(shutdownCtx, "shutdown", nil, nil)
	_ = c.conn.Notify(shutdownCtx, "exit", nil)
	_ = c.conn.Close()
	exited := make(chan struct{})
	go func() { _ = c.cmd.Wait(); close(exited) }()
	select {
	case <-exited:
	case <-time.After(exitGrace):
		_ = c.cmd.Process.Kill()
		<-exited
	}
}
