package delve

import (
	"encoding/json"
	"strings"

	"github.com/codeplai/VizcachaIDE/wails/internal/domain"
	protodap "github.com/codeplai/VizcachaIDE/wails/internal/protocol/dap"
	"github.com/google/go-dap"
)

const (
	subtleHint  = "subtle" // Delve marks runtime frames (the machinery of a panic) as subtle
	noiseOutput = "Type 'dlv help' for list of commands."
	localsScope = "Locals"
	panicReason = "panic"
)

// flavor holds what is particular to "dlv dap".
type flavor struct {
	debugBinary string // where Delve writes the debug build
}

var _ protodap.Flavor = flavor{}

func (flavor) AdapterID() string { return "go" }

func (f flavor) Launch(config domain.RunConfiguration, env map[string]string) (json.RawMessage, error) {
	return launchRaw(config, env, f.debugBinary)
}

func (flavor) ExceptionFilters() []string { return nil }

// StopReason treats Delve's "panic" as an exception and keeps its "panic: ..." description.
func (flavor) StopReason(event *dap.StoppedEvent) (domain.StopReason, string) {
	if event.Body.Reason == panicReason {
		return domain.StopException, protodap.StopDescription(event)
	}
	return protodap.StandardStopReason(event.Body.Reason), protodap.StopDescription(event)
}

func (flavor) KeepFrame(frame dap.StackFrame) bool { return frame.PresentationHint != subtleHint }

func (flavor) KeepVariable(dap.Variable) bool { return true }

// IsLocalsScope also matches "Locals (warning: optimized function)".
func (flavor) IsLocalsScope(name string) bool { return strings.HasPrefix(name, localsScope) }

func (flavor) Output(event *dap.OutputEvent) (string, string, bool) {
	if strings.TrimSpace(event.Body.Output) == noiseOutput {
		return "", "", false
	}
	return protodap.StandardOutput(event)
}
