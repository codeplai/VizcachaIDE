package debugpy

import (
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/codeplai/VizcachaIDE/wails/internal/app"
	"github.com/codeplai/VizcachaIDE/wails/internal/domain"
)

// eventSink records what reaches the frontend.
type eventSink struct {
	stops      chan domain.DebugState
	terminated chan int

	mu      sync.Mutex
	debug   []string // "category:text"
	runs    int      // run:* events received
	outputs strings.Builder
}

var _ app.EventSink = (*eventSink)(nil)

func newEventSink() *eventSink {
	return &eventSink{stops: make(chan domain.DebugState, 8), terminated: make(chan int, 4)}
}

func (e *eventSink) DebugStopped(state domain.DebugState) { e.stops <- state }
func (e *eventSink) DebugTerminated(code int)             { e.terminated <- code }
func (e *eventSink) DebugOutput(text, category string) {
	e.mu.Lock()
	defer e.mu.Unlock()
	e.debug = append(e.debug, category+":"+text)
	e.outputs.WriteString(text)
}
func (e *eventSink) DebugVariables(int, []domain.Variable) {}
func (e *eventSink) RunOutput(string, string)              { e.countRun() }
func (e *eventSink) RunStarted(domain.RunConfiguration)    { e.countRun() }
func (e *eventSink) RunFinished(int, int64)                { e.countRun() }
func (e *eventSink) Diagnostics(string, []domain.Diagnostic) {
}
func (e *eventSink) LanguageServerStatus(domain.ServerStatus) {}
func (e *eventSink) Explained([]domain.ExplainedDiagnostic)   {}
func (e *eventSink) SettingsChanged(domain.Settings)          {}

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
	case <-time.After(30 * time.Second):
		t.Fatal("timed out waiting for an event")
		var zero T
		return zero
	}
}
