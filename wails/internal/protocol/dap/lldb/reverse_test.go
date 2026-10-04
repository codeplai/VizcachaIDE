package lldb

import (
	"context"
	"errors"
	"sync"
	"testing"

	"github.com/codeplai/VizcachaIDE/wails/internal/app"
	"github.com/codeplai/VizcachaIDE/wails/internal/domain"
	"github.com/codeplai/VizcachaIDE/wails/internal/protocol/process"
	"github.com/google/go-dap"
)

// recordingSink counts debug output and run events; the rest of the sink does nothing.
type recordingSink struct {
	app.EventSink // nil: the handler must not call anything else
	mu            sync.Mutex
	debug         []string
	runs          int
}

func (s *recordingSink) DebugOutput(text, category string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.debug = append(s.debug, category+":"+text)
}

func (s *recordingSink) RunOutput(string, string) { s.runs++ }

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
	sink := &recordingSink{}
	starter := &fakeStarter{}
	events := NewEvents(sink)
	handler := NewReverse(starter, events, domain.RunConfiguration{}, map[string]string{"A": "base", "B": "gone"})

	args := dap.RunInTerminalRequestArguments{
		Args: []string{"lldb-dap", "--comm-file", "pipe", "--launch-target", "prog.exe"}, Cwd: "/work",
		Env: map[string]any{"A": "override", "B": nil, "C": "new"},
	}
	pid, err := handler.RunInTerminal(args)
	if err != nil || pid != 0 {
		t.Fatalf("pid=%d err=%v", pid, err)
	}
	job := starter.job
	if job.Mode != process.Terminal || job.Command != "lldb-dap" || len(job.Args) != 4 || job.Dir != "/work" {
		t.Errorf("job = %+v", job)
	}
	if job.Env["A"] != "override" || job.Env["C"] != "new" {
		t.Errorf("env = %v", job.Env)
	}
	if _, has := job.Env["B"]; has {
		t.Error("a null value removes the variable")
	}
	if len(sink.debug) != 2 || sink.debug[0] != "stdout:hello\n" || sink.debug[1] != "stderr:oops\n" || sink.runs != 0 {
		t.Errorf("debug=%q runs=%d", sink.debug, sink.runs)
	}
	if !events.Alive() {
		t.Error("the program started and has not finished")
	}
	events.RunFinished(0, 1)
	if events.Alive() {
		t.Error("finished")
	}
}

func TestRunInTerminalErrors(t *testing.T) {
	handler := NewReverse(&fakeStarter{}, NewEvents(&recordingSink{}), domain.RunConfiguration{}, nil)
	if _, err := handler.RunInTerminal(dap.RunInTerminalRequestArguments{}); err == nil {
		t.Error("a request without a command must fail")
	}
	busy := NewReverse(&fakeStarter{err: app.ErrBusy}, NewEvents(&recordingSink{}), domain.RunConfiguration{}, nil)
	if _, err := busy.RunInTerminal(dap.RunInTerminalRequestArguments{Args: []string{"x"}}); !errors.Is(err, app.ErrBusy) {
		t.Errorf("err = %v", err)
	}
}
