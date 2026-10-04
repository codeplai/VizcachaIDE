package dap

import (
	"strings"

	"github.com/codeplai/VizcachaIDE/wails/internal/domain"
	"github.com/google/go-dap"
)

var standardStopReasons = map[string]domain.StopReason{
	"breakpoint":             domain.StopBreakpoint,
	"function breakpoint":    domain.StopBreakpoint,
	"data breakpoint":        domain.StopBreakpoint,
	"instruction breakpoint": domain.StopBreakpoint,
	"step":                   domain.StopStep,
	"goto":                   domain.StopStep,
	"entry":                  domain.StopEntry,
	"exception":              domain.StopException,
	"pause":                  domain.StopPause,
}

// StandardStopReason maps the stop reasons of the DAP specification; anything unknown is a
// plain pause. Flavors call it for what they do not treat specially.
func StandardStopReason(reason string) domain.StopReason {
	if known, found := standardStopReasons[reason]; found {
		return known
	}
	return domain.StopPause
}

// StopDescription joins the description and text of a stop: "panic: runtime error: ...".
func StopDescription(event *dap.StoppedEvent) string {
	parts := make([]string, 0, 2)
	for _, part := range []string{event.Body.Description, event.Body.Text} {
		if part != "" {
			parts = append(parts, strings.Trim(part, `"`))
		}
	}
	return strings.Join(parts, ": ")
}

// StandardOutput returns the text and category ("stdout", "stderr", "console") of an output
// event, or false when it is not meant for the user (telemetry, empty).
func StandardOutput(event *dap.OutputEvent) (text, category string, ok bool) {
	text, category = event.Body.Output, event.Body.Category
	if category == "telemetry" || text == "" {
		return "", "", false
	}
	if category == "stdout" || category == "stderr" {
		return text, category, true
	}
	return text, "console", true
}
