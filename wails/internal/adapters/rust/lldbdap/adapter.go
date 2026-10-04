// Package lldbdap implements app.Debugger for Rust with lldb-dap, the LLVM debug adapter, over
// stdio. The program is compiled with debug information first and runs in a pseudoterminal on
// the shared supervisor (DAP runInTerminal), so stdin works while debugging. The DAP flavor is
// the shared one of protocol/dap/lldb plus what is particular to Rust: the formatters of std
// types, hidden library frames and a stop at panic!.
package lldbdap

import (
	"context"
	"sync"

	"github.com/codeplai/VizcachaIDE/wails/internal/adapters/rust"
	"github.com/codeplai/VizcachaIDE/wails/internal/app"
	"github.com/codeplai/VizcachaIDE/wails/internal/domain"
	protodap "github.com/codeplai/VizcachaIDE/wails/internal/protocol/dap"
	"github.com/codeplai/VizcachaIDE/wails/internal/protocol/dap/lldb"
	"github.com/codeplai/VizcachaIDE/wails/internal/protocol/process"
)

// Debugger is the lldb-dap implementation of app.Debugger.
type Debugger struct {
	sink       app.EventSink
	supervisor *process.Supervisor
	locator    *rust.Locator
	compiler   Compiler
	text       func(key string) string
	book       *app.BreakpointBook
	tracker    *app.ChangeTracker
	// hasTerminal and hasPython say whether a pseudoterminal and the Python of LLDB work with the
	// adapter; tests replace them.
	hasTerminal func(adapter string) bool
	hasPython   func(adapter string) bool

	mu      sync.Mutex
	current *protodap.Session
	cancel  context.CancelFunc // set while the program is being compiled
}

var _ app.Debugger = (*Debugger)(nil)

// New creates a debugger that reports through sink. The program runs on supervisor, which is
// shared with the runner, so only one program runs at a time; its environment is
// locator.Environment(). translate returns the text of an i18n key.
func New(sink app.EventSink, supervisor *process.Supervisor, locator *rust.Locator, translate func(key string) string) *Debugger {
	return &Debugger{
		sink: sink, supervisor: supervisor, locator: locator, text: translate,
		compiler: buildCompiler{locator: locator},
		book:     app.NewBreakpointBook(), tracker: app.NewChangeTracker(),
		hasTerminal: terminalWorks, hasPython: pythonWorks,
	}
}

// UseCompiler replaces the build step (the runner's own, for instance). Call it before Start.
func (d *Debugger) UseCompiler(compiler Compiler) { d.compiler = compiler }

// Start compiles the program with debug information and launches it under lldb-dap with the
// breakpoints. It returns once the tools are found; the compilation and the rest of the
// start-up are asynchronous. A compile error ends the session: the compiler output arrives as
// debug:output (stderr) followed by debug:terminated.
func (d *Debugger) Start(ctx context.Context, config domain.RunConfiguration, breakpoints []domain.Breakpoint) error {
	d.mu.Lock()
	defer d.mu.Unlock()
	if d.current != nil || d.cancel != nil || d.supervisor.IsRunning() {
		return app.ErrBusy
	}
	adapter := d.locator.Tool(rust.ToolLldbDap)
	if adapter.Source == domain.ToolMissing {
		return app.MissingTool(rust.ToolLldbDap)
	}
	toolchain, err := d.locator.Toolchain(ctx)
	if err != nil {
		return err
	}
	d.book.Reset(breakpoints)
	d.tracker.Reset()
	building, cancel := context.WithCancel(context.Background())
	d.cancel = cancel
	go d.build(building, adapter.Path, toolchain, config)
	return nil
}

// build compiles, then starts the session.
func (d *Debugger) build(ctx context.Context, adapter string, toolchain rust.Toolchain, config domain.RunConfiguration) {
	exe, output, err := d.compiler.CompileForDebug(ctx, config)
	d.mu.Lock()
	defer d.mu.Unlock()
	d.cancel = nil
	switch {
	case ctx.Err() != nil:
		d.sink.DebugTerminated(domain.TerminatedByUser)
	case err != nil:
		d.buildFailed(output, err)
	default:
		d.launch(adapter, exe, toolchain, config)
	}
}

// buildFailed shows the compiler's words (the Assistant explains them) and ends the session.
func (d *Debugger) buildFailed(output string, err error) {
	if output == "" {
		output = err.Error()
	}
	d.sink.DebugOutput(ensureNewline(output), "stderr")
	d.sink.DebugTerminated(1)
}

func ensureNewline(text string) string {
	if text == "" || text[len(text)-1] == '\n' {
		return text
	}
	return text + "\n"
}

// launch creates the session; d.mu is held.
func (d *Debugger) launch(adapter, exe string, toolchain rust.Toolchain, config domain.RunConfiguration) {
	terminal, python := d.hasTerminal(adapter), d.hasPython(adapter)
	folder := sourceRoot(config)
	environment := d.locator.Environment()
	panics := &panicWatch{EventSink: d.sink}
	events := lldb.NewEvents(panics)
	options := lldb.Options{
		Program: exe, Dir: workingDirectory(config), Roots: []string{folder}, RunInTerminal: terminal,
		InitCommands: InitCommands(toolchain, python, enumsFixPath()),
	}
	started := protodap.NewSession(protodap.SessionDeps{
		Sink: d.sink, Book: d.book, Tracker: d.tracker,
		Transport: &protodap.StdioTransport{Command: adapter, Dir: folder, Env: environment},
		Flavor:    newFlavor(options, panics),
		Reverse:   lldb.NewReverse(d.supervisor, events, config, environmentMap(environment)),
		Texts:     d.text,
	})
	started.OnFinish(func() { d.ended(started, events) })
	d.current = started
	if !python {
		d.sink.DebugOutput(d.text("errors.rustLldbNoPython")+"\n", "console")
	}
	if !terminal {
		d.sink.DebugOutput(d.text("run.debugStdin")+"\n", "console")
	}
	go started.Begin(config, nil)
}

// ended forgets the session and stops the program if the adapter left it running.
func (d *Debugger) ended(finished *protodap.Session, events *lldb.Events) {
	if events.Alive() {
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
