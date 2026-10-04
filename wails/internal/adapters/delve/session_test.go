package delve

import (
	"errors"
	"path/filepath"
	"testing"

	"github.com/codeplai/VizcachaIDE/wails/internal/app"
	"github.com/codeplai/VizcachaIDE/wails/internal/domain"
	"github.com/google/go-dap"
)

func stoppedEvent(reason string) *dap.StoppedEvent {
	event := &dap.StoppedEvent{Event: dap.Event{ProtocolMessage: dap.ProtocolMessage{Type: "event"}, Event: "stopped"}}
	event.Body = dap.StoppedEventBody{Reason: reason, ThreadId: 1, AllThreadsStopped: true}
	return event
}

func variablesOf(state domain.DebugState) map[string]domain.Variable {
	byName := map[string]domain.Variable{}
	for _, variable := range state.Variables {
		byName[variable.Name] = variable
	}
	return byName
}

func TestSessionReplaysTheRecordedSession(t *testing.T) {
	dir := t.TempDir()
	messages := functionsSession(t, dir)
	server, conn := newFakeServer(t, messages)
	recorded := responseFor[*dap.VariablesResponse](t, messages).Body.Variables
	afterStep := append([]dap.Variable(nil), recorded...)
	afterStep[3].Value = "36" // result changes between the two stops
	server.variables = [][]dap.Variable{recorded, afterStep}
	sink := newEventSink()
	s := connectedSession(t, server, conn, sink)

	server.push(stoppedEvent("breakpoint"))
	first := receive(t, sink.stops)

	if first.Reason != domain.StopBreakpoint || first.Frames[0].Function != "main.multiply" {
		t.Fatalf("first stop = %+v", first)
	}
	if first.Frames[0].Location.File != filepath.Join(dir, "functions.go") || first.Frames[0].Location.Line != 13 {
		t.Errorf("location = %+v", first.Frames[0].Location)
	}
	values := variablesOf(first)
	if values["a"].Value != "5" || values["b"].Value != "7" || values["result"].Value != "35" {
		t.Errorf("variables = %+v", first.Variables)
	}
	for _, variable := range first.Variables {
		if variable.Changed {
			t.Errorf("%s must not be changed on the first stop", variable.Name)
		}
	}
	if len(first.Threads) != 6 || first.Threads[1].Location == nil || *first.CurrentThread != 1 {
		t.Errorf("goroutines = %+v", first.Threads)
	}

	server.after["next"] = []dap.Message{stoppedEvent("step")}
	if err := s.execute("next"); err != nil {
		t.Fatal(err)
	}
	second := receive(t, sink.stops)

	if second.Reason != domain.StopStep {
		t.Errorf("reason = %q, want step", second.Reason)
	}
	changed := variablesOf(second)
	if !changed["result"].Changed || changed["a"].Changed || changed["b"].Changed {
		t.Errorf("only result must be marked as changed: %+v", second.Variables)
	}
}

func TestTerminatedEventEndsTheSessionWithTheExitCode(t *testing.T) {
	server, conn := newFakeServer(t, functionsSession(t, t.TempDir()))
	sink := newEventSink()
	connectedSession(t, server, conn, sink)

	server.push(&dap.ExitedEvent{Event: dap.Event{ProtocolMessage: dap.ProtocolMessage{Type: "event"}, Event: "exited"}, Body: dap.ExitedEventBody{ExitCode: 3}})
	server.push(&dap.TerminatedEvent{Event: dap.Event{ProtocolMessage: dap.ProtocolMessage{Type: "event"}, Event: "terminated"}})

	if code := receive(t, sink.terminated); code != 3 {
		t.Errorf("exit code = %d, want 3", code)
	}
}

func TestStopByTheUserIsTerminatedByUserAndDisconnects(t *testing.T) {
	server, conn := newFakeServer(t, functionsSession(t, t.TempDir()))
	sink := newEventSink()
	s := connectedSession(t, server, conn, sink)

	s.finish(domain.TerminatedByUser)
	s.finish(0) // a second end is ignored

	if code := receive(t, sink.terminated); code != domain.TerminatedByUser {
		t.Errorf("exit code = %d, want %d", code, domain.TerminatedByUser)
	}
	if len(sink.terminated) != 0 {
		t.Error("debug:terminated must be emitted once")
	}
	if last := server.requests(); len(last) == 0 || last[len(last)-1] != "disconnect" {
		t.Errorf("requests = %v, want a final disconnect", last)
	}
}

func TestRequestVariablesAnswersAsynchronously(t *testing.T) {
	messages := functionsSession(t, t.TempDir())
	server, conn := newFakeServer(t, messages)
	server.variables = [][]dap.Variable{responseFor[*dap.VariablesResponse](t, messages).Body.Variables}
	sink := newEventSink()
	s := connectedSession(t, server, conn, sink)

	s.requestVariables(1001) // not paused: empty answer, always emitted
	if got := receive(t, sink.variables); len(got) != 0 {
		t.Errorf("while running the answer must be empty, got %+v", got)
	}

	server.push(stoppedEvent("breakpoint"))
	receive(t, sink.stops)
	s.requestVariables(1001)
	if got := receive(t, sink.variables); len(got) != 4 {
		t.Errorf("children = %+v, want 4 variables", got)
	}
}

func TestRunToUsesATemporaryBreakpointAndContinues(t *testing.T) {
	dir := t.TempDir()
	server, conn := newFakeServer(t, functionsSession(t, dir))
	s := connectedSession(t, server, conn, newEventSink())
	s.threadID = 1
	file := filepath.Join(dir, "functions.go")

	if err := s.runTo(domain.SourceLocation{File: file, Line: 36, Column: 1}); err != nil {
		t.Fatal(err)
	}

	requests := server.requests()
	if len(requests) != 2 || requests[0] != "setBreakpoints" || requests[1] != "continue" {
		t.Errorf("requests = %v, want setBreakpoints then continue", requests)
	}
	if lines := s.book.For(file); len(lines) != 1 || lines[0].Location.Line != 36 {
		t.Errorf("temporary breakpoint = %+v", lines)
	}
	s.book.ClearTemporary()
	if len(s.book.For(file)) != 0 {
		t.Error("the temporary breakpoint must go away after the stop")
	}
}

func TestControlsWithoutSessionReturnErrNoSession(t *testing.T) {
	debugger := New(newEventSink(), Options{})
	if err := debugger.StepOver(); !errors.Is(err, app.ErrNoSession) {
		t.Errorf("StepOver without session = %v, want ErrNoSession", err)
	}
	if debugger.IsActive() {
		t.Error("no session is active")
	}
}

func TestDelveNotFoundIsToolNotFound(t *testing.T) {
	debugger := New(newEventSink(), Options{DelvePath: func() string { return filepath.Join(t.TempDir(), "no-dlv") }})

	err := debugger.Start(t.Context(), domain.NewFileRunConfiguration(domain.CodeLanguageGo, "x/main.go", nil), nil)

	if !errors.Is(err, app.ErrToolNotFound) {
		t.Fatalf("err = %v, want ErrToolNotFound", err)
	}
}
