package runner

import (
	"context"
	"errors"
	"strings"
	"testing"
	"time"

	"github.com/codeplai/VizcachaIDE/wails/internal/adapters/python"
	"github.com/codeplai/VizcachaIDE/wails/internal/app"
	"github.com/codeplai/VizcachaIDE/wails/internal/domain"
)

func sleep() { time.Sleep(25 * time.Millisecond) }

func TestToolsWithTheRealInterpreter(t *testing.T) {
	r, _ := realRunner(t)
	statuses := r.Tools(context.Background())
	if len(statuses) != 4 {
		t.Fatalf("statuses = %+v", statuses)
	}
	want := []string{python.ToolPython, python.ToolDebugpy, python.ToolPylsp, python.ToolRuff}
	for index, status := range statuses {
		if status.ID != want[index] || status.Version == "" || status.Path == "" || status.Source != domain.ToolConfigured {
			t.Errorf("status %d = %+v", index, status)
		}
		if status.CodeLanguage != domain.CodeLanguagePython {
			t.Errorf("status %d language = %v", index, status.CodeLanguage)
		}
	}
	if env := r.Environment(); env["PYTHONUTF8"] != "1" {
		t.Errorf("environment = %v", env)
	}
}

func TestRunReadsInputAndExits(t *testing.T) {
	r, sink := realRunner(t)
	path := writeProgram(t, "print('hello')\nname = input('Name: ')\nprint('bye', name)\n")
	config := r.Configure(path, nil)

	if err := r.Run(context.Background(), config); err != nil {
		t.Fatal(err)
	}
	if !r.IsRunning() {
		t.Fatal("the run slot should be busy")
	}
	sink.waitOutput(t, "Name:")
	if err := r.WriteInput("Ana"); err != nil {
		t.Fatal(err)
	}
	sink.waitOutput(t, "bye Ana")
	if code := sink.waitFinished(t); code != 0 {
		t.Fatalf("exit code %d: %s", code, sink.output())
	}
	if len(sink.started) != 1 || !sink.started[0].Echo {
		t.Fatalf("started = %+v", sink.started)
	}
	if r.IsRunning() {
		t.Fatal("the slot should be free")
	}
}

func TestRunPassesArgumentsAndExitCode(t *testing.T) {
	r, sink := realRunner(t)
	path := writeProgram(t, "import sys\nprint('args', sys.argv[1:])\nsys.exit(3)\n")

	if err := r.Run(context.Background(), r.Configure(path, []string{"one", "two words"})); err != nil {
		t.Fatal(err)
	}
	if code := sink.waitFinished(t); code != 3 {
		t.Fatalf("exit code %d: %s", code, sink.output())
	}
	if out := sink.output(); !strings.Contains(out, "['one', 'two words']") {
		t.Fatalf("output = %q", out)
	}
}

func TestSecondRunIsBusyAndStopEndsIt(t *testing.T) {
	r, sink := realRunner(t)
	path := writeProgram(t, "import time\nprint('up', flush=True)\ntime.sleep(60)\n")
	if err := r.Run(context.Background(), r.Configure(path, nil)); err != nil {
		t.Fatal(err)
	}
	sink.waitOutput(t, "up")
	if err := r.Run(context.Background(), r.Configure(path, nil)); !errors.Is(err, app.ErrBusy) {
		t.Fatalf("second Run = %v", err)
	}
	if err := r.Stop(); err != nil {
		t.Fatal(err)
	}
	sink.waitFinished(t)
}
