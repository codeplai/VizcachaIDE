package delve

import (
	"errors"
	"path/filepath"
	"testing"

	"github.com/codeplai/VizcachaIDE/wails/internal/app"
	"github.com/codeplai/VizcachaIDE/wails/internal/domain"
	"github.com/google/go-dap"
)

func stopped(reason, description, text string) *dap.StoppedEvent {
	event := &dap.StoppedEvent{}
	event.Body = dap.StoppedEventBody{Reason: reason, Description: description, Text: text}
	return event
}

func TestStopReasons(t *testing.T) {
	cases := map[string]domain.StopReason{
		"breakpoint":          domain.StopBreakpoint,
		"function breakpoint": domain.StopBreakpoint,
		"step":                domain.StopStep,
		"entry":               domain.StopEntry,
		"exception":           domain.StopException,
		"pause":               domain.StopPause,
		"something new":       domain.StopPause,
	}
	for reason, want := range cases {
		if got, _ := (flavor{}).StopReason(stopped(reason, "", "")); got != want {
			t.Errorf("StopReason(%q) = %q, want %q", reason, got, want)
		}
	}
}

func TestPanicIsAnExceptionWithItsDescription(t *testing.T) {
	event := stopped("panic", "panic", `"runtime error: index out of range [3] with length 0"`)

	reason, description := flavor{}.StopReason(event)

	if want := "panic: runtime error: index out of range [3] with length 0"; description != want {
		t.Errorf("description = %q, want %q", description, want)
	}
	if reason != domain.StopException {
		t.Errorf("reason = %q, want exception", reason)
	}
}

func TestSubtleFramesAreDropped(t *testing.T) {
	if (flavor{}).KeepFrame(dap.StackFrame{PresentationHint: "subtle"}) || !(flavor{}).KeepFrame(dap.StackFrame{}) {
		t.Error("only subtle frames are dropped")
	}
}

func TestLocalsScopeEvenForOptimizedFunctions(t *testing.T) {
	for _, name := range []string{"Locals", "Locals (warning: optimized function)"} {
		if !(flavor{}).IsLocalsScope(name) {
			t.Errorf("%q is a locals scope", name)
		}
	}
	if (flavor{}).IsLocalsScope("Arguments") {
		t.Error("Arguments is not a locals scope")
	}
}

func TestOutputDropsDelvesNoise(t *testing.T) {
	event := func(category, text string) *dap.OutputEvent {
		return &dap.OutputEvent{Body: dap.OutputEventBody{Category: category, Output: text}}
	}
	if _, category, ok := (flavor{}).Output(event("stdout", "hi\n")); !ok || category != "stdout" {
		t.Errorf("stdout: ok=%v category=%q", ok, category)
	}
	if _, _, ok := (flavor{}).Output(event("console", "Type 'dlv help' for list of commands.\n")); ok {
		t.Error("the dlv banner must be dropped")
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
