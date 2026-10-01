package delve

import (
	"os"
	"os/exec"
	"path/filepath"
	"testing"

	"github.com/codeplai/VizcachaIDE/wails/internal/domain"
)

// findDelve returns the dlv executable, or skips the test when it is not installed.
func findDelve(t *testing.T) string {
	t.Helper()
	if _, err := exec.LookPath("go"); err != nil {
		t.Skip("go is not installed")
	}
	if path, err := exec.LookPath("dlv"); err == nil {
		return path
	}
	output, err := exec.Command("go", "env", "GOPATH").Output()
	if err == nil {
		candidate := filepath.Join(string(output[:len(output)-1]), "bin", "dlv")
		if path, lookErr := exec.LookPath(candidate); lookErr == nil {
			return path
		}
	}
	t.Skip("dlv is not installed")
	return ""
}

// copyExample puts examples/functions/functions.go in a temporary folder and returns its path.
func copyExample(t *testing.T) string {
	t.Helper()
	source, err := os.ReadFile(filepath.Join("..", "..", "..", "..", "examples", "functions", "functions.go"))
	if err != nil {
		t.Skipf("examples/functions/functions.go not found: %v", err)
	}
	path := filepath.Join(t.TempDir(), "functions.go")
	if err := os.WriteFile(path, source, 0o600); err != nil {
		t.Fatal(err)
	}
	return path
}

func TestRealDelveDebugsFunctionsExample(t *testing.T) {
	dlv := findDelve(t)
	path := copyExample(t)
	sink := newEventSink()
	debugger := New(sink, Options{DelvePath: func() string { return dlv }})
	breakpoint := domain.Breakpoint{Location: domain.SourceLocation{File: path, Line: 13, Column: 1}}

	err := debugger.Start(t.Context(), domain.NewFileRunConfiguration(path, nil), []domain.Breakpoint{breakpoint})
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = debugger.Stop() })

	first := receive(t, sink.stops)
	if first.Reason != domain.StopBreakpoint || first.Frames[0].Function != "main.multiply" {
		t.Fatalf("first stop = %+v", first)
	}
	if first.Frames[0].Location.Line != 13 {
		t.Errorf("stopped at line %d, want 13", first.Frames[0].Location.Line)
	}
	values := variablesOf(first)
	if values["a"].Value != "5" || values["b"].Value != "7" || values["result"].Value != "35" {
		t.Errorf("variables = %+v", first.Variables)
	}

	if err := debugger.StepOver(); err != nil {
		t.Fatal(err)
	}
	second := receive(t, sink.stops)
	if second.Reason != domain.StopStep || second.Frames[0].Location.Line == 13 {
		t.Errorf("after the step: %+v", second)
	}

	if err := debugger.Stop(); err != nil {
		t.Fatal(err)
	}
	if code := receive(t, sink.terminated); code != domain.TerminatedByUser {
		t.Errorf("exit code = %d, want %d", code, domain.TerminatedByUser)
	}
	if debugger.IsActive() {
		t.Error("the session must be over")
	}
}

func TestRealDelveRunsToTheEnd(t *testing.T) {
	dlv := findDelve(t)
	path := copyExample(t)
	sink := newEventSink()
	debugger := New(sink, Options{DelvePath: func() string { return dlv }})

	if err := debugger.Start(t.Context(), domain.NewFileRunConfiguration(path, nil), nil); err != nil {
		t.Fatal(err)
	}

	if code := receive(t, sink.terminated); code != 0 {
		t.Errorf("exit code = %d, want 0", code)
	}
}
