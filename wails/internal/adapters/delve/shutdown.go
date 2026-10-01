package delve

import (
	"context"
	"errors"
	"strings"

	"github.com/google/go-dap"
)

// fail reports the error in the output and ends the session.
func (s *session) fail(err error) {
	if s.isFinished() || errors.Is(err, errConnectionClosed) {
		return
	}
	s.sink.DebugOutput(strings.TrimRight(err.Error(), "\n")+"\n", "stderr")
	s.finish(failedExitCode)
}

func (s *session) isFinished() bool {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.finished
}

// connectionClosed runs when Delve hangs up on its own.
func (s *session) connectionClosed() {
	if s.isFinished() {
		return
	}
	s.sink.DebugOutput(s.text("run.debugAdapterStopped")+"\n", "console")
	s.finish(failedExitCode)
}

// finish ends the session once: tells the UI, disconnects, kills Delve by PID.
func (s *session) finish(exitCode int) {
	s.mu.Lock()
	if s.finished {
		s.mu.Unlock()
		return
	}
	s.finished, s.configured = true, false
	client, cancel := s.client, s.stopCancel
	s.mu.Unlock()
	close(s.done)
	if cancel != nil {
		cancel()
	}
	s.sink.DebugTerminated(exitCode)
	if client != nil {
		disconnect(client)
		client.Close()
	}
	s.process.kill()
	s.onFinish()
}

func disconnect(client *Client) {
	ctx, cancel := context.WithTimeout(context.Background(), disconnectTimeout)
	defer cancel()
	request := &dap.DisconnectRequest{
		Request:   newRequest("disconnect"),
		Arguments: &dap.DisconnectArguments{TerminateDebuggee: true},
	}
	_, _ = client.Call(ctx, request)
}
