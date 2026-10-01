package delve

import (
	"context"
	"os"

	"github.com/google/go-dap"
)

func osProcessID() int { return os.Getpid() }

// handleEvent runs on the connection's reading goroutine: it must never block.
func (s *session) handleEvent(event dap.EventMessage) {
	switch typed := event.(type) {
	case *dap.InitializedEvent:
		s.readyOnce.Do(func() { close(s.ready) })
	case *dap.OutputEvent:
		s.forwardOutput(typed)
	case *dap.StoppedEvent:
		s.onStopped(typed)
	case *dap.ExitedEvent:
		s.mu.Lock()
		s.exitCode = typed.Body.ExitCode
		s.mu.Unlock()
	case *dap.TerminatedEvent:
		go s.finish(s.recordedExitCode())
	}
}

func (s *session) recordedExitCode() int {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.exitCode
}

func (s *session) forwardOutput(event *dap.OutputEvent) {
	text, category, ok := outputText(event)
	if ok && !s.isFinished() { // skip Delve's "Detaching..." after the end
		s.sink.DebugOutput(text, category)
	}
}

// onStopped starts a stop context and builds the DebugState in the background.
func (s *session) onStopped(event *dap.StoppedEvent) {
	s.mu.Lock()
	if s.finished {
		s.mu.Unlock()
		return
	}
	if event.Body.ThreadId != 0 {
		s.threadID = event.Body.ThreadId
	}
	if s.stopCancel != nil {
		s.stopCancel()
	}
	ctx, cancel := context.WithCancel(context.Background())
	s.stopCtx, s.stopCancel = ctx, cancel
	s.mu.Unlock()
	if file := s.book.ClearTemporary(); file != "" {
		go func() { _ = s.sendBreakpoints(file) }()
	}
	go s.inspect(ctx, event)
}

// resumed cancels the stop context: answers still in flight are meaningless now.
func (s *session) resumed() {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.stopCancel != nil {
		s.stopCancel()
	}
	s.stopCtx, s.stopCancel = nil, nil
}

// pausedContext returns the context of the current stop, or nil while running.
func (s *session) pausedContext() context.Context {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.stopCtx
}
