package bridge

import (
	"testing"

	"github.com/codeplai/VizcachaIDE/wails/internal/app"
	"github.com/codeplai/VizcachaIDE/wails/internal/domain"
)

type recordingSink struct {
	calls []string
}

func (r *recordingSink) RunOutput(string, string)           { r.calls = append(r.calls, "output") }
func (r *recordingSink) RunStarted(domain.RunConfiguration) { r.calls = append(r.calls, "started") }
func (r *recordingSink) RunFinished(int, int64)             { r.calls = append(r.calls, "finished") }
func (r *recordingSink) DebugStopped(domain.DebugState)     { r.calls = append(r.calls, "stopped") }
func (r *recordingSink) DebugVariables(int, []domain.Variable) {
	r.calls = append(r.calls, "variables")
}
func (r *recordingSink) DebugOutput(string, string) { r.calls = append(r.calls, "debugOutput") }
func (r *recordingSink) DebugTerminated(int)        { r.calls = append(r.calls, "terminated") }
func (r *recordingSink) Diagnostics(string, []domain.Diagnostic) {
	r.calls = append(r.calls, "diagnostics")
}
func (r *recordingSink) LanguageServerStatus(domain.ServerStatus) {
	r.calls = append(r.calls, "status")
}
func (r *recordingSink) Explained([]domain.ExplainedDiagnostic) {
	r.calls = append(r.calls, "explained")
}
func (r *recordingSink) SettingsChanged(domain.Settings) { r.calls = append(r.calls, "settings") }

func TestRunServiceEmitsStartOutputFinish(t *testing.T) {
	sink := &recordingSink{}
	if _, err := NewRunService(sink).Run("hola-go/main.go", nil); err != nil {
		t.Fatal(err)
	}
	want := []string{"started", "output", "finished"}
	if len(sink.calls) != len(want) {
		t.Fatalf("calls = %v, want %v", sink.calls, want)
	}
	for i := range want {
		if sink.calls[i] != want[i] {
			t.Fatalf("calls = %v, want %v", sink.calls, want)
		}
	}
}

// stoppingDebugger is an app.Debugger whose Stop reports the user's stop, like the real one.
type stoppingDebugger struct {
	app.Debugger
	sink app.EventSink
}

func (d stoppingDebugger) Stop() error {
	d.sink.DebugTerminated(domain.TerminatedByUser)
	return nil
}

func TestDebugServiceStopIsTerminatedByUser(t *testing.T) {
	sink := &recordingSink{}
	if err := NewDebugService(stoppingDebugger{sink: sink}).Stop(); err != nil {
		t.Fatal(err)
	}
	if len(sink.calls) != 1 || sink.calls[0] != "terminated" {
		t.Errorf("calls = %v, want [terminated]", sink.calls)
	}
}

func TestSettingsServiceNotifiesOnSave(t *testing.T) {
	sink := &recordingSink{}
	service := NewSettingsService(sink, NewMemorySettingsStore())
	if err := service.Save(domain.DefaultSettings()); err != nil {
		t.Fatal(err)
	}
	if len(sink.calls) != 1 || sink.calls[0] != "settings" {
		t.Errorf("calls = %v, want [settings]", sink.calls)
	}
}

func TestFilesServiceReadsSampleFiles(t *testing.T) {
	service := NewFilesService()
	if text, err := service.ReadFile(sampleMain); err != nil || text == "" {
		t.Errorf("ReadFile(main) = %q, %v", text, err)
	}
	if _, err := service.ReadFile("nope.go"); err == nil {
		t.Error("ReadFile of an unknown file must fail")
	}
}
