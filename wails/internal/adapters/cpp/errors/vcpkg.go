package errors

import (
	"regexp"
	"strings"

	"github.com/codeplai/VizcachaIDE/wails/internal/domain"
	"github.com/codeplai/VizcachaIDE/wails/internal/protocol/errorcatalog"
)

const sourceVcpkg = "vcpkg"

// vcpkgIDPrefix is the prefix of the catalog ids of the libraries (docs/PLAN_CPP_CMAKE.md 4.2).
const vcpkgIDPrefix = "CPP-VCPKG-"

// Shapes of what vcpkg and CMake print while the libraries of the project are installed:
//
//	error: building fmt:x64-mingw-static failed with: BUILD_FAILED
//	C:/vcpkg/ports/nope: error: nope does not exist
//	CMake Error at CMakeLists.txt:12 (find_package):
//	  Could not find a package configuration file provided by "fmt" ...
var (
	vcpkgError = regexp.MustCompile(`^(?:.*?: )?error: (?P<message>.+)$`)
	cmakeError = regexp.MustCompile(`^CMake Error at (?P<path>.+?):(?P<line>\d+) \(`)
)

// vcpkgLine reads an error of vcpkg, or an error of CMake about a library, when the catalog has an
// explanation for it. Any other line is left to the other readers: the catalog is what says that a
// message is about a library, so an unrelated "error: ..." never becomes a diagnostic here.
func (p *parser) vcpkgLine(lines []string) (int, bool) {
	line := strings.TrimRight(lines[0], " \t\r")
	if groups := errorcatalog.NamedGroups(vcpkgError, line); groups != nil && p.aboutLibraries(groups["message"]) {
		p.add(domain.Diagnostic{
			Severity: domain.SeverityError, Message: groups["message"], Source: sourceVcpkg,
		}, []string{line}, true)
		return 0, true
	}
	groups := errorcatalog.NamedGroups(cmakeError, line)
	if groups == nil || len(lines) < 2 {
		return 0, false
	}
	message := strings.TrimSpace(lines[1])
	if !p.aboutLibraries(message) {
		return 0, false
	}
	location := errorcatalog.LocationFrom(groups, p.workingDir)
	p.add(domain.Diagnostic{
		Location: &location, Severity: domain.SeverityError, Message: message, Source: sourceVcpkg,
	}, []string{line, lines[1]}, true)
	return 1, true
}

func (p *parser) aboutLibraries(message string) bool {
	return strings.HasPrefix(p.identify(message), vcpkgIDPrefix)
}
