package errors

import (
	"regexp"
	"strings"

	"github.com/codeplai/VizcachaIDE/wails/internal/domain"
	"github.com/codeplai/VizcachaIDE/wails/internal/protocol/errorcatalog"
)

// Recognised shapes of the output of go build / go run / go vet:
//
//	# command-line-arguments                       (package header, ignored)
//	# [command-line-arguments]                     (go vet header: what follows is vet)
//	./main.go:12:5: declared and not used: x       (relative, absolute, Windows or POSIX)
//	\thave (number)                                (continuation of the previous message)
//	panic: runtime error: index out of range [5] with length 3
//	goroutine 1 [running]:
//	main.main()
//	\t/home/ana/hello/main.go:9 +0x1d               (first user frame = location)
var (
	compilerLine = regexp.MustCompile(`^(?P<path>(?:[A-Za-z]:)?[^:\r\n]+?\.go):(?P<line>\d+)(?::(?P<column>\d+))?:\s*(?P<message>.*)$`)
	frameLine    = regexp.MustCompile(`^\t(?P<path>.+?\.go):(?P<line>\d+)(?: \+0x[0-9a-fA-F]+)?\s*$`)
	panicStart   = regexp.MustCompile(`^(?:panic: |fatal error: )`)
	vetHeader    = regexp.MustCompile(`^# \[.*\]\s*$`)
	noise        = regexp.MustCompile(`^(?:# .*|exit status \d+|too many errors)\s*$`)
)

// Diagnostic sources.
const (
	sourceCompiler = "compiler"
	sourceVet      = "vet"
	sourcePanic    = "panic"
)

var runtimePrefixes = []string{"runtime.", "runtime/", "panic(", "internal/"}

// parser turns raw Go tool output into diagnostics. identify returns the catalog id of a message.
type parser struct {
	identify func(message string) string
}

// ParseOutput is the errorcatalog.OutputParser of Go: compiler errors, vet warnings and panics.
func ParseOutput(rawOutput, workingDir string, identify func(message string) string) []domain.Diagnostic {
	return parser{identify: identify}.parse(rawOutput, workingDir)
}

func (p parser) parse(rawOutput, workingDir string) []domain.Diagnostic {
	lines := errorcatalog.SplitLines(rawOutput)
	var diagnostics []domain.Diagnostic
	source := sourceCompiler
	for index, line := range lines {
		if panicStart.MatchString(line) {
			return append(diagnostics, p.panicDiagnostic(lines[index:], workingDir))
		}
		if vetHeader.MatchString(line) {
			source = sourceVet
			continue
		}
		if strings.HasPrefix(line, "\t") && len(diagnostics) > 0 {
			diagnostics[len(diagnostics)-1].RawText += "\n" + line
			continue
		}
		if diagnostic, ok := p.line(line, workingDir, source); ok {
			diagnostics = append(diagnostics, diagnostic)
		}
	}
	return diagnostics
}

func (p parser) line(line, workingDir, source string) (domain.Diagnostic, bool) {
	if groups := errorcatalog.NamedGroups(compilerLine, line); groups != nil {
		location := errorcatalog.LocationFrom(groups, workingDir)
		return p.diagnostic(&location, strings.TrimSpace(groups["message"]), line, source), true
	}
	text := strings.TrimSpace(line)
	if text == "" || noise.MatchString(text) {
		return domain.Diagnostic{}, false
	}
	if p.identify(text) != "" || strings.HasPrefix(text, "go: ") {
		return p.diagnostic(nil, text, line, source), true
	}
	return domain.Diagnostic{}, false
}

func (p parser) diagnostic(location *domain.SourceLocation, message, rawText, source string) domain.Diagnostic {
	code := p.identify(message)
	if strings.HasPrefix(code, "V-") {
		source = sourceVet
	}
	severity := domain.SeverityError
	if source == sourceVet {
		severity = domain.SeverityWarning
	}
	return domain.Diagnostic{
		Location: location,
		Severity: severity,
		Message:  message,
		RawText:  strings.TrimRight(rawText, " \t\r\n"),
		Source:   source,
		Code:     code,
	}
}

func (p parser) panicDiagnostic(lines []string, workingDir string) domain.Diagnostic {
	var block []string
	for _, line := range lines {
		if !noise.MatchString(line) {
			block = append(block, line)
		}
	}
	message := strings.TrimSpace(block[0])
	return domain.Diagnostic{
		Location: firstUserFrame(block, workingDir),
		Severity: domain.SeverityError,
		Message:  message,
		RawText:  strings.TrimSpace(strings.Join(block, "\n")),
		Source:   sourcePanic,
		Code:     p.identify(message),
	}
}

func firstUserFrame(block []string, workingDir string) *domain.SourceLocation {
	for index := 1; index < len(block); index++ {
		groups := errorcatalog.NamedGroups(frameLine, block[index])
		if groups == nil || isRuntimeFrame(block[index-1], groups["path"]) {
			continue
		}
		location := errorcatalog.LocationFrom(groups, workingDir)
		return &location
	}
	return nil
}

func isRuntimeFrame(functionLine, path string) bool {
	for _, prefix := range runtimePrefixes {
		if strings.HasPrefix(functionLine, prefix) {
			return true
		}
	}
	return strings.Contains(strings.ReplaceAll(path, "\\", "/"), "/src/runtime/")
}
