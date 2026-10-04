package errors

import (
	"regexp"
	"strings"

	"github.com/codeplai/VizcachaIDE/wails/internal/domain"
	"github.com/codeplai/VizcachaIDE/wails/internal/protocol/errorcatalog"
)

// Diagnostic sources.
const (
	sourceCompiler = "compiler"
	sourceClippy   = "clippy"
	sourceRuntime  = "runtime"
)

// maxRawLines caps the context a runtime diagnostic keeps (a backtrace can be long).
const maxRawLines = 15

var (
	jsonLine = regexp.MustCompile(`^\s*\{`)
	// summaryMessage matches what rustc and cargo print at the end and that explains nothing:
	// "aborting due to 1 previous error", "4 warnings emitted", "could not compile `x`...".
	summaryMessage = regexp.MustCompile("^(?:aborting due to |\\d+ (?:warnings?|errors?) emitted|could not compile |`[^`]+` \\(.*\\) generated \\d+ warnings?)")
	mainErrorLine  = regexp.MustCompile(`^Error: .+$`)
)

// reader walks the lines of the output; each source (JSON, text, panic, crash) consumes its own.
type reader struct {
	lines      []string
	pos        int
	workingDir string
	identify   func(message string) string
}

// ParseOutput is the errorcatalog.OutputParser of Rust. The raw output is what rustc, cargo or
// clippy printed (JSON lines or the human text) or what the program wrote on stderr.
func ParseOutput(rawOutput, workingDir string, identify func(message string) string) []domain.Diagnostic {
	r := &reader{lines: errorcatalog.SplitLines(rawOutput), workingDir: workingDir, identify: identify}
	var found []domain.Diagnostic
	for r.pos < len(r.lines) {
		found = appendUnique(found, r.readNext())
	}
	if len(found) == 0 {
		return r.mainError()
	}
	return found
}

// readNext consumes the diagnostic that starts at the current line (and its continuation) or
// skips the line.
func (r *reader) readNext() *domain.Diagnostic {
	line := r.lines[r.pos]
	if jsonLine.MatchString(line) {
		r.pos++
		return r.jsonDiagnostic(line)
	}
	if header := errorcatalog.NamedGroups(textHeader, line); header != nil {
		return r.textDiagnostic(header)
	}
	if header := panicHeaderGroups(line); header != nil {
		return r.panicDiagnostic(header)
	}
	r.pos++
	return r.failureLine(line)
}

// appendUnique adds a diagnostic; the same runtime failure reported twice (the program's
// "has overflowed its stack" plus the runner's "Stack overflow") stays one.
func appendUnique(found []domain.Diagnostic, next *domain.Diagnostic) []domain.Diagnostic {
	if next == nil {
		return found
	}
	if next.Source == sourceRuntime && next.Location == nil {
		for _, old := range found {
			if old.Source == sourceRuntime && old.Location == nil && old.Code == next.Code {
				return found
			}
		}
	}
	return append(found, *next)
}

// codeOrID is the code rustc gave, or else the catalog id of the message ("" when unknown).
func (r *reader) codeOrID(code, message string) string {
	if code != "" {
		return code
	}
	return r.identify(message)
}

func sourceOf(code string) string {
	if strings.HasPrefix(code, "clippy::") {
		return sourceClippy
	}
	return sourceCompiler
}

func severityOf(level string) (domain.Severity, bool) {
	switch level {
	case "error", "error: internal compiler error":
		return domain.SeverityError, true
	case "warning":
		return domain.SeverityWarning, true
	}
	return "", false // note, help, failure-note: folded into the diagnostic they belong to
}
