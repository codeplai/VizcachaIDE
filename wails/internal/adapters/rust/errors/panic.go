package errors

import (
	"regexp"
	"strings"

	"github.com/codeplai/VizcachaIDE/wails/internal/domain"
	"github.com/codeplai/VizcachaIDE/wails/internal/protocol/errorcatalog"
)

// panicExplicit is the id of a panic the catalog has no specific entry for.
const panicExplicit = "RS-PANIC-EXPLICIT"

var (
	// panicHeader is the current format (the id of the thread is in parentheses since 1.9x; before
	// it was absent): the message is on the next line.
	panicHeader = regexp.MustCompile(`^thread '(?P<thread>[^']*)'(?: \((?P<id>\d+)\))? panicked at (?P<path>.+):(?P<line>\d+):(?P<column>\d+):$`)
	// oldPanicHeader is the format before 1.73: everything on one line.
	oldPanicHeader = regexp.MustCompile(`^thread '(?P<thread>[^']*)' panicked at '(?P<message>.*)', (?P<path>.+):(?P<line>\d+):(?P<column>\d+)$`)
	backtraceAt    = regexp.MustCompile(`^\s+at (?P<path>.+):(?P<line>\d+):(?P<column>\d+)$`)
	absolutePath   = regexp.MustCompile(`^(?:[A-Za-z]:[\\/]|/|\\)`)
	// standardLibrary matches the paths of code that is not the student's: the standard library
	// (/rustc/<hash>/library, the sysroot's rust-src) and the registry of cargo.
	standardLibrary = regexp.MustCompile(`/rustc/|lib/rustlib/src/|/\.cargo/registry/|(?:^|/)library/(?:core|std|alloc|proc_macro|panic_unwind|unwind)/`)
)

func panicHeaderGroups(line string) map[string]string {
	if groups := errorcatalog.NamedGroups(panicHeader, line); groups != nil {
		return groups
	}
	return errorcatalog.NamedGroups(oldPanicHeader, line)
}

// panicDiagnostic reads a panic of the program: the header, the message and the backtrace and
// notes that follow. The location is the student's line: the one of the header, or, when the
// panic happened inside the standard library (an unwrap in a function, an index in a Vec), the
// first frame of the backtrace that is in the student's code.
func (r *reader) panicDiagnostic(header map[string]string) *domain.Diagnostic {
	start := r.pos
	r.pos++
	message, hasMessage := header["message"]
	if !hasMessage && r.pos < len(r.lines) {
		message = r.lines[r.pos]
		r.pos++
	}
	for r.pos < len(r.lines) && continuesPanic(r.lines[r.pos]) {
		r.pos++
	}
	block := r.lines[start:r.pos]
	location := r.panicLocation(header, block)
	code := r.identify(message)
	if !strings.HasPrefix(code, "RS-PANIC-") { // a panic message can look like a compiler one
		code = panicExplicit
	}
	return &domain.Diagnostic{
		Location: location, Severity: domain.SeverityError, Message: message,
		RawText: strings.Join(block[:min(len(block), maxRawLines)], "\n"),
		Source:  sourceRuntime, Code: code,
	}
}

// continuesPanic recognises the lines after the message: extra message lines and frames (both
// indented), "stack backtrace:" and the notes of the runtime.
func continuesPanic(line string) bool {
	return strings.HasPrefix(line, " ") || strings.HasPrefix(line, "\t") ||
		line == "stack backtrace:" || strings.HasPrefix(line, "note: ")
}

func (r *reader) panicLocation(header map[string]string, block []string) *domain.SourceLocation {
	if !isStandardLibrary(header["path"]) {
		location := errorcatalog.LocationFrom(header, r.workingDir)
		return &location
	}
	for _, line := range block {
		frame := errorcatalog.NamedGroups(backtraceAt, line)
		if frame != nil && r.isStudentPath(frame["path"]) {
			location := errorcatalog.LocationFrom(frame, r.workingDir)
			return &location
		}
	}
	return nil
}

func isStandardLibrary(path string) bool {
	return standardLibrary.MatchString(strings.ReplaceAll(path, `\`, "/"))
}

// isStudentPath is a path of the backtrace that is not the library's and is relative (rustc
// prints ".\main.rs") or inside the working directory.
func (r *reader) isStudentPath(path string) bool {
	if isStandardLibrary(path) {
		return false
	}
	if !absolutePath.MatchString(path) {
		return true
	}
	root := strings.ToLower(strings.ReplaceAll(r.workingDir, `\`, "/"))
	return root != "" && strings.HasPrefix(strings.ToLower(strings.ReplaceAll(path, `\`, "/")), strings.TrimSuffix(root, "/")+"/")
}
