package lldbdap

import (
	"context"
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/codeplai/VizcachaIDE/wails/internal/adapters/cpp"
	"github.com/codeplai/VizcachaIDE/wails/internal/adapters/cpp/cpptest"
	"github.com/codeplai/VizcachaIDE/wails/internal/app"
	"github.com/codeplai/VizcachaIDE/wails/internal/domain"
	"github.com/codeplai/VizcachaIDE/wails/internal/protocol/process"
)

// fakeTools makes a bin folder with empty files named like the tools; the probe accepts them.
func fakeTools(t *testing.T, names ...string) *cpp.Locator {
	t.Helper()
	bin := t.TempDir()
	for _, name := range names {
		if err := os.WriteFile(filepath.Join(bin, cpptest.Exe(name)), nil, 0o600); err != nil {
			t.Fatal(err)
		}
	}
	return cpp.NewLocator(cpp.Options{
		AppDir: t.TempDir(), BaseEnvironment: []string{"PATH=" + bin},
		Probe:      func(context.Context, string) (string, error) { return "clang version 23", nil },
		XcodeReady: func() bool { return true },
	})
}

func fakeDebugger(locator *cpp.Locator) (*Debugger, *eventSink) {
	sink := newEventSink()
	return New(sink, process.New(sink), locator, nil, translate), sink
}

var config = domain.NewFileRunConfiguration(domain.CodeLanguageCpp, filepath.Join("work", "main.cpp"), nil)

func TestMissingAdapterAndCompiler(t *testing.T) {
	debugger, _ := fakeDebugger(fakeTools(t))
	err := debugger.Start(t.Context(), config, nil)
	if !errors.Is(err, app.ErrToolNotFound) || !strings.Contains(err.Error(), `"lldb-dap"`) {
		t.Errorf("err = %v, want missing lldb-dap", err)
	}
	debugger, _ = fakeDebugger(fakeTools(t, "lldb-dap"))
	err = debugger.Start(t.Context(), config, nil)
	if !errors.Is(err, app.ErrToolNotFound) || !strings.Contains(err.Error(), `"cxx"`) {
		t.Errorf("err = %v, want missing cxx", err)
	}
}

type fakeCompiler struct {
	output string
	err    error
	gate   chan struct{} // when set, the compilation waits for it or for ctx
}

func (f fakeCompiler) CompileForDebug(ctx context.Context, _ domain.RunConfiguration) (string, string, error) {
	if f.gate != nil {
		select {
		case <-f.gate:
		case <-ctx.Done():
			return "", "", ctx.Err()
		}
	}
	return "", f.output, f.err
}

func TestCompileErrorEndsTheSession(t *testing.T) {
	debugger, sink := fakeDebugger(fakeTools(t, "lldb-dap", "clang++"))
	debugger.UseCompiler(fakeCompiler{output: "main.cpp:2:13: error: expected expression", err: ErrBuildFailed})
	if err := debugger.Start(t.Context(), config, nil); err != nil {
		t.Fatal(err)
	}
	if code := receive(t, sink.terminated); code != 1 {
		t.Errorf("exit code = %d, want 1", code)
	}
	events := sink.debugEvents()
	if len(events) != 1 || events[0] != "stderr:main.cpp:2:13: error: expected expression\n" {
		t.Errorf("debug output = %q", events)
	}
	if debugger.IsActive() {
		t.Error("the session must be over")
	}
}

func TestStopDuringCompilation(t *testing.T) {
	debugger, sink := fakeDebugger(fakeTools(t, "lldb-dap", "clang++"))
	debugger.UseCompiler(fakeCompiler{gate: make(chan struct{})})
	if err := debugger.Start(t.Context(), config, nil); err != nil {
		t.Fatal(err)
	}
	if !debugger.IsActive() {
		t.Error("compiling counts as active")
	}
	if err := debugger.Start(t.Context(), config, nil); !errors.Is(err, app.ErrBusy) {
		t.Errorf("second start: %v", err)
	}
	if err := debugger.Stop(); err != nil {
		t.Fatal(err)
	}
	if code := receive(t, sink.terminated); code != domain.TerminatedByUser {
		t.Errorf("exit code = %d", code)
	}
	if err := debugger.Stop(); !errors.Is(err, app.ErrNoSession) {
		t.Errorf("stop without a session: %v", err)
	}
}

func TestControlsWithoutSession(t *testing.T) {
	debugger, sink := fakeDebugger(fakeTools(t))
	if err := debugger.StepOver(); !errors.Is(err, app.ErrNoSession) {
		t.Errorf("StepOver: %v", err)
	}
	if err := debugger.SetBreakpoints("a.cpp", []int{3}); err != nil {
		t.Error(err)
	}
	if err := debugger.RequestVariables(7); err != nil {
		t.Error(err)
	}
	if got := receive(t, sink.children); len(got) != 0 {
		t.Errorf("variables = %v", got)
	}
}

func TestEnvironmentMap(t *testing.T) {
	env := environmentMap([]string{"A=1", "B=x=y", "=C:=C:\\work", "broken"})
	if env["A"] != "1" || env["B"] != "x=y" || env["=C:"] != "C:\\work" || len(env) != 3 {
		t.Errorf("env = %v", env)
	}
}
