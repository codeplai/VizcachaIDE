package errors

import (
	"regexp"
	"strings"

	"github.com/codeplai/VizcachaIDE/wails/internal/domain"
	"github.com/codeplai/VizcachaIDE/wails/internal/protocol/errorcatalog"
)

// Shapes of the linker output:
//
//	GNU ld:  C:/src/main.cpp:3:(.text+0x13): undefined reference to `calcular(int)'
//	lld:     ld.lld: error: undefined symbol: calcular(int)
//	         >>> referenced by C:\src\main.cpp:3
//
// A symbol defined twice (two files with main) is "multiple definition of `main'" in GNU ld and
// "duplicate symbol: main" with ">>> defined at C:\src\main.cpp:3" lines in lld.
var (
	undefinedReference = regexp.MustCompile("undefined reference to `.+'$")
	gnuLinkLocation    = regexp.MustCompile(`^(?P<path>.+?):(?P<line>\d+):\(`)
	undefinedSymbol    = regexp.MustCompile(`^(?:\S*ld\.lld|lld): error: (?P<message>undefined symbol: .+)$`)
	referencedBy       = regexp.MustCompile(`^>>>\s+(?:referenced by|defined at) (?P<path>.+):(?P<line>\d+)\s*$`)
	duplicateSymbol    = regexp.MustCompile(`^(?:\S*ld\.lld|lld): error: (?P<message>duplicate symbol: .+)$`)
	multipleDefinition = regexp.MustCompile("multiple definition of `[^']+'")
)

// undefinedReferenceLine reads an undefined reference of GNU ld, with the file and line of the
// call when the object was built with -g.
func (p *parser) undefinedReferenceLine(lines []string) (int, bool) {
	line := strings.TrimRight(lines[0], " \t\r")
	span := undefinedReference.FindStringIndex(line)
	if span == nil {
		return 0, false
	}
	var location *domain.SourceLocation
	if groups := errorcatalog.NamedGroups(gnuLinkLocation, line); groups != nil {
		location = p.ownLocation(groups)
	}
	p.add(domain.Diagnostic{
		Location: location,
		Severity: domain.SeverityError,
		Message:  line[span[0]:],
		Source:   sourceLinker,
	}, []string{line}, false)
	return 0, true
}

// undefinedSymbolLine reads an undefined symbol of lld. Its ">>> referenced by" lines follow as
// context and give the location.
func (p *parser) undefinedSymbolLine(lines []string) (int, bool) {
	line := strings.TrimRight(lines[0], " \t\r")
	groups := errorcatalog.NamedGroups(undefinedSymbol, line)
	if groups == nil {
		return 0, false
	}
	p.add(domain.Diagnostic{
		Severity: domain.SeverityError,
		Message:  groups["message"],
		Source:   sourceLinker,
	}, []string{line}, true)
	return 0, true
}

// duplicateSymbolLine reads a symbol defined twice by lld; its ">>> defined at" lines follow as
// context and give the location of the first definition.
func (p *parser) duplicateSymbolLine(lines []string) (int, bool) {
	line := strings.TrimRight(lines[0], " \t\r")
	groups := errorcatalog.NamedGroups(duplicateSymbol, line)
	if groups == nil {
		return 0, false
	}
	p.add(domain.Diagnostic{Severity: domain.SeverityError, Message: groups["message"], Source: sourceLinker}, []string{line}, true)
	return 0, true
}

// multipleDefinitionLine reads a symbol defined twice by GNU ld, at the file and line of the
// second definition when the object was built with -g.
func (p *parser) multipleDefinitionLine(lines []string) (int, bool) {
	line := strings.TrimRight(lines[0], " \t\r")
	span := multipleDefinition.FindStringIndex(line)
	if span == nil {
		return 0, false
	}
	var location *domain.SourceLocation
	if groups := errorcatalog.NamedGroups(gnuLinkLocation, line); groups != nil {
		location = p.ownLocation(groups)
	}
	p.add(domain.Diagnostic{Location: location, Severity: domain.SeverityError, Message: line[span[0]:span[1]], Source: sourceLinker}, []string{line}, false)
	return 0, true
}

// referencedBy returns the location of a ">>> referenced by file:line" line of lld.
func (p *parser) referencedBy(line string) *domain.SourceLocation {
	groups := errorcatalog.NamedGroups(referencedBy, line)
	if groups == nil {
		return nil
	}
	return p.ownLocation(groups)
}

// ownLocation is the location of a linker reference when it is in the student's folder; the
// C runtime (crtexewin.c) and the libraries are not code the student can open.
func (p *parser) ownLocation(groups map[string]string) *domain.SourceLocation {
	location := errorcatalog.LocationFrom(groups, p.workingDir)
	if p.workingDir != "" && !inside(location.File, p.workingDir) {
		return nil
	}
	return &location
}

// inside reports whether path is inside dir (separators and case are not compared strictly).
func inside(path, dir string) bool {
	normal := func(text string) string {
		return strings.ToLower(strings.TrimRight(strings.ReplaceAll(text, `\`, "/"), "/"))
	}
	return strings.HasPrefix(normal(path), normal(dir)+"/")
}
