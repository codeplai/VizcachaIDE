package delve

import (
	"path/filepath"
	"regexp"
	"strings"
	"unicode/utf8"

	"github.com/codeplai/VizcachaIDE/wails/internal/domain"
	"github.com/google/go-dap"
)

const (
	maxValueLength = 200
	ellipsis       = "…"
	localsScope    = "Locals"
	subtleHint     = "subtle" // Delve marks runtime frames (the machinery of a panic) as subtle
	noiseOutput    = "Type 'dlv help' for list of commands."
)

var currentThreadMark = regexp.MustCompile(`^\*\s*`)

var stopReasons = map[string]domain.StopReason{
	"breakpoint":             domain.StopBreakpoint,
	"function breakpoint":    domain.StopBreakpoint,
	"data breakpoint":        domain.StopBreakpoint,
	"instruction breakpoint": domain.StopBreakpoint,
	"step":                   domain.StopStep,
	"goto":                   domain.StopStep,
	"entry":                  domain.StopEntry,
	"exception":              domain.StopException,
	"panic":                  domain.StopException,
	"pause":                  domain.StopPause,
}

// stopReason maps a DAP stop reason; anything unknown is a plain pause.
func stopReason(reason string) domain.StopReason {
	if known, found := stopReasons[reason]; found {
		return known
	}
	return domain.StopPause
}

// stopDescription joins Delve's description and text: "panic: runtime error: ...".
func stopDescription(event *dap.StoppedEvent) string {
	parts := make([]string, 0, 2)
	for _, part := range []string{event.Body.Description, event.Body.Text} {
		if part != "" {
			parts = append(parts, strings.Trim(part, `"`))
		}
	}
	return strings.Join(parts, ": ")
}

func truncateValue(value string) string {
	if utf8.RuneCountInString(value) <= maxValueLength {
		return value
	}
	runes := []rune(value)
	return string(runes[:maxValueLength-1]) + ellipsis
}

func mapVariables(raw []dap.Variable) []domain.Variable {
	variables := make([]domain.Variable, 0, len(raw))
	for _, item := range raw {
		variables = append(variables, domain.Variable{
			Name:      item.Name,
			TypeName:  item.Type,
			Value:     truncateValue(item.Value),
			Reference: item.VariablesReference,
			Children:  []domain.Variable{},
		})
	}
	return variables
}

func mapLocation(source *dap.Source, line, column int) *domain.SourceLocation {
	if source == nil || source.Path == "" || line <= 0 {
		return nil
	}
	return &domain.SourceLocation{File: filepath.Clean(source.Path), Line: line, Column: max(column, 1)}
}

// mapFrames converts a stack. With skipRuntime the leading runtime frames (the
// machinery of a panic) are dropped so the user sees their own code first.
func mapFrames(raw []dap.StackFrame, skipRuntime bool) []domain.StackFrame {
	if skipRuntime {
		raw = withoutLeadingRuntime(raw)
	}
	frames := make([]domain.StackFrame, 0, len(raw))
	for _, item := range raw {
		frames = append(frames, domain.StackFrame{
			FrameID:  item.Id,
			Function: item.Name,
			Location: mapLocation(item.Source, item.Line, item.Column),
		})
	}
	return frames
}

func withoutLeadingRuntime(raw []dap.StackFrame) []dap.StackFrame {
	for index, item := range raw {
		if item.PresentationHint != subtleHint {
			return raw[index:]
		}
	}
	return raw
}

// localsReference is the reference of the Locals scope (also "Locals (warning:
// optimized function)"), or 0 when there is none.
func localsReference(scopes []dap.Scope) int {
	for _, scope := range scopes {
		if strings.HasPrefix(scope.Name, localsScope) {
			return scope.VariablesReference
		}
	}
	return 0
}

func mapGoroutines(threads []dap.Thread) []domain.Thread {
	goroutines := make([]domain.Thread, 0, len(threads))
	for _, thread := range threads {
		name := currentThreadMark.ReplaceAllString(thread.Name, "")
		goroutines = append(goroutines, domain.Thread{ThreadID: thread.Id, Name: name})
	}
	return goroutines
}

// outputText returns the text and category ("stdout", "stderr", "console") of an
// output event, or false when it is not meant for the user.
func outputText(event *dap.OutputEvent) (string, string, bool) {
	text, category := event.Body.Output, event.Body.Category
	if category == "telemetry" || text == "" || strings.TrimSpace(text) == noiseOutput {
		return "", "", false
	}
	if category == "stdout" || category == "stderr" {
		return text, category, true
	}
	return text, "console", true
}
