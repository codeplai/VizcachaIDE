package errors

import (
	"regexp"

	"github.com/codeplai/VizcachaIDE/wails/internal/domain"
)

// runtimeFailure matches the lines of the failures that are not a panic: the program's own
// stack overflow message and fatal runtime errors, and the lines the runner writes when the
// program ends by a Windows exception code ("Segmentation fault", "Aborted"...).
var runtimeFailure = regexp.MustCompile(`^(?:thread '[^']*'(?: \(\d+\))? has overflowed its stack|fatal runtime error: .+|Segmentation fault|Aborted|Stack overflow|Floating point exception)$`)

// failureLine turns one of those lines into a diagnostic without location (the debugger gives
// it), or nil when the line is not one or the catalog does not know it.
func (r *reader) failureLine(line string) *domain.Diagnostic {
	if !runtimeFailure.MatchString(line) {
		return nil
	}
	code := r.identify(line)
	if code == "" {
		return nil
	}
	return &domain.Diagnostic{Severity: domain.SeverityError, Message: line, RawText: line, Source: sourceRuntime, Code: code}
}

// mainError reads the "Error: <debug>" that main prints when it returns Err: the last line of
// stderr, and only when nothing else was found (any program can print a line that starts so).
func (r *reader) mainError() []domain.Diagnostic {
	for index := len(r.lines) - 1; index >= 0; index-- {
		line := r.lines[index]
		if line == "" {
			continue
		}
		if !mainErrorLine.MatchString(line) || r.identify(line) == "" {
			return nil
		}
		return []domain.Diagnostic{{
			Severity: domain.SeverityError, Message: line, RawText: line,
			Source: sourceRuntime, Code: r.identify(line),
		}}
	}
	return nil
}
