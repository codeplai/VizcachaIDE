// Package lldbdap implements app.Debugger for C++ with lldb-dap, the LLVM debug adapter, over
// stdio. The program is compiled with debug information first and runs in a pseudoterminal on
// the shared supervisor (DAP runInTerminal), so std::cin works while debugging. The DAP flavor
// is shared with Rust: protocol/dap/lldb.
package lldbdap

import (
	"context"
	"path/filepath"
	"sync"

	"github.com/codeplai/VizcachaIDE/wails/internal/adapters/cpp"
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
	locator    *cpp.Locator
	compiler   Compiler
	base       []string
	text       func(key string) string
	book       *app.BreakpointBook
	tracker    *app.ChangeTracker
	// hasTerminal says whether a pseudoterminal works with the adapter; tests replace it.
	hasTerminal func(adapter string) bool

	mu      sync.Mutex
	current *protodap.Session
	cancel  context.CancelFunc // set while the program is being compiled
}

var _ app.Debugger = (*Debugger)(nil)

// New creates a debugger that reports through sink. The program runs on supervisor, which is
// shared with the runner, so only one program runs at a time. base is the "NAME=value"
// environment of the program; translate returns the text of an i18n key.
func New(sink app.EventSink, supervisor *process.Supervisor, locator *cpp.Locator, base []string, translate func(key string) string) *Debugger {
	return &Debugger{
		sink: sink, supervisor: supervisor, locator: locator, base: base, text: translate,
		compiler: buildCompiler{locator: locator},
		book:     app.NewBreakpointBook(), tracker: app.NewChangeTracker(), hasTerminal: terminalWorks,
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
	adapter := d.locator.Tool(ctx, cpp.ToolLldbDap)
	if adapter.Source == domain.ToolMissing {
		return app.MissingTool(cpp.ToolLldbDap)
	}
	if _, err := d.locator.Compiler(ctx); err != nil {
		return err
	}
	d.book.Reset(breakpoints)
	d.tracker.Reset()
	building, cancel := context.WithCancel(context.Background())
	d.cancel = cancel
	go d.build(building, adapter.Path, config)
	return nil
}

// build compiles, then starts the session.
func (d *Debugger) build(ctx context.Context, adapter string, config domain.RunConfiguration) {
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
		d.launch(adapter, exe, config)
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
func (d *Debugger) launch(adapter, exe string, config domain.RunConfiguration) {
	terminal := d.hasTerminal(adapter)
	folder := projectFolder(config)
	environment := environmentMap(d.base)
	events := lldb.NewEvents(d.sink)
	started := protodap.NewSession(protodap.SessionDeps{
		Sink: d.sink, Book: d.book, Tracker: d.tracker,
		Transport: &protodap.StdioTransport{Command: adapter, Dir: folder, Env: d.base},
		Flavor:    lldb.New(lldb.Options{Program: exe, Dir: workingDirectory(config), Roots: []string{folder}, RunInTerminal: terminal}),
		Reverse:   lldb.NewReverse(d.supervisor, events, config, environment),
		Texts:     d.text,
	})
	started.OnFinish(func() { d.ended(started, events) })
	d.current = started
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

// projectFolder is the folder with the user's source: the project, or the folder of the file.
func projectFolder(config domain.RunConfiguration) string {
	folder := config.Target
	if config.Mode != domain.RunProject {
		folder = filepath.Dir(config.Target)
	}
	if absolute, err := filepath.Abs(folder); err == nil {
		return absolute
	}
	return folder
}

func workingDirectory(config domain.RunConfiguration) string {
	if config.WorkingDir != "" {
		return config.WorkingDir
	}
	return projectFolder(config)
}
