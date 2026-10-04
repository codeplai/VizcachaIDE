package debugpy

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/codeplai/VizcachaIDE/wails/internal/adapters/python"
	"github.com/codeplai/VizcachaIDE/wails/internal/adapters/python/pythontest"
	"github.com/codeplai/VizcachaIDE/wails/internal/domain"
	"github.com/codeplai/VizcachaIDE/wails/internal/protocol/process"
)

const factorialSource = `def factorial(n):
    result = 1
    for i in range(2, n + 1):
        result *= i
    return result


print(factorial(5))
`

type configuredPython struct{ path string }

func (c configuredPython) Load() (domain.Settings, error) {
	settings := domain.DefaultSettings()
	settings.ToolPaths = map[string]string{python.ToolPython: c.path}
	return settings, nil
}
func (configuredPython) Save(domain.Settings) error { return nil }

func translate(key string) string { return key }

// realDebugger builds a debugger on the real interpreter and a real supervisor.
func realDebugger(t *testing.T) (*Debugger, *eventSink, *process.Supervisor) {
	t.Helper()
	path := pythontest.Interpreter(t)
	sink := newEventSink()
	supervisor := process.New(sink)
	locator := python.NewLocator(python.Options{Settings: configuredPython{path: path}})
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
	config := domain.NewFileRunConfiguration(domain.CodeLanguagePython, path, nil)
	if err := debugger.Start(t.Context(), config, breakpoints); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = debugger.Stop() })
}

func TestRealDebugpyBreakpointStepAndVariables(t *testing.T) {
	debugger, sink, _ := realDebugger(t)
	path := writeProgram(t, "factorial.py", factorialSource)
	start(t, debugger, path, 4)

	first := receive(t, sink.stops)
	if first.Reason != domain.StopBreakpoint || first.Frames[0].Location.Line != 4 {
		t.Fatalf("first stop = %+v", first)
	}
	values := variablesOf(first)
	if values["n"].Value != "5" || values["result"].Value != "1" || values["i"].Value != "2" {
		t.Errorf("variables = %+v", first.Variables)
	}
	for name := range values {
		if strings.HasPrefix(name, "__") || strings.HasSuffix(name, " variables") {
			t.Errorf("hidden variable %q leaked", name)
		}
	}

	if err := debugger.StepOver(); err != nil {
		t.Fatal(err)
	}
	second := receive(t, sink.stops)
	if second.Reason != domain.StopStep || second.Frames[0].Location.Line != 3 {
		t.Fatalf("after the step: %+v", second)
	}
	if got := variablesOf(second)["result"].Value; got != "2" {
		t.Errorf("result after the step = %q, want 2", got)
	}
	if sink.runEvents() != 0 {
		t.Error("debugging must not emit run:* events")
	}
}

func TestRealDebugpyUncaughtException(t *testing.T) {
	debugger, sink, _ := realDebugger(t)
	path := writeProgram(t, "boom.py", "total = 1\nprint(x)\n")
	start(t, debugger, path)

	stop := receive(t, sink.stops)
	if stop.Reason != domain.StopException || stop.Description != "NameError: name 'x' is not defined" {
		t.Fatalf("stop = %+v", stop)
	}
	if stop.Frames[0].Location.Line != 2 {
		t.Errorf("stopped at line %d, want 2", stop.Frames[0].Location.Line)
	}
}

func TestRealDebugpyInputThroughSupervisor(t *testing.T) {
	debugger, sink, supervisor := realDebugger(t)
	path := writeProgram(t, "hello.py", "name = input('Name: ')\nprint('Hello ' + name)\n")
	start(t, debugger, path)

	waitFor(t, sink, "Name:")
	if err := supervisor.WriteInput("Ana"); err != nil {
		t.Fatal(err)
	}
	waitFor(t, sink, "Hello Ana")
	if code := receive(t, sink.terminated); code != 0 {
		t.Errorf("exit code = %d", code)
	}
	if sink.runEvents() != 0 {
		t.Error("the program output must not become run:* events")
	}
}

func TestRealDebugpyWithoutTerminalUsesInternalConsole(t *testing.T) {
	debugger, sink, _ := realDebugger(t)
	debugger.hasTerminal = func(string) bool { return false }
	path := writeProgram(t, "factorial.py", factorialSource)
	start(t, debugger, path, 4)

	first := receive(t, sink.stops)
	if first.Reason != domain.StopBreakpoint || variablesOf(first)["n"].Value != "5" {
		t.Fatalf("first stop = %+v", first)
	}
	if !strings.Contains(sink.text(), "run.debugStdin") {
		t.Errorf("the no-keyboard notice is missing: %q", sink.debugEvents())
	}
}

func TestMissingInterpreter(t *testing.T) {
	sink := newEventSink()
	missing := configuredPython{path: filepath.Join(t.TempDir(), "nope")}
	locator := python.NewLocator(python.Options{
		Settings: missing, AppDir: t.TempDir(), BaseEnvironment: []string{"PATH="},
		Launcher: func(context.Context) string { return "" },
	})
	debugger := New(sink, process.New(sink), locator, nil, translate)
	config := domain.NewFileRunConfiguration(domain.CodeLanguagePython, "x.py", nil)
	err := debugger.Start(t.Context(), config, nil)
	if err == nil || !strings.Contains(err.Error(), `"python"`) {
		t.Errorf("err = %v, want missing python", err)
	}
}

func waitFor(t *testing.T, sink *eventSink, text string) {
	t.Helper()
	deadline := time.Now().Add(30 * time.Second)
	for time.Now().Before(deadline) {
		if strings.Contains(sink.text(), text) {
			return
		}
		time.Sleep(50 * time.Millisecond)
	}
	t.Fatalf("%q never appeared in the output: %q", text, sink.text())
}
