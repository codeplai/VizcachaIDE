package lldb

import (
	"encoding/json"
	"path/filepath"
	"testing"

	"github.com/codeplai/VizcachaIDE/wails/internal/domain"
	"github.com/google/go-dap"
)

func stopped(reason, description, text string) *dap.StoppedEvent {
	event := &dap.StoppedEvent{}
	event.Body.Reason, event.Body.Description, event.Body.Text = reason, description, text
	return event
}

func TestLaunchJSON(t *testing.T) {
	flavor := New(Options{Program: filepath.Join("work", "main.exe"), RunInTerminal: true, InitCommands: []string{"settings set x y"}})
	config := domain.NewFileRunConfiguration(domain.CodeLanguageCpp, "main.cpp", []string{"a", "b"})
	raw, err := flavor.Launch(config, map[string]string{"X": "1"})
	if err != nil {
		t.Fatal(err)
	}
	var got map[string]any
	if err := json.Unmarshal(raw, &got); err != nil {
		t.Fatal(err)
	}
	want := map[string]any{"request": "launch", "program": filepath.Join("work", "main.exe"), "cwd": "work", "stopOnEntry": false, "runInTerminal": true}
	for key, value := range want {
		if got[key] != value {
			t.Errorf("%s = %v, want %v", key, got[key], value)
		}
	}
	if args, _ := got["args"].([]any); len(args) != 2 {
		t.Errorf("args = %v", got["args"])
	}
	if env, _ := got["env"].(map[string]any); env["X"] != "1" {
		t.Errorf("env = %v", got["env"])
	}
	if commands, _ := got["initCommands"].([]any); len(commands) != 1 {
		t.Errorf("initCommands = %v", got["initCommands"])
	}
	if flavor.AdapterID() != "lldb" {
		t.Error("adapter id")
	}
}

func TestLaunchWithoutTerminalOrEnvironment(t *testing.T) {
	raw, err := New(Options{Program: "p", Dir: "/w"}).Launch(domain.RunConfiguration{}, nil)
	if err != nil {
		t.Fatal(err)
	}
	var got map[string]any
	_ = json.Unmarshal(raw, &got)
	if got["runInTerminal"] != false || got["cwd"] != "/w" {
		t.Errorf("launch = %v", got)
	}
	if _, has := got["env"]; has {
		t.Error("an empty environment must be omitted")
	}
	if args, ok := got["args"].([]any); !ok || len(args) != 0 {
		t.Errorf("args must be [] not null: %v", got["args"])
	}
}

func TestExceptionFilters(t *testing.T) {
	if filters := New(Options{}).ExceptionFilters(); len(filters) != 0 {
		t.Errorf("filters = %v", filters)
	}
	if filters := New(Options{CatchThrow: true}).ExceptionFilters(); len(filters) != 1 || filters[0] != "cpp_throw" {
		t.Errorf("filters = %v", filters)
	}
}

func TestStopReasons(t *testing.T) {
	cases := []struct {
		event       *dap.StoppedEvent
		reason      domain.StopReason
		description string
	}{
		{stopped("breakpoint", "", ""), domain.StopBreakpoint, ""},
		{stopped("step", "", ""), domain.StopStep, ""},
		{stopped("pause", "", ""), domain.StopPause, ""},
		{stopped("entry", "", ""), domain.StopEntry, ""},
		{stopped("exception", "signal SIGSEGV", ""), domain.StopException, "Segmentation fault: signal SIGSEGV"},
		{stopped("exception", "Exception 0xc0000005 encountered at address 0x1400", ""), domain.StopException, "Segmentation fault: Exception 0xc0000005 encountered at address 0x1400"},
		{stopped("exception", "EXC_BAD_ACCESS (code=1, address=0x0)", ""), domain.StopException, "Segmentation fault: EXC_BAD_ACCESS (code=1, address=0x0)"},
		{stopped("signal", "signal SIGFPE", ""), domain.StopException, "Floating point exception: signal SIGFPE"},
		{stopped("exception", "signal SIGABRT", ""), domain.StopException, "Aborted: signal SIGABRT"},
		{stopped("exception", "Exception 0xc00000fd encountered", ""), domain.StopException, "Stack overflow: Exception 0xc00000fd encountered"},
		{stopped("exception", "signal SIGUSR1", ""), domain.StopException, "signal SIGUSR1"},
		{stopped("exception", "", ""), domain.StopException, ""},
	}
	for _, c := range cases {
		reason, description := New(Options{}).StopReason(c.event)
		if reason != c.reason || description != c.description {
			t.Errorf("%+v: got (%s, %q), want (%s, %q)", c.event.Body, reason, description, c.reason, c.description)
		}
	}
}

func frame(path string) dap.StackFrame {
	if path == "" {
		return dap.StackFrame{Name: "ntdll"}
	}
	return dap.StackFrame{Name: "f", Source: &dap.Source{Path: path}}
}

func TestKeepFrameWithRoots(t *testing.T) {
	root := filepath.Join(t.TempDir(), "proj")
	flavor := New(Options{Roots: []string{root}})
	if !flavor.KeepFrame(frame(filepath.Join(root, "main.cpp"))) || !flavor.KeepFrame(frame(filepath.Join(root, "src", "a.cpp"))) {
		t.Error("user frames must be kept")
	}
	for _, path := range []string{"", filepath.Join(filepath.Dir(root), "proj2", "x.cpp"), "/mingw64/include/c++/string", filepath.Join(filepath.Dir(root), "other.cpp")} {
		if flavor.KeepFrame(frame(path)) {
			t.Errorf("%q must be hidden", path)
		}
	}
}

func TestKeepFrameWithoutRootsNeedsSource(t *testing.T) {
	flavor := New(Options{})
	if !flavor.KeepFrame(frame("/any/file.cpp")) || flavor.KeepFrame(frame("")) {
		t.Error("without roots only the source decides")
	}
}

func TestVariablesScopesAndOutput(t *testing.T) {
	flavor := New(Options{})
	if !flavor.KeepVariable(dap.Variable{Name: "__x"}) || !flavor.IsLocalsScope("Locals") || flavor.IsLocalsScope("Registers") {
		t.Error("variables and scopes")
	}
	output := func(category, text string) (string, string, bool) {
		event := &dap.OutputEvent{}
		event.Body.Category, event.Body.Output = category, text
		return flavor.Output(event)
	}
	if _, category, _ := output("stdout", "a"); category != "stdout" {
		t.Error("stdout passes")
	}
	if _, category, _ := output("stderr", "a"); category != "stderr" {
		t.Error("stderr passes")
	}
	if _, category, _ := output("important", "a"); category != "console" {
		t.Error("the rest is console")
	}
	if _, _, ok := output("telemetry", "a"); ok {
		t.Error("telemetry is dropped")
	}
}
