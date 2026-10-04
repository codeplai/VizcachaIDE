package errors

import (
	"regexp"
	"strings"

	"github.com/codeplai/VizcachaIDE/wails/internal/domain"
	"github.com/codeplai/VizcachaIDE/wails/internal/protocol/errorcatalog"
)

// gnuDiagnostic is the line both families print (and the checker too):
//
//	C:\mis programas\main.cpp:3:12: error: 'totl' was not declared in this scope; did you mean 'total'?
//
// The path may start with a drive letter and contain spaces, accents and any character but ':'.
var gnuDiagnostic = regexp.MustCompile(`^(?P<path>(?:[A-Za-z]:)?[^:\r\n]+?):(?P<line>\d+):(?P<column>\d+): (?P<kind>fatal error|error|warning|note): (?P<message>.*)$`)

// compilerLine reads a compiler diagnostic. Errors and warnings open a diagnostic; a note is
// appended to the raw text of the one before it (it explains it: "declared here", "to match this '{'").
func (p *parser) compilerLine(lines []string) (int, bool) {
	line := strings.TrimRight(lines[0], " \t\r")
	groups := errorcatalog.NamedGroups(gnuDiagnostic, line)
	if groups == nil {
		return 0, false
	}
	if groups["kind"] == "note" {
		p.noteLine(line)
		return 0, true
	}
	location := errorcatalog.LocationFrom(groups, p.workingDir)
	severity := domain.SeverityError
	if groups["kind"] == "warning" {
		severity = domain.SeverityWarning
	}
	p.add(domain.Diagnostic{
		Location: &location,
		Severity: severity,
		Message:  strings.TrimSpace(groups["message"]),
		Source:   sourceCompiler,
	}, []string{line}, true)
	return 0, true
}

// noteLine attaches a note to the last compiler diagnostic; a note without one is dropped.
func (p *parser) noteLine(line string) {
	last := len(p.diagnostics) - 1
	if last < 0 || p.diagnostics[last].Source != sourceCompiler {
		p.context = -1
		return
	}
	p.appendRaw(last, line)
	p.context = last
}
