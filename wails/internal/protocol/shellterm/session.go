package shellterm

import (
	"errors"
	"fmt"
	"os/exec"
	"sync/atomic"
	"time"

	"github.com/codeplai/VizcachaIDE/wails/internal/app"
	"github.com/codeplai/VizcachaIDE/wails/internal/protocol/process"
	"github.com/codeplai/VizcachaIDE/wails/internal/protocol/pty"
)

const (
	readSize = 4096
	// drainTime is how long the last output of an ended shell may still arrive.
	drainTime = 300 * time.Millisecond
)

// session is one shell in a pseudoterminal.
type session struct {
	id       string
	terminal *pty.Terminal
	sink     app.TerminalSink
	closed   atomic.Bool // the user closed it: no exit event
	drained  chan struct{}
}

func startSession(id string, program pty.Program, sink app.TerminalSink) (*session, error) {
	terminal, err := pty.Start(program)
	if err != nil {
		return nil, fmt.Errorf("start the shell %s: %w", program.Command, err)
	}
	s := &session{id: id, terminal: terminal, sink: sink, drained: make(chan struct{})}
	go s.read()
	return s, nil
}

// read sends the output in chunks that end on a character boundary, until the pty closes.
func (s *session) read() {
	defer close(s.drained)
	buffer := make([]byte, 0, 2*readSize)
	chunk := make([]byte, readSize)
	for {
		count, err := s.terminal.Read(chunk)
		buffer = append(buffer, chunk[:count]...)
		if cut := completeLength(buffer); cut > 0 {
			s.sink.TerminalOutput(s.id, string(buffer[:cut]))
			buffer = append(buffer[:0], buffer[cut:]...)
		}
		if err != nil {
			if len(buffer) > 0 {
				s.sink.TerminalOutput(s.id, string(buffer))
			}
			return
		}
	}
}

// watch waits for the shell to end, lets its last output arrive and reports the exit code.
func (s *session) watch(forget func()) {
	err := s.terminal.Wait()
	select {
	case <-s.drained:
	case <-time.After(drainTime):
	}
	_ = s.terminal.Close()
	<-s.drained
	forget()
	if !s.closed.Load() {
		s.sink.TerminalExit(s.id, exitCode(err))
	}
}

// close ends the shell and everything it started.
func (s *session) close() {
	s.closed.Store(true)
	process.KillTree(s.terminal.Pid())
}

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
