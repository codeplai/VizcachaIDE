package lldbdap

import (
	"os"
	"path/filepath"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/codeplai/VizcachaIDE/wails/internal/adapters/rust"
	"github.com/codeplai/VizcachaIDE/wails/internal/adapters/rust/rusttest"
	"github.com/codeplai/VizcachaIDE/wails/internal/app"
	"github.com/codeplai/VizcachaIDE/wails/internal/domain"
	"github.com/codeplai/VizcachaIDE/wails/internal/protocol/process"
)

// eventSink records what reaches the frontend.
type eventSink struct {
	stops      chan domain.DebugState
	children   chan []domain.Variable
	terminated chan int

	mu      sync.Mutex
	debug   []string // "category:text"
	runs    int      // run:* events received
	outputs strings.Builder
}

var _ app.EventSink = (*eventSink)(nil)

func newEventSink() *eventSink {
	return &eventSink{children: make(chan []domain.Variable, 8), stops: make(chan domain.DebugState, 8), terminated: make(chan int, 4)}
}

func (e *eventSink) DebugStopped(state domain.DebugState) { e.stops <- state }
func (e *eventSink) DebugTerminated(code int)             { e.terminated <- code }
func (e *eventSink) DebugOutput(text, category string) {
	e.mu.Lock()
	defer e.mu.Unlock()
	e.debug = append(e.debug, category+":"+text)
	e.outputs.WriteString(text)
}
func (e *eventSink) DebugVariables(_ int, variables []domain.Variable)             { e.children <- variables }
func (e *eventSink) RunOutput(string, string)                                      { e.countRun() }
func (e *eventSink) RunStarted(domain.RunConfiguration)                            { e.countRun() }
func (e *eventSink) RunFinished(int, int64)                                        { e.countRun() }
func (e *eventSink) Diagnostics(string, []domain.Diagnostic)                       {}
func (e *eventSink) LanguageServerStatus(domain.CodeLanguage, domain.ServerStatus) {}
func (e *eventSink) Explained([]domain.ExplainedDiagnostic)                        {}
func (e *eventSink) SettingsChanged(domain.Settings)                               {}

func (e *eventSink) countRun() {
	e.mu.Lock()
	defer e.mu.Unlock()
	e.runs++
}

func (e *eventSink) text() string {
	e.mu.Lock()
	defer e.mu.Unlock()
	return e.outputs.String()
}

func (e *eventSink) debugEvents() []string {
	e.mu.Lock()
	defer e.mu.Unlock()
	return append([]string(nil), e.debug...)
}

func (e *eventSink) runEvents() int {
	e.mu.Lock()
	defer e.mu.Unlock()
	return e.runs
}

func translate(key string) string { return key }

// realDebugger builds a debugger with the real Rust and lldb-dap of the machine, or skips.
func realDebugger(t *testing.T) (*Debugger, *eventSink, *process.Supervisor) {
	t.Helper()
	env := rusttest.Environment(t)
	// lldb-dap is configured by path: llvm-mingw's bin must not reach rustc's PATH, where its
	// x86_64-w64-mingw32-gcc wrapper would replace the linker of the toolchain.
	settings := fakeSettings{rust.ToolLldbDap: rusttest.LldbDap(t)}
	sink := newEventSink()
	supervisor := process.New(sink)
	locator := rust.NewLocator(rust.Options{AppDir: t.TempDir(), BaseEnvironment: env, Settings: settings})
	return New(sink, supervisor, locator, translate), sink, supervisor
}

// fakeSettings is a SettingsStore with only tool paths.
type fakeSettings map[string]string

func (f fakeSettings) Load() (domain.Settings, error) {
	settings := domain.DefaultSettings()
	settings.ToolPaths = f
	return settings, nil
}

func (fakeSettings) Save(domain.Settings) error { return nil }

func writeProgram(t *testing.T, name, source string) string {
	t.Helper()
	path := filepath.Join(t.TempDir(), name)
	if err := os.WriteFile(path, []byte(source), 0o600); err != nil {
		t.Fatal(err)
	}
	return path
}

func start(t *testing.T, debugger *Debugger, path string, lines ...int) {
	t.Helper()
	var breakpoints []domain.Breakpoint
	for _, line := range lines {
		breakpoints = append(breakpoints, domain.Breakpoint{Location: domain.SourceLocation{File: path, Line: line, Column: 1}})
	}
	config := domain.NewFileRunConfiguration(domain.CodeLanguageRust, path, nil)
	if err := debugger.Start(t.Context(), config, breakpoints); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = debugger.Stop() })
}

func variablesOf(state domain.DebugState) map[string]domain.Variable {
	byName := map[string]domain.Variable{}
	for _, variable := range state.Variables {
		byName[variable.Name] = variable
	}
	return byName
}

func receive[T any](t *testing.T, channel <-chan T) T {
	t.Helper()
	select {
	case value := <-channel:
		return value
	case <-time.After(90 * time.Second):
		t.Fatal("timed out waiting for an event")
		var zero T
		return zero
	}
}

func waitFor(t *testing.T, sink *eventSink, text string) {
	t.Helper()
	deadline := time.Now().Add(60 * time.Second)
	for time.Now().Before(deadline) {
		if strings.Contains(sink.text(), text) {
			return
		}
		time.Sleep(50 * time.Millisecond)
	}
	t.Fatalf("%q never appeared in the output: %q", text, sink.text())
}
