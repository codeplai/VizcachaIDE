package lldbdap

import (
	"os"
	"path/filepath"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/codeplai/VizcachaIDE/wails/internal/adapters/cpp"
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

// debuggerWith builds a debugger whose tools come from one bin folder (PATH only).
func debuggerWith(t *testing.T, bins ...string) (*Debugger, *eventSink, *process.Supervisor) {
	t.Helper()
	sink := newEventSink()
	supervisor := process.New(sink)
	locator := cpp.NewLocator(cpp.Options{
		AppDir: t.TempDir(), BaseEnvironment: []string{"PATH=" + strings.Join(bins, string(os.PathListSeparator))},
		XcodeReady: func() bool { return true },
	})
	return New(sink, supervisor, locator, os.Environ(), translate), sink, supervisor
}

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
	config := domain.NewFileRunConfiguration(domain.CodeLanguageCpp, path, nil)
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
	case <-time.After(60 * time.Second):
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
