package dap

import (
	"encoding/json"
	"strings"

	"github.com/codeplai/VizcachaIDE/wails/internal/domain"
	"github.com/google/go-dap"
)

// testFlavor behaves like the Delve flavor so the recorded Delve transcripts keep working.
type testFlavor struct{}

func (testFlavor) AdapterID() string { return "go" }

func (testFlavor) Launch(domain.RunConfiguration, map[string]string) (json.RawMessage, error) {
	return json.RawMessage(`{}`), nil
}

func (testFlavor) ExceptionFilters() []string { return nil }

func (testFlavor) StopReason(event *dap.StoppedEvent) (domain.StopReason, string) {
	return StandardStopReason(event.Body.Reason), StopDescription(event)
}

func (testFlavor) KeepFrame(frame dap.StackFrame) bool { return frame.PresentationHint != "subtle" }

func (testFlavor) KeepVariable(dap.Variable) bool { return true }

func (testFlavor) IsLocalsScope(name string) bool { return strings.HasPrefix(name, "Locals") }

func (testFlavor) Output(event *dap.OutputEvent) (string, string, bool) {
	return StandardOutput(event)
}
