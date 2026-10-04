package debugpy

import (
	"encoding/json"
	"strings"

	"github.com/codeplai/VizcachaIDE/wails/internal/domain"
	protodap "github.com/codeplai/VizcachaIDE/wails/internal/protocol/dap"
	"github.com/google/go-dap"
)

// hiddenGroups are the synthetic variable groups debugpy adds to every scope.
var hiddenGroups = map[string]bool{
	"special variables":  true,
	"function variables": true,
	"class variables":    true,
}

// flavor holds what is particular to debugpy.
type flavor struct {
	interpreter string
	console     string // "integratedTerminal" (PTY) or "internalConsole"
}

var _ protodap.Flavor = flavor{}

func (flavor) AdapterID() string { return "python" }

func (f flavor) Launch(config domain.RunConfiguration, env map[string]string) (json.RawMessage, error) {
	return launchRaw(config, env, f.interpreter, f.console)
}

// ExceptionFilters stops on exceptions nobody catches.
func (flavor) ExceptionFilters() []string { return []string{"uncaught"} }

// genericException is what older debugpy versions put in Description of an exception stop.
const genericException = "Paused on exception"

// StopReason shows "NameError: name 'x' is not defined" for an exception: debugpy 1.8 puts the
// class in Text and the message in Description (older ones a generic "Paused on exception").
func (flavor) StopReason(event *dap.StoppedEvent) (domain.StopReason, string) {
	reason := protodap.StandardStopReason(event.Body.Reason)
	if reason != domain.StopException {
		return reason, protodap.StopDescription(event)
	}
	parts := make([]string, 0, 2)
	for _, part := range []string{event.Body.Text, event.Body.Description} {
		if part != "" && part != genericException {
			parts = append(parts, strings.Trim(part, `"`))
		}
	}
	return reason, strings.Join(parts, ": ")
}

func (flavor) KeepFrame(dap.StackFrame) bool { return true }

// KeepVariable hides debugpy's synthetic groups and the __dunder__ names.
func (flavor) KeepVariable(variable dap.Variable) bool {
	if hiddenGroups[variable.Name] {
		return false
	}
	name := variable.Name
	dunder := len(name) > 4 && strings.HasPrefix(name, "__") && strings.HasSuffix(name, "__")
	return !dunder
}

func (flavor) IsLocalsScope(name string) bool { return name == "Locals" }

func (flavor) Output(event *dap.OutputEvent) (string, string, bool) {
	return protodap.StandardOutput(event)
}
