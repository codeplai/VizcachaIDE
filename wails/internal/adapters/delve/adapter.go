// Package delve implements app.Debugger with a DAP client (github.com/google/go-dap)
// talking to "dlv dap" over TCP.
package delve

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"sync"

	"github.com/codeplai/VizcachaIDE/wails/internal/app"
	"github.com/codeplai/VizcachaIDE/wails/internal/domain"
)

// Options configures the adapter. Every field is optional.
type Options struct {
	// DelvePath returns the dlv executable each time a session starts, so a change in the
	// settings needs no restart. Nil or an empty result means "dlv" from PATH.
	DelvePath func() string
	// Environment returns the variables the program runs with (Toolchain.Environment).
	Environment func() map[string]string
	// Translate returns the user-facing text of an i18n key; nil means English.
	Translate func(key string) string
}

var englishTexts = map[string]string{
	"errors.delveNotFound":    "The debugger (Delve) isn't available. Reinstall VizcachaIDE or install Delve with the command below.",
	"run.debugStdin":          "While debugging, your program can't read the keyboard.",
	"run.debugAdapterStopped": "The debugger stopped unexpectedly. Try starting it again.",
}

// Debugger is the Delve implementation of app.Debugger.
type Debugger struct {
	sink    app.EventSink
	options Options
	book    *app.BreakpointBook
	tracker *app.ChangeTracker

	mu      sync.Mutex
	current *session
}

var _ app.Debugger = (*Debugger)(nil)

// New creates a debugger that reports through sink.
func New(sink app.EventSink, options Options) *Debugger {
	return &Debugger{sink: sink, options: options, book: app.NewBreakpointBook(), tracker: app.NewChangeTracker()}
}

func (d *Debugger) text(key string) string {
	if d.options.Translate != nil {
		return d.options.Translate(key)
	}
	return englishTexts[key]
}

func (d *Debugger) delveExecutable() (string, error) {
	name := ""
	if d.options.DelvePath != nil {
		name = d.options.DelvePath()
	}
	if name == "" {
		name = "dlv"
	}
	path, err := exec.LookPath(name)
	if err != nil {
		return "", fmt.Errorf("%w: %s", app.ErrToolNotFound, d.text("errors.delveNotFound"))
	}
	return path, nil
}

// Start compiles with debug info and launches the program with the breakpoints.
// It returns once Delve is running; the rest of the start-up is asynchronous.
func (d *Debugger) Start(_ context.Context, config domain.RunConfiguration, breakpoints []domain.Breakpoint) error {
	d.mu.Lock()
	defer d.mu.Unlock()
	if d.current != nil {
		return app.ErrBusy
	}
	executable, err := d.delveExecutable()
	if err != nil {
		return err
	}
	environment := map[string]string(nil)
	if d.options.Environment != nil {
		environment = d.options.Environment()
	}
	removeStaleDebugBinaries(os.Getpid(), os.TempDir()) // best effort
	process, err := startAdapter(executable, workingDirectory(config), environment, d.text("errors.delveNotFound"))
	if err != nil {
		return err
	}
	d.book.Reset(breakpoints)
	d.tracker.Reset()
	started := newSession(d.sink, d.book, d.tracker, process)
	started.text = d.text
	started.onFinish = func() { d.ended(started) }
	d.current = started
	d.sink.DebugOutput(d.text("run.debugStdin")+"\n", "console")
	go started.begin(config, environment)
	return nil
}

func (d *Debugger) ended(finished *session) {
	d.mu.Lock()
	defer d.mu.Unlock()
	if d.current == finished {
		d.current = nil
	}
}

func (d *Debugger) active() *session {
	d.mu.Lock()
	defer d.mu.Unlock()
	return d.current
}

// SetBreakpoints replaces the breakpoints of one file, also while running.
func (d *Debugger) SetBreakpoints(file string, lines []int) error {
	if current := d.active(); current != nil {
		return current.setBreakpoints(file, lines)
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
	return current.execute(command)
}

// RunTo runs until the given location.
func (d *Debugger) RunTo(location domain.SourceLocation) error {
	current := d.active()
	if current == nil {
		return app.ErrNoSession
	}
	return current.runTo(location)
}

// RequestVariables asks for the children of a variable; the answer is debug:variables.
func (d *Debugger) RequestVariables(reference int) error {
	current := d.active()
	if current == nil {
		d.sink.DebugVariables(reference, []domain.Variable{})
		return nil
	}
	current.requestVariables(reference)
	return nil
}

// FrameVariables returns the arguments and locals of one frame of the paused stack.
func (d *Debugger) FrameVariables(frameID int) (domain.FrameVariables, error) {
	current := d.active()
	if current == nil {
		return domain.FrameVariables{Arguments: []domain.Variable{}, Locals: []domain.Variable{}}, nil
	}
	return current.frameVariables(frameID), nil
}

// Stop ends the session; debug:terminated carries domain.TerminatedByUser.
func (d *Debugger) Stop() error {
	current := d.active()
	if current == nil {
		return app.ErrNoSession
	}
	current.finish(domain.TerminatedByUser)
	return nil
}

// IsActive reports whether a session is alive.
func (d *Debugger) IsActive() bool { return d.active() != nil }
