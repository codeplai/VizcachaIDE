package delve

import (
	"testing"
	"time"

	"github.com/codeplai/VizcachaIDE/wails/internal/app"
	"github.com/codeplai/VizcachaIDE/wails/internal/domain"
)

// eventSink records what reaches the frontend.
type eventSink struct {
	stops      chan domain.DebugState
	terminated chan int
}

var _ app.EventSink = (*eventSink)(nil)

func newEventSink() *eventSink {
	return &eventSink{stops: make(chan domain.DebugState, 8), terminated: make(chan int, 4)}
}

func (e *eventSink) DebugStopped(state domain.DebugState)     { e.stops <- state }
func (e *eventSink) DebugTerminated(code int)                 { e.terminated <- code }
func (e *eventSink) DebugOutput(string, string)               {}
func (e *eventSink) DebugVariables(int, []domain.Variable)    {}
func (e *eventSink) RunOutput(string, string)                 {}
func (e *eventSink) RunStarted(domain.RunConfiguration)       {}
func (e *eventSink) RunFinished(int, int64)                   {}
func (e *eventSink) Diagnostics(string, []domain.Diagnostic)  {}
func (e *eventSink) LanguageServerStatus(domain.ServerStatus) {}
func (e *eventSink) Explained([]domain.ExplainedDiagnostic)   {}
func (e *eventSink) SettingsChanged(domain.Settings)          {}

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
	// Generous: the real-Delve tests compile a program first, which is slow on a busy machine.
	case <-time.After(30 * time.Second):
		t.Fatal("timed out waiting for an event")
		var zero T
		return zero
	}
}
