package delve

import (
	"context"
	"fmt"

	"github.com/codeplai/VizcachaIDE/wails/internal/app"
	"github.com/codeplai/VizcachaIDE/wails/internal/domain"
	"github.com/google/go-dap"
)

// execute resumes the program with next, stepIn, stepOut or continue.
func (s *session) execute(command string) error {
	s.mu.Lock()
	threadID, ready := s.threadID, s.configured && !s.finished
	s.mu.Unlock()
	if !ready || threadID == 0 {
		return app.ErrNoSession
	}
	s.resumed()
	request, err := executionRequest(command, threadID)
	if err != nil {
		return err
	}
	_, err = s.call(context.Background(), request)
	return err
}

func executionRequest(command string, threadID int) (dap.RequestMessage, error) {
	switch command {
	case "next":
		return &dap.NextRequest{Request: newRequest(command), Arguments: dap.NextArguments{ThreadId: threadID}}, nil
	case "stepIn":
		return &dap.StepInRequest{Request: newRequest(command), Arguments: dap.StepInArguments{ThreadId: threadID}}, nil
	case "stepOut":
		return &dap.StepOutRequest{Request: newRequest(command), Arguments: dap.StepOutArguments{ThreadId: threadID}}, nil
	case "continue":
		return &dap.ContinueRequest{Request: newRequest(command), Arguments: dap.ContinueArguments{ThreadId: threadID}}, nil
	}
	return nil, fmt.Errorf("unknown execution command %q", command)
}

// runTo puts a temporary breakpoint at the location and continues.
func (s *session) runTo(location domain.SourceLocation) error {
	if !s.isConfigured() {
		return app.ErrNoSession
	}
	file := s.book.SetTemporary(location)
	if err := s.sendBreakpoints(file); err != nil {
		return err
	}
	return s.execute("continue")
}

// setBreakpoints replaces the lines of one file and tells Delve when it is ready.
func (s *session) setBreakpoints(file string, lines []int) error {
	key := s.book.Replace(file, lines)
	if !s.isConfigured() {
		return nil
	}
	return s.sendBreakpoints(key)
}

// requestVariables answers asynchronously with debug:variables. The answer is
// always emitted, empty on failure, except when the program resumed first.
func (s *session) requestVariables(reference int) {
	ctx := s.pausedContext()
	if ctx == nil || reference == 0 {
		s.sink.DebugVariables(reference, []domain.Variable{})
		return
	}
	go func() {
		children := s.children(ctx, reference)
		if ctx.Err() == nil {
			s.sink.DebugVariables(reference, children)
		}
	}()
}
