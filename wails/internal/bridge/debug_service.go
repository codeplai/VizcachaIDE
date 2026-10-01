package bridge

import (
	"github.com/codeplai/VizcachaIDE/wails/internal/app"
	"github.com/codeplai/VizcachaIDE/wails/internal/domain"
)

// DebugService drives the debugger. W0 STUB: owned by track G2, which replaces
// the body of each method with calls to app.Debugger. Emits debug:stopped,
// debug:variables, debug:output and debug:terminated.
type DebugService struct {
	sink app.EventSink
}

// NewDebugService creates the service.
func NewDebugService(sink app.EventSink) *DebugService { return &DebugService{sink: sink} }

// Start launches the program under the debugger and stops at the first breakpoint.
func (s *DebugService) Start(path string, breakpoints []domain.Breakpoint) error {
	s.sink.DebugOutput("Starting the debugger…", "console")
	s.sink.DebugStopped(sampleDebugState())
	return nil
}

// SetBreakpoints replaces the breakpoints of one file.
func (s *DebugService) SetBreakpoints(file string, lines []int) error { return nil }

// StepOver runs the current line without entering functions ("Next line").
func (s *DebugService) StepOver() error { return s.stopAgain(domain.StopStep) }

// StepInto enters the function called on the current line ("Go into function").
func (s *DebugService) StepInto() error { return s.stopAgain(domain.StopStep) }

// StepOut runs until the current function returns ("Leave function").
func (s *DebugService) StepOut() error { return s.stopAgain(domain.StopStep) }

// Resume continues until the next breakpoint.
func (s *DebugService) Resume() error {
	s.sink.DebugTerminated(0)
	return nil
}

// RunTo runs until the given location ("Run to here").
func (s *DebugService) RunTo(location domain.SourceLocation) error {
	return s.stopAgain(domain.StopStep)
}

// RequestVariables asks for the children of a variable; the answer is debug:variables.
func (s *DebugService) RequestVariables(reference int) error {
	s.sink.DebugVariables(reference, []domain.Variable{})
	return nil
}

// Stop ends the session.
func (s *DebugService) Stop() error {
	s.sink.DebugTerminated(domain.TerminatedByUser)
	return nil
}

func (s *DebugService) stopAgain(reason domain.StopReason) error {
	state := sampleDebugState()
	state.Reason = reason
	s.sink.DebugStopped(state)
	return nil
}
