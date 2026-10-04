package errors

import (
	"regexp"
	"strings"

	"github.com/codeplai/VizcachaIDE/wails/internal/domain"
	"github.com/codeplai/VizcachaIDE/wails/internal/protocol/errorcatalog"
)

// A syntax error has no traceback:
//
//	  File "C:\ana\hola.py", line 1
//	    for i in range(3)
//	                     ^
//	SyntaxError: expected ':'
var (
	syntaxLine = regexp.MustCompile(`^(?:SyntaxError|IndentationError|TabError)(?:: .*)?$`)
	caretLine  = regexp.MustCompile(`^(?P<pad>\s*)\^+\s*$`)
)

// codeIndent is the indentation Python gives the code line it prints under "File ...".
const codeIndent = 4

func (p parser) syntaxError(lines []string) (domain.Diagnostic, bool) {
	var location *domain.SourceLocation
	start := 0
	for index, line := range lines {
		if groups := errorcatalog.NamedGroups(frameLine, line); groups != nil {
			found := errorcatalog.LocationFrom(groups, p.workingDir)
			location, start = &found, index
			continue
		}
		if !syntaxLine.MatchString(line) {
			continue
		}
		if location != nil {
			location.Column = caretColumn(lines[start+1 : index])
		}
		raw := strings.Join(lines[start:index+1], "\n")
		return p.diagnostic(location, strings.TrimSpace(line), raw, sourceSyntax), true
	}
	return domain.Diagnostic{}, false
}

// caretColumn is the 1-based column of the first ^ under the code line, 1 when there is none.
// Python prints the code without its leading spaces, so for an indented line it is relative.
func caretColumn(between []string) int {
	for _, line := range between {
		if groups := errorcatalog.NamedGroups(caretLine, line); groups != nil {
			return max(len(groups["pad"])-codeIndent, 0) + 1
		}
	}
	return 1
}
