package debugpy

import (
	"github.com/codeplai/VizcachaIDE/wails/internal/app"
	"github.com/codeplai/VizcachaIDE/wails/internal/domain"
	"github.com/codeplai/VizcachaIDE/wails/internal/protocol/pty"
)

// terminalWorks reports whether the interpreter can run in a pseudoterminal on this machine.
func terminalWorks(interpreter string) bool {
	terminal, err := pty.Start(pty.Program{Command: interpreter, Args: []string{"-c", "pass"}})
	if err != nil {
		return false
	}
	_ = terminal.Wait()
	_ = terminal.Close()
	return true
}

// SetBreakpoints replaces the breakpoints of one file, also while running.
func (d *Debugger) SetBreakpoints(file string, lines []int) error {
	if current := d.active(); current != nil {
		return current.SetBreakpoints(file, lines)
	}
	d.book.Replace(file, lines)
	return nil
}

// StepOver runs the current line without entering functions.
func (d *Debugger) StepOver() error { return d.execute("next") }

// StepInto enters the function called on the current line.
func (d *Debugger) StepInto() error { return d.execute("stepIn") }

// StepOut runs until the current function returns.
func (d *Debugger) StepOut() error { return d.execute("stepOut") }

// Resume continues until the next breakpoint.
func (d *Debugger) Resume() error { return d.execute("continue") }

func (d *Debugger) execute(command string) error {
	current := d.active()
	if current == nil {
		return app.ErrNoSession
	}
	return current.Execute(command)
}

// RunTo runs until the given location.
func (d *Debugger) RunTo(location domain.SourceLocation) error {
	current := d.active()
	if current == nil {
		return app.ErrNoSession
	}
	return current.RunTo(location)
}

// RequestVariables asks for the children of a variable; the answer is debug:variables.
func (d *Debugger) RequestVariables(reference int) error {
	current := d.active()
	if current == nil {
		d.sink.DebugVariables(reference, []domain.Variable{})
		return nil
	}
	current.RequestVariables(reference)
	return nil
}

// FrameVariables returns the arguments and locals of one frame of the paused stack.
func (d *Debugger) FrameVariables(frameID int) (domain.FrameVariables, error) {
	current := d.active()
	if current == nil {
		return domain.FrameVariables{Arguments: []domain.Variable{}, Locals: []domain.Variable{}}, nil
	}
	return current.FrameVariables(frameID), nil
}

// Stop ends the session; debug:terminated carries domain.TerminatedByUser.
func (d *Debugger) Stop() error {
	current := d.active()
	if current == nil {
		return app.ErrNoSession
	}
	current.Finish(domain.TerminatedByUser)
	return nil
}

// IsActive reports whether a session is alive.
func (d *Debugger) IsActive() bool { return d.active() != nil }
