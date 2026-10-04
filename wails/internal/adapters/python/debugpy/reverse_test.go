package debugpy

import (
	"context"
	"errors"
	"testing"

	"github.com/codeplai/VizcachaIDE/wails/internal/app"
	"github.com/codeplai/VizcachaIDE/wails/internal/domain"
	"github.com/codeplai/VizcachaIDE/wails/internal/protocol/process"
	"github.com/google/go-dap"
)

// fakeStarter records the job and plays its events like a supervisor would.
type fakeStarter struct {
	job process.Job
	err error
}

func (f *fakeStarter) Start(_ context.Context, job process.Job) error {
	f.job = job
	if f.err != nil {
		return f.err
	}
	job.Events.RunStarted(job.Config)
	job.Events.RunOutput("stdout", "hello\n")
	job.Events.RunOutput("stderr", "oops\n")
	return nil
}

func TestRunInTerminalForwardsOutputAsDebugOutput(t *testing.T) {
	sink := newEventSink()
	starter := &fakeStarter{}
	events := &debugEvents{sink: sink}
	handler := &reverseHandler{supervisor: starter, events: events, base: map[string]string{"A": "base", "B": "gone"}}

	args := dap.RunInTerminalRequestArguments{
		Args: []string{"python", "-m", "launcher", "prog.py"}, Cwd: "/work",
		Env: map[string]any{"A": "override", "B": nil, "C": "new"},
	}
	pid, err := handler.RunInTerminal(args)
	if err != nil || pid != 0 {
		t.Fatalf("pid=%d err=%v", pid, err)
	}
	job := starter.job
	if job.Mode != process.Terminal || job.Command != "python" || len(job.Args) != 3 || job.Dir != "/work" {
		t.Errorf("job = %+v", job)
	}
	if job.Env["A"] != "override" || job.Env["C"] != "new" {
		t.Errorf("env = %v", job.Env)
	}
	if _, kept := job.Env["B"]; kept {
		t.Error("a null value must remove the variable")
	}
	got := sink.debugEvents()
	if len(got) != 2 || got[0] != "stdout:hello\n" || got[1] != "stderr:oops\n" {
		t.Errorf("debug output = %q", got)
	}
	if sink.runEvents() != 0 {
		t.Error("the output must not become run:* events")
	}
	if !events.alive() {
		t.Error("the program should count as alive until RunFinished")
	}
	events.RunFinished(0, 1)
	if events.alive() {
		t.Error("RunFinished ends it")
	}
}

func TestRunInTerminalBusyAndEmpty(t *testing.T) {
	starter := &fakeStarter{err: app.ErrBusy}
	handler := &reverseHandler{supervisor: starter, events: &debugEvents{sink: newEventSink()}, config: domain.RunConfiguration{}}
	if _, err := handler.RunInTerminal(dap.RunInTerminalRequestArguments{Args: []string{"python"}}); !errors.Is(err, app.ErrBusy) {
		t.Errorf("err = %v, want ErrBusy", err)
	}
	if _, err := handler.RunInTerminal(dap.RunInTerminalRequestArguments{}); err == nil {
		t.Error("an empty command must fail")
	}
}
