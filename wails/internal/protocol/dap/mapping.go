package dap

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
	argumentsScope = "Arguments"
)

var currentThreadMark = regexp.MustCompile(`^\*\s*`)

func truncateValue(value string) string {
	if utf8.RuneCountInString(value) <= maxValueLength {
		return value
	}
	runes := []rune(value)
	return string(runes[:maxValueLength-1]) + ellipsis
}

// mapVariables converts variables, dropping those the flavor hides.
func mapVariables(raw []dap.Variable, flavor Flavor) []domain.Variable {
	variables := make([]domain.Variable, 0, len(raw))
	for _, item := range raw {
		if !flavor.KeepVariable(item) {
			continue
		}
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

// mapFrames converts a stack. With skipHidden the leading frames the flavor does not keep (the
// machinery of a panic) are dropped so the user sees their own code first.
func mapFrames(raw []dap.StackFrame, flavor Flavor, skipHidden bool) []domain.StackFrame {
	if skipHidden {
		raw = withoutLeadingHidden(raw, flavor)
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

func withoutLeadingHidden(raw []dap.StackFrame, flavor Flavor) []dap.StackFrame {
	for index, item := range raw {
		if flavor.KeepFrame(item) {
			return raw[index:]
		}
	}
	return raw
}

// localsReference is the reference of the Locals scope, or 0 when there is none.
func localsReference(scopes []dap.Scope, flavor Flavor) int {
	for _, scope := range scopes {
		if flavor.IsLocalsScope(scope.Name) {
			return scope.VariablesReference
		}
	}
	return 0
}

// scopeReferences returns the references of the Arguments and Locals scopes (0 when a scope
// is missing).
func scopeReferences(scopes []dap.Scope, flavor Flavor) (arguments, locals int) {
	for _, scope := range scopes {
		switch {
		case strings.HasPrefix(scope.Name, argumentsScope):
			arguments = scope.VariablesReference
		case flavor.IsLocalsScope(scope.Name):
			locals = scope.VariablesReference
		}
	}
	return arguments, locals
}

func mapThreads(threads []dap.Thread) []domain.Thread {
	mapped := make([]domain.Thread, 0, len(threads))
	for _, thread := range threads {
		name := currentThreadMark.ReplaceAllString(thread.Name, "")
		mapped = append(mapped, domain.Thread{ThreadID: thread.Id, Name: name})
	}
	return mapped
}
