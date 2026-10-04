package errors

import (
	"regexp"
	"strings"

	"github.com/codeplai/VizcachaIDE/wails/internal/domain"
	"github.com/codeplai/VizcachaIDE/wails/internal/protocol/errorcatalog"
)

var (
	// textHeader is the first line of a human diagnostic: "error[E0382]: msg", "warning: msg".
	textHeader = regexp.MustCompile(`^(?P<level>error|warning)(?:\[(?P<code>[A-Za-z0-9_:]+)\])?: (?P<message>.+)$`)
	// textArrow is the line that follows the header: " --> src\main.rs:4:23".
	textArrow = regexp.MustCompile(`^\s*--> (?P<path>.+):(?P<line>\d+):(?P<column>\d+)$`)
	// lintNote names the lint of a warning that has no code in the header.
	lintNote = regexp.MustCompile(`#\[(?:warn|deny|forbid)\((?P<lint>[A-Za-z0-9_:]+)\)\]`)
)

// textDiagnostic reads a human diagnostic: the header, its ` --> file:line:col` line and the
// continuation up to the blank line. It gives nil for the summaries and for lines that look
// like a header but are not rustc's (a program that prints "error: ..." itself).
func (r *reader) textDiagnostic(header map[string]string) *domain.Diagnostic {
	block := r.textBlock()
	severity, _ := severityOf(header["level"])
	message := header["message"]
	if summaryMessage.MatchString(message) {
		return nil
	}
	code := header["code"]
	if note := errorcatalog.NamedGroups(lintNote, strings.Join(block, "\n")); code == "" && note != nil {
		code = note["lint"]
	}
	location := r.arrowLocation(block)
	if code == "" && location == nil && r.identify(message) == "" {
		return nil
	}
	return &domain.Diagnostic{
		Location: location, Severity: severity, Message: message,
		RawText: strings.TrimRight(strings.Join(block, "\n"), "\n"),
		Source:  sourceOf(code), Code: r.codeOrID(code, message),
	}
}

// textBlock consumes lines from the header to the blank line that ends the diagnostic (cargo's
// "Caused by:" paragraphs are part of it).
func (r *reader) textBlock() []string {
	start := r.pos
	r.pos++
	for r.pos < len(r.lines) && !r.endsBlock(r.pos) {
		r.pos++
	}
	return r.lines[start:r.pos]
}

func (r *reader) endsBlock(index int) bool {
	line := r.lines[index]
	if strings.TrimSpace(line) == "" {
		next := index + 1
		return next >= len(r.lines) || !strings.HasPrefix(r.lines[next], "Caused by:")
	}
	return jsonLine.MatchString(line) || textHeader.MatchString(line) || panicHeaderGroups(line) != nil
}

func (r *reader) arrowLocation(block []string) *domain.SourceLocation {
	for _, line := range block[1:] {
		if groups := errorcatalog.NamedGroups(textArrow, line); groups != nil {
			location := errorcatalog.LocationFrom(groups, r.workingDir)
			return &location
		}
	}
	return nil
}
