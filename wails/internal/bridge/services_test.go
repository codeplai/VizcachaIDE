package bridge

import (
	"github.com/codeplai/VizcachaIDE/wails/internal/domain"
	"testing"
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

func TestSettingsServiceNotifiesOnSave(t *testing.T) {
	sink := &recordingSink{}
	service := NewSettingsService(sink, NewMemorySettingsStore(), NewLanguageResolver(NewMemorySettingsStore(), nil), newTestRegistry(t))
	if err := service.Save(domain.DefaultSettings()); err != nil {
		t.Fatal(err)
	}
	if len(sink.calls) != 1 || sink.calls[0] != "settings" {
		t.Errorf("calls = %v, want [settings]", sink.calls)
	}
}
