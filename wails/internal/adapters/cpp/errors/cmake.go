package errors

import (
	"regexp"
	"strings"

	"github.com/codeplai/VizcachaIDE/wails/internal/domain"
	"github.com/codeplai/VizcachaIDE/wails/internal/protocol/errorcatalog"
)

// sourceCMake is the Source of the diagnostics CMake prints.
const sourceCMake = "cmake"

const cmakeSevere = "Error"

// genericCMakeError is the catalog entry of a CMake error nothing more specific recognises.
const genericCMakeError = "CPP-CMAKE-ERROR"

// Shapes of CMake's messages: the first line says where, and the message follows indented.
//
//	CMake Error at CMakeLists.txt:12 (find_package):
//	  Could not find a package configuration file provided by "fmt"
//	CMake Warning in C:/proyecto/build/CMakeFiles/x.dir:
//	  The object file directory ...
//	CMake Error: The source directory "C:/x" does not appear to contain CMakeLists.txt.
var (
	cmakeAt     = regexp.MustCompile(`^CMake (?P<kind>Error|Warning|Deprecation Warning)(?: \((?:dev|author)\))? at (?P<path>.+?):(?P<line>\d+)(?: \([^)]*\))?:$`)
	cmakeIn     = regexp.MustCompile(`^CMake (?P<kind>Error|Warning|Deprecation Warning)(?: \((?:dev|author)\))? in .+:$`)
	cmakeInline = regexp.MustCompile(`^CMake (?P<kind>Error|Warning): (?P<message>.+)$`)
)

// cmakeLine reads a CMake error or warning. The message is the first line after the header (the
// rest, indented, stays in the raw text); the file and line come from the header when it has them.
func (p *parser) cmakeLine(lines []string) (int, bool) {
	line := strings.TrimRight(lines[0], " \t\r")
	if groups := errorcatalog.NamedGroups(cmakeInline, line); groups != nil {
		p.addCMake(groups["kind"], groups["message"], nil, []string{line})
		return 0, true
	}
	groups := errorcatalog.NamedGroups(cmakeAt, line)
	var location *domain.SourceLocation
	if groups != nil {
		found := errorcatalog.LocationFrom(groups, p.workingDir)
		location = &found
	} else if groups = errorcatalog.NamedGroups(cmakeIn, line); groups == nil {
		return 0, false
	}
	message, extra := firstText(lines[1:])
	p.addCMake(groups["kind"], message, location, append([]string{line}, lines[1:1+extra]...))
	return extra, true
}

// firstText is the first non-blank line of the message that follows a header, and how many lines
// were consumed to reach it.
func firstText(lines []string) (text string, consumed int) {
	for index, line := range lines {
		trimmed := strings.TrimSpace(line)
		if trimmed == "" {
			continue
		}
		if !strings.HasPrefix(line, " ") {
			break // not the message of this header
		}
		return trimmed, index + 1
	}
	return "", 0
}

func (p *parser) addCMake(kind, message string, location *domain.SourceLocation, raw []string) {
	severity := domain.SeverityError
	if kind != cmakeSevere {
		severity = domain.SeverityWarning
	}
	p.add(domain.Diagnostic{Location: location, Severity: severity, Message: message, Source: sourceCMake}, raw, true)
	last := &p.diagnostics[len(p.diagnostics)-1]
	if last.Code == "" && severity == domain.SeverityError {
		last.Code = genericCMakeError
	}
}
