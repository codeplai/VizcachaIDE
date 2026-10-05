package bridge

import (
	"testing"

	"github.com/codeplai/VizcachaIDE/wails/internal/app"
)

type fakeTerminalHost struct{ calls []string }

func (f *fakeTerminalHost) Start(dir string, cols, rows int) (string, error) {
	f.calls = append(f.calls, "start "+dir)
	return "t1", nil
}
func (f *fakeTerminalHost) Write(id, data string) error {
	f.calls = append(f.calls, "write "+id+" "+data)
	return nil
}
func (f *fakeTerminalHost) Resize(id string, cols, rows int) error {
	f.calls = append(f.calls, "resize "+id)
	return nil
}
func (f *fakeTerminalHost) Close(id string) error { f.calls = append(f.calls, "close "+id); return nil }
func (f *fakeTerminalHost) CloseAll()             {}

var _ app.TerminalHost = (*fakeTerminalHost)(nil)

func TestTerminalServiceForwardsToTheHost(t *testing.T) {
	host := &fakeTerminalHost{}
	service := NewTerminalService(host)
	id, err := service.Start("/work", 80, 24)
	if err != nil || id != "t1" {
		t.Fatalf("start: %q %v", id, err)
	}
	_ = service.Write(id, "ls\r")
	_ = service.Resize(id, 100, 30)
	_ = service.Close(id)
	want := []string{"start /work", "write t1 ls\r", "resize t1", "close t1"}
	for i, call := range want {
		if host.calls[i] != call {
			t.Errorf("call %d = %q, want %q", i, host.calls[i], call)
		}
	}
}
