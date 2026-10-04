// Package debugpy implements app.Debugger with a DAP client talking to "python -m debugpy.adapter"
// over stdio. The debugged program runs in a pseudoterminal on the shared supervisor (DAP
// runInTerminal), so input() works while debugging.
package debugpy

import (
	"context"
	"sync"

	"github.com/codeplai/VizcachaIDE/wails/internal/adapters/python"
	"github.com/codeplai/VizcachaIDE/wails/internal/app"
	"github.com/codeplai/VizcachaIDE/wails/internal/domain"
	protodap "github.com/codeplai/VizcachaIDE/wails/internal/protocol/dap"
	"github.com/codeplai/VizcachaIDE/wails/internal/protocol/process"
)

// Debugger is the debugpy implementation of app.Debugger.
type Debugger struct {
	sink       app.EventSink
	supervisor *process.Supervisor
	locator    *python.Locator
	base       []string
	text       func(key string) string
	book       *app.BreakpointBook
	tracker    *app.ChangeTracker
	// hasTerminal says whether a pseudoterminal works with the interpreter; tests replace it.
	hasTerminal func(interpreter string) bool

	mu      sync.Mutex
	current *protodap.Session
}

var _ app.Debugger = (*Debugger)(nil)

// New creates a debugger that reports through sink. The program runs on supervisor, which is
// shared with the runner, so only one program runs at a time. base is the "NAME=value"
// environment Python starts from; translate returns the text of an i18n key.
func New(sink app.EventSink, supervisor *process.Supervisor, locator *python.Locator, base []string, translate func(key string) string) *Debugger {
	return &Debugger{
		sink: sink, supervisor: supervisor, locator: locator, base: base, text: translate,
		book: app.NewBreakpointBook(), tracker: app.NewChangeTracker(), hasTerminal: terminalWorks,
	}
}

// Start launches the program under debugpy with the breakpoints. It returns once the adapter is
// running; the rest of the start-up is asynchronous.
func (d *Debugger) Start(ctx context.Context, config domain.RunConfiguration, breakpoints []domain.Breakpoint) error {
	d.mu.Lock()
	defer d.mu.Unlock()
	if d.current != nil || d.supervisor.IsRunning() {
		return app.ErrBusy
	}
	interpreter, err := d.locator.Find(ctx, workingDirectory(config))
	if err != nil {
		return err
	}
	environment := python.Environment(d.base, interpreter)
	if err := requireDebugpy(ctx, interpreter, environment); err != nil {
		return err
	}
	console, terminal := consoleInternal, d.hasTerminal(interpreter.Path)
	if terminal {
		console = consoleTerminal
	}
	d.book.Reset(breakpoints)
	d.tracker.Reset()
	events := &debugEvents{sink: d.sink}
	started := protodap.NewSession(protodap.SessionDeps{
		Sink: d.sink, Book: d.book, Tracker: d.tracker,
		Transport: adapterTransport(interpreter, environment, config),
		Flavor:    flavor{interpreter: interpreter.Path, console: console},
		Reverse:   &reverseHandler{supervisor: d.supervisor, events: events, config: config, base: environment},
		Texts:     d.text,
	})
	started.OnFinish(func() { d.ended(started, events) })
	d.current = started
	if !terminal {
		d.sink.DebugOutput(d.text("run.debugStdin")+"\n", "console")
	}
	go started.Begin(config, environment)
	return nil
}

// ended forgets the session and stops the program if the adapter left it running.
func (d *Debugger) ended(finished *protodap.Session, events *debugEvents) {
	if events.alive() {
		_ = d.supervisor.Stop()
	}
	d.mu.Lock()
	defer d.mu.Unlock()
	if d.current == finished {
		d.current = nil
	}
}

func (d *Debugger) active() *protodap.Session {
	d.mu.Lock()
	defer d.mu.Unlock()
	return d.current
}

// requireDebugpy fails with app.MissingTool("debugpy") when the interpreter lacks the module.
func requireDebugpy(ctx context.Context, interpreter python.Interpreter, environment map[string]string) error {
	versions := python.ModuleVersions(ctx, interpreter, environment)
	if versions == nil {
		return app.MissingTool(python.ToolPython)
	}
	if versions[python.ToolDebugpy] == "" {
		return app.MissingTool(python.ToolDebugpy)
	}
	return nil
}

func adapterTransport(interpreter python.Interpreter, environment map[string]string, config domain.RunConfiguration) *protodap.StdioTransport {
	return &protodap.StdioTransport{
		Command: interpreter.Path, Args: []string{"-m", "debugpy.adapter"},
		Dir: workingDirectory(config), Env: python.EnvironmentList(environment),
	}
}
