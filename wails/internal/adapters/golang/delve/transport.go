package delve

import (
	"context"
	"fmt"
	"io"
	"net"

	protodap "github.com/codeplai/VizcachaIDE/wails/internal/protocol/dap"
)

// tcpTransport reaches the "dlv dap" server over TCP once it announces its address.
type tcpTransport struct {
	process *adapterProcess
}

var _ protodap.Transport = tcpTransport{}

// Open waits for Delve to say where it listens and dials it.
func (t tcpTransport) Open(ctx context.Context) (io.ReadWriteCloser, error) {
	address, err := t.process.waitAddress(ctx)
	if err != nil {
		return nil, err
	}
	var dialer net.Dialer
	conn, err := dialer.DialContext(ctx, "tcp", address)
	if err != nil {
		return nil, fmt.Errorf("connecting to delve: %w", err)
	}
	return conn, nil
}

// Close kills Delve (and the debuggee with it) by its own PID.
func (t tcpTransport) Close() error {
	t.process.kill()
	return nil
}
