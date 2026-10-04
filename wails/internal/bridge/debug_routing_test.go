package bridge

import (
	"errors"
	"reflect"
	"sync"
	"testing"

	"github.com/codeplai/VizcachaIDE/wails/internal/app"
	"github.com/codeplai/VizcachaIDE/wails/internal/domain"
)

func TestDebugStartResolvesByPathAndKeepsOneSession(t *testing.T) {
	l := newTestLanguages(t)
	service := NewDebugService(l.registry)
	breakpoints := []domain.Breakpoint{{Condition: "i > 2"}}

	if err := service.Start("/w/main.go", breakpoints, `-n 3 "two words"`); err != nil {
		t.Fatal(err)
	}
	config := l.goDebugger.started[0]
	if want := []string{"-n", "3", "two words"}; !reflect.DeepEqual(config.ProgramArgs, want) || config.CodeLanguage != "go" {
		t.Errorf("config = %+v", config)
	}
	if len(l.goDebugger.breakons) != 1 {
		t.Errorf("breakpoints = %v", l.goDebugger.breakons)
	}

	// Another language cannot start while the Go session is alive.
	if err := service.Start("/w/tool.py", nil, ""); !errors.Is(err, app.ErrBusy) {
		t.Errorf("python start during a go session: %v", err)
	}
	if len(l.pyDebugger.started) != 0 {
		t.Error("the Python debugger must not have started")
	}

	if err := service.StepOver(); err != nil || l.goDebugger.steps != 1 {
		t.Errorf("StepOver: %v, steps = %d", err, l.goDebugger.steps)
	}
	if err := service.Stop(); err != nil || l.goDebugger.stopped != 1 {
		t.Errorf("Stop: %v", err)
	}
	if err := service.Start("/w/tool.py", nil, ""); err != nil || len(l.pyDebugger.started) != 1 {
		t.Errorf("python start after the session ended: %v", err)
	}
	if err := service.StepOver(); err != nil || l.pyDebugger.steps != 1 {
		t.Errorf("steps now go to the Python session: %v", err)
	}
}

func TestDebugStartRejectsUnclosedQuotesAndUnknownFiles(t *testing.T) {
	service := NewDebugService(newTestLanguages(t).registry)
	if err := service.Start("main.go", nil, `"oops`); err == nil {
		t.Error("an unclosed quote must be an error")
	}
	if err := service.Start("notes.txt", nil, ""); !errors.Is(err, app.ErrUnknownCodeLanguage) {
		t.Errorf("unknown file: %v", err)
	}
}

func TestDebugStartIsReservedWhileAnotherIsStarting(t *testing.T) {
	l := newTestLanguages(t)
	l.goDebugger.entered = make(chan struct{})
	l.goDebugger.blockOnce = make(chan struct{})
	service := NewDebugService(l.registry)

	var wg sync.WaitGroup
	wg.Add(1)
	var first error
	go func() {
		defer wg.Done()
		first = service.Start("a.go", nil, "")
	}()
	<-l.goDebugger.entered // the first Start is inside the adapter, the session is not active yet

	if err := service.Start("b.py", nil, ""); !errors.Is(err, app.ErrBusy) {
		t.Errorf("second start while the first is starting: %v", err)
	}
	close(l.goDebugger.blockOnce)
	wg.Wait()
	if first != nil {
		t.Errorf("first start: %v", first)
	}
}

func TestADebugStartThatFailsFreesTheSlot(t *testing.T) {
	l := newTestLanguages(t)
	l.goDebugger.startErr = app.MissingTool("dlv")
	service := NewDebugService(l.registry)
	if err := service.Start("a.go", nil, ""); !errors.Is(err, app.ErrToolNotFound) {
		t.Fatalf("start error = %v", err)
	}
	l.goDebugger.startErr = nil
	if err := service.Start("a.go", nil, ""); err != nil {
		t.Errorf("a failed start must not keep the slot: %v", err)
	}
}
