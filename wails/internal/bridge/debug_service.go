package bridge

import (
	"context"
	"fmt"
	"sync"

	"github.com/codeplai/VizcachaIDE/wails/internal/app"
	"github.com/codeplai/VizcachaIDE/wails/internal/domain"
)

// DebugService drives the debugger of the language of the file. Each call goes to the debugger
// of the last session started (or to the default language's before any), which emits
// debug:stopped, debug:variables, debug:output and debug:terminated through its EventSink.
// There is one session at a time in the whole IDE.
type DebugService struct {
	supportRouter

	mu       sync.Mutex
	starting bool         // a Start is in progress: the session slot is reserved
	current  app.Debugger // debugger of the last Start
}

// NewDebugService creates the service.
func NewDebugService(registry *app.LanguageRegistry) *DebugService {
	return &DebugService{supportRouter: supportRouter{registry: registry}}
}

// Start launches the program under the debugger and stops at the first breakpoint.
// It debugs what Run would run (the package when the file is inside a Go module) with
// the same program arguments text (split like a shell). It returns app.ErrBusy while any
// session, of any language, is alive or starting.
func (s *DebugService) Start(path string, breakpoints []domain.Breakpoint, argsText string) error {
	args, err := app.SplitProgramArguments(argsText)
	if err != nil {
		return fmt.Errorf("debug %s: %w", path, err)
	}
	support, err := s.supportFor(path)
	if err != nil {
		return fmt.Errorf("debug %s: %w", path, err)
	}
	previous, err := s.reserve(support.Debugger)
	if err != nil {
		return fmt.Errorf("debug %s: %w", path, err)
	}
	config := support.Runner.Configure(path, args)
	err = support.Debugger.Start(context.Background(), config, breakpoints)
	s.release(previous, err != nil)
	return err
}

// reserve checks that no session is alive and takes the slot, under one lock, so two Starts can
// never both pass the check.
// It returns the debugger that served before, to put it back if the Start fails.
func (s *DebugService) reserve(debugger app.Debugger) (app.Debugger, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.starting || s.anySessionActive() {
		return nil, app.ErrBusy
	}
	previous := s.current
	s.starting = true
	s.current = debugger
	return previous, nil
}

func (s *DebugService) release(previous app.Debugger, failed bool) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.starting = false
	if failed {
		s.current = previous
	}
}

func (s *DebugService) anySessionActive() bool {
	for _, support := range s.registry.All() {
		if support.Debugger.IsActive() {
			return true
		}
	}
	return false
}

// debugger is the one the stepping calls go to.
func (s *DebugService) debugger() app.Debugger {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.current != nil {
		return s.current
	}
	return s.registry.Default().Debugger
}

// SetBreakpoints replaces the breakpoints of one file, in the debugger of its language.
func (s *DebugService) SetBreakpoints(file string, lines []int) error {
	support, err := s.supportFor(file)
	if err != nil {
		return fmt.Errorf("breakpoints %s: %w", file, err)
	}
	return support.Debugger.SetBreakpoints(file, lines)
}

// StepOver runs the current line without entering functions ("Next line").
func (s *DebugService) StepOver() error { return s.debugger().StepOver() }

// StepInto enters the function called on the current line ("Go into function").
func (s *DebugService) StepInto() error { return s.debugger().StepInto() }

// StepOut runs until the current function returns ("Leave function").
func (s *DebugService) StepOut() error { return s.debugger().StepOut() }

// Resume continues until the next breakpoint.
func (s *DebugService) Resume() error { return s.debugger().Resume() }

// RunTo runs until the given location ("Run to here").
func (s *DebugService) RunTo(location domain.SourceLocation) error {
	return s.debugger().RunTo(location)
}

// RequestVariables asks for the children of a variable; the answer is debug:variables.
func (s *DebugService) RequestVariables(reference int) error {
	return s.debugger().RequestVariables(reference)
}

// FrameVariables returns the arguments and locals of one frame of the paused stack.
func (s *DebugService) FrameVariables(frameID int) (domain.FrameVariables, error) {
	return s.debugger().FrameVariables(frameID)
}

// Stop ends the session.
func (s *DebugService) Stop() error { return s.debugger().Stop() }
