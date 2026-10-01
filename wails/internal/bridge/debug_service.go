package bridge

import (
	"context"

	"github.com/codeplai/VizcachaIDE/wails/internal/app"
	"github.com/codeplai/VizcachaIDE/wails/internal/domain"
)

// DebugService drives the debugger. It forwards each call to app.Debugger, which
// emits debug:stopped, debug:variables, debug:output and debug:terminated through
// the EventSink it was built with.
type DebugService struct {
	debugger app.Debugger
}

// NewDebugService creates the service on top of a debugger.
func NewDebugService(debugger app.Debugger) *DebugService {
	return &DebugService{debugger: debugger}
}

// Start launches the program under the debugger and stops at the first breakpoint.
func (s *DebugService) Start(path string, breakpoints []domain.Breakpoint) error {
	config := domain.NewFileRunConfiguration(path, nil)
	return s.debugger.Start(context.Background(), config, breakpoints)
}

// SetBreakpoints replaces the breakpoints of one file.
func (s *DebugService) SetBreakpoints(file string, lines []int) error {
	return s.debugger.SetBreakpoints(file, lines)
}

// StepOver runs the current line without entering functions ("Next line").
func (s *DebugService) StepOver() error { return s.debugger.StepOver() }

// StepInto enters the function called on the current line ("Go into function").
func (s *DebugService) StepInto() error { return s.debugger.StepInto() }

// StepOut runs until the current function returns ("Leave function").
func (s *DebugService) StepOut() error { return s.debugger.StepOut() }

// Resume continues until the next breakpoint.
func (s *DebugService) Resume() error { return s.debugger.Resume() }

// RunTo runs until the given location ("Run to here").
func (s *DebugService) RunTo(location domain.SourceLocation) error {
	return s.debugger.RunTo(location)
}

// RequestVariables asks for the children of a variable; the answer is debug:variables.
func (s *DebugService) RequestVariables(reference int) error {
	return s.debugger.RequestVariables(reference)
}

// Stop ends the session.
func (s *DebugService) Stop() error { return s.debugger.Stop() }
