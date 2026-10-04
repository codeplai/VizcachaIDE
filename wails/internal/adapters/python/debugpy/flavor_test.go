package debugpy

import (
	"encoding/json"
	"testing"

	"github.com/codeplai/VizcachaIDE/wails/internal/domain"
	"github.com/google/go-dap"
)

func TestLaunchJSON(t *testing.T) {
	config := domain.NewFileRunConfiguration(domain.CodeLanguagePython, "prog.py", []string{"a", "b"})
	raw, err := flavor{interpreter: "/py/python", console: consoleTerminal}.Launch(config, map[string]string{"X": "1"})
	if err != nil {
		t.Fatal(err)
	}
	var got map[string]any
	if err := json.Unmarshal(raw, &got); err != nil {
		t.Fatal(err)
	}
	want := map[string]any{
		"request": "launch", "type": "python", "console": "integratedTerminal",
		"justMyCode": true, "stopOnEntry": false, "redirectOutput": false,
	}
	for key, value := range want {
		if got[key] != value {
			t.Errorf("%s = %v, want %v", key, got[key], value)
		}
	}
	if python, _ := got["python"].([]any); len(python) != 1 || python[0] != "/py/python" {
		t.Errorf("python = %v", got["python"])
	}
	if args, _ := got["args"].([]any); len(args) != 2 {
		t.Errorf("args = %v", got["args"])
	}
	if env, _ := got["env"].(map[string]any); env["X"] != "1" {
		t.Errorf("env = %v", got["env"])
	}
}

func TestFilters(t *testing.T) {
	filters := flavor{}.ExceptionFilters()
	if len(filters) != 1 || filters[0] != "uncaught" {
		t.Errorf("filters = %v", filters)
	}
	if (flavor{}).AdapterID() != "python" {
		t.Error("adapter id")
	}
}

func TestHiddenVariables(t *testing.T) {
	hidden := []string{"special variables", "function variables", "class variables", "__name__", "__doc__"}
	for _, name := range hidden {
		if (flavor{}).KeepVariable(dap.Variable{Name: name}) {
			t.Errorf("%q should be hidden", name)
		}
	}
	for _, name := range []string{"n", "result", "_private", "__", "____x", "x__"} {
		if !(flavor{}).KeepVariable(dap.Variable{Name: name}) {
			t.Errorf("%q should be kept", name)
		}
	}
	if !(flavor{}).IsLocalsScope("Locals") || (flavor{}).IsLocalsScope("Globals") {
		t.Error("locals scope")
	}
}

func TestStopReasons(t *testing.T) {
	cases := []struct {
		reason, text string
		want         domain.StopReason
		description  string
	}{
		{"breakpoint", "", domain.StopBreakpoint, ""},
		{"step", "", domain.StopStep, ""},
		{"pause", "", domain.StopPause, ""},
		{"entry", "", domain.StopEntry, ""},
		{"exception", "NameError: name 'x' is not defined", domain.StopException, "NameError: name 'x' is not defined"},
	}
	for _, c := range cases {
		event := &dap.StoppedEvent{}
		event.Body.Reason, event.Body.Text = c.reason, c.text
		got, description := flavor{}.StopReason(event)
		if got != c.want || description != c.description {
			t.Errorf("%s: got %v %q", c.reason, got, description)
		}
	}
}

func TestExceptionDescriptionJoinsClassAndMessage(t *testing.T) {
	cases := [][2]string{
		{"NameError", "name 'x' is not defined"},
		{"NameError: name 'x' is not defined", "Paused on exception"},
	}
	for _, c := range cases {
		event := &dap.StoppedEvent{}
		event.Body.Reason, event.Body.Text, event.Body.Description = "exception", c[0], c[1]
		reason, description := flavor{}.StopReason(event)
		if reason != domain.StopException || description != "NameError: name 'x' is not defined" {
			t.Errorf("%v: got %v %q", c, reason, description)
		}
	}
}

func TestOutputCategories(t *testing.T) {
	cases := map[string]string{"stdout": "stdout", "stderr": "stderr", "console": "console", "important": "console"}
	for category, want := range cases {
		event := &dap.OutputEvent{}
		event.Body.Category, event.Body.Output = category, "text"
		if _, got, ok := (flavor{}).Output(event); !ok || got != want {
			t.Errorf("%s -> %q %v", category, got, ok)
		}
	}
	telemetry := &dap.OutputEvent{}
	telemetry.Body.Category, telemetry.Body.Output = "telemetry", "x"
	if _, _, ok := (flavor{}).Output(telemetry); ok {
		t.Error("telemetry must be dropped")
	}
}
