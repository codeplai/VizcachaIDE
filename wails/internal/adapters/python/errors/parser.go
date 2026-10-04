package errors

import (
	"regexp"
	"strings"

	"github.com/codeplai/VizcachaIDE/wails/internal/domain"
	"github.com/codeplai/VizcachaIDE/wails/internal/protocol/errorcatalog"
)

// Recognised shapes of the output of python:
//
//	Traceback (most recent call last):
//	  File "C:\ana\hola.py", line 7, in <module>     (frames, outermost first)
//	    main()                                       (code and ^^^^ markers, indented)
//	NameError: name 'x' is not defined               (last line = message)
var (
	tracebackStart = regexp.MustCompile(`Traceback \(most recent call last\):\s*$`)
	frameLine      = regexp.MustCompile(`^\s*File "(?P<path>.+)", line (?P<line>\d+)(?:, in (?P<function>.+))?\s*$`)
	cannotOpen     = regexp.MustCompile(`^.*: can't open file '.*': \[Errno \d+\] .*$`)
)

// Diagnostic sources.
const (
	sourceRuntime = "runtime"
	sourceSyntax  = "syntax"
	sourceRuff    = "ruff"
)

// parser turns raw Python output into diagnostics. identify returns the catalog id of a message.
type parser struct {
	identify   func(message string) string
	workingDir string
}

// ParseOutput is the errorcatalog.OutputParser of Python. The raw output is either what a
// program printed (tracebacks, syntax errors, "can't open file") or the JSON array of
// `ruff check --output-format json`.
func ParseOutput(rawOutput, workingDir string, identify func(message string) string) []domain.Diagnostic {
	p := parser{identify: identify, workingDir: workingDir}
	if strings.HasPrefix(strings.TrimSpace(rawOutput), "[") {
		return p.ruffDiagnostics(rawOutput)
	}
	lines := errorcatalog.SplitLines(rawOutput)
	if start := lastTraceback(lines); start >= 0 {
		return []domain.Diagnostic{p.traceback(lines[start:])}
	}
	if diagnostic, ok := p.syntaxError(lines); ok {
		return []domain.Diagnostic{diagnostic}
	}
	for _, line := range lines {
		if cannotOpen.MatchString(line) {
			return []domain.Diagnostic{p.diagnostic(nil, strings.TrimSpace(line), line, sourceRuntime)}
		}
	}
	return nil
}

// lastTraceback returns the index of the last "Traceback" header (a chained exception prints
// several; the last one is the error that stopped the program), or -1.
func lastTraceback(lines []string) int {
	for index := len(lines) - 1; index >= 0; index-- {
		if tracebackStart.MatchString(lines[index]) {
			return index
		}
	}
	return -1
}

func (p parser) traceback(lines []string) domain.Diagnostic {
	lines = append([]string{tracebackStart.FindString(lines[0])}, lines[1:]...) // drops "Nombre: " that a prompt left before it
	end := len(lines) - 1
	for index := 1; index < len(lines); index++ {
		if lines[index] != "" && !strings.HasPrefix(lines[index], " ") && !strings.HasPrefix(lines[index], "\t") {
			end = index
			break
		}
	}
	block := lines[:end+1]
	message := strings.TrimSpace(lines[end])
	return p.diagnostic(p.userFrame(block), message, strings.Join(block, "\n"), sourceRuntime)
}

// userFrame is the last frame of the student's code: inside the working directory, and never
// the standard library or installed packages. Without a working directory (or when the file
// lives elsewhere) it is the last frame that is not of the library.
func (p parser) userFrame(block []string) *domain.SourceLocation {
	var fallback *domain.SourceLocation
	for index := len(block) - 1; index >= 0; index-- {
		groups := errorcatalog.NamedGroups(frameLine, block[index])
		if groups == nil || strings.HasPrefix(groups["path"], "<") {
			continue
		}
		location := errorcatalog.LocationFrom(groups, p.workingDir)
		if p.workingDir != "" && inside(location.File, p.workingDir) {
			return &location
		}
		if fallback == nil && !isLibrary(groups["path"]) {
			fallback = &location
		}
	}
	return fallback
}

func (p parser) diagnostic(location *domain.SourceLocation, message, rawText, source string) domain.Diagnostic {
	return domain.Diagnostic{
		Location: location,
		Severity: domain.SeverityError,
		Message:  message,
		RawText:  strings.TrimRight(rawText, " \t\r\n"),
		Source:   source,
		Code:     p.identify(message),
	}
}

// libraryPath matches the folders of Python itself and of installed packages: lib/python3.x on
// POSIX, <python folder>/Lib on Windows, and site-packages anywhere.
var libraryPath = regexp.MustCompile(`/(?:site-packages|dist-packages)/|/lib/python3|/(?:cpython|python)[^/]*/lib/`)

// isLibrary reports whether a frame belongs to Python itself, to an installed package or to
// code that is not a file ("<frozen runpy>", "<string>", "<stdin>").
func isLibrary(path string) bool {
	normal := strings.ToLower(strings.ReplaceAll(path, `\`, "/"))
	return strings.HasPrefix(normal, "<") || libraryPath.MatchString(normal)
}

// inside reports whether path is inside dir (separators and case are not compared strictly).
func inside(path, dir string) bool {
	normal := func(text string) string {
		return strings.ToLower(strings.TrimRight(strings.ReplaceAll(text, `\`, "/"), "/"))
	}
	return strings.HasPrefix(normal(path), normal(dir)+"/")
}
