package dap

import (
	"context"
	"fmt"

	"github.com/codeplai/VizcachaIDE/wails/internal/app"
	"github.com/codeplai/VizcachaIDE/wails/internal/domain"
	"github.com/google/go-dap"
)

// Execute resumes the program with next, stepIn, stepOut or continue.
func (s *Session) Execute(command string) error { return s.execute(command) }

// execute resumes the program with next, stepIn, stepOut or continue.
func (s *Session) execute(command string) error {
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

// RunTo puts a temporary breakpoint at the location and continues.
func (s *Session) RunTo(location domain.SourceLocation) error { return s.runTo(location) }

// SetBreakpoints replaces the lines of one file, also while running.
func (s *Session) SetBreakpoints(file string, lines []int) error {
	return s.setBreakpoints(file, lines)
}

// RequestVariables answers asynchronously with debug:variables.
func (s *Session) RequestVariables(reference int) { s.requestVariables(reference) }

// FrameVariables reads the arguments and locals of one frame of the paused stack.
func (s *Session) FrameVariables(frameID int) domain.FrameVariables { return s.frameVariables(frameID) }

// runTo puts a temporary breakpoint at the location and continues.
func (s *Session) runTo(location domain.SourceLocation) error {
	if !s.isConfigured() {
		return app.ErrNoSession
	}
	file := s.book.SetTemporary(location)
	if err := s.sendBreakpoints(file); err != nil {
		return err
	}
	return s.execute("continue")
}

// setBreakpoints replaces the lines of one file and tells the adapter when it is ready.
func (s *Session) setBreakpoints(file string, lines []int) error {
	key := s.book.Replace(file, lines)
	if !s.isConfigured() {
		return nil
	}
	return s.sendBreakpoints(key)
}

// requestVariables answers asynchronously with debug:variables. The answer is
// always emitted, empty on failure, except when the program resumed first.
func (s *Session) requestVariables(reference int) {
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
