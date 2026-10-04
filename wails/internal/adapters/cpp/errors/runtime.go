package errors

import (
	"regexp"
	"strings"

	"github.com/codeplai/VizcachaIDE/wails/internal/domain"
	"github.com/codeplai/VizcachaIDE/wails/internal/protocol/errorcatalog"
)

// Shapes of the runtime failures:
//
//	Segmentation fault                                   (the line the runner writes when the program crashes)
//	terminate called after throwing an instance of 'std::runtime_error'
//	  what():  algo salio mal                            (libstdc++, GCC)
//	libc++abi: terminating due to uncaught exception of type std::runtime_error: algo salio mal   (libc++, Clang)
var (
	crashText        = regexp.MustCompile(`^(?:Segmentation fault|Stack overflow|Floating point exception|Aborted)(?: \(core dumped\))?$`)
	gnuTerminate     = regexp.MustCompile(`^terminate called after throwing an instance of '(?P<type>[^']+)'$`)
	gnuWhat          = regexp.MustCompile(`^\s*what\(\):\s*(?P<what>.*)$`)
	libcxxTerminates = regexp.MustCompile(`^libc\+\+abi: terminating due to uncaught exception of type `)
)

const abortedLine = "Aborted"

// crashLine reads the line the runner writes when the program ends by a signal or an exception
// code. "Aborted" after an uncaught exception is part of that exception, not another error.
func (p *parser) crashLine(lines []string) (int, bool) {
	line := strings.TrimRight(lines[0], " \t\r")
	if !crashText.MatchString(line) {
		return 0, false
	}
	if strings.HasPrefix(line, abortedLine) && p.lastIsRuntime() {
		return 0, true
	}
	p.addRuntime(line, []string{line})
	return 0, true
}

// gnuTerminateLine reads the uncaught exception of libstdc++ and its what() line. The message
// keeps both ("...'std::runtime_error': algo salio mal") so it reads like the one of libc++.
func (p *parser) gnuTerminateLine(lines []string) (int, bool) {
	line := strings.TrimRight(lines[0], " \t\r")
	if !gnuTerminate.MatchString(line) {
		return 0, false
	}
	message, raw, extra := line, []string{line}, 0
	if len(lines) > 1 {
		if what := errorcatalog.NamedGroups(gnuWhat, strings.TrimRight(lines[1], " \t\r")); what != nil {
			message += ": " + strings.TrimSpace(what["what"])
			raw, extra = append(raw, lines[1]), 1
		}
	}
	p.addRuntime(message, raw)
	return extra, true
}

// libcxxTerminateLine reads the uncaught exception of libc++, which carries what() in one line.
func (p *parser) libcxxTerminateLine(lines []string) (int, bool) {
	line := strings.TrimRight(lines[0], " \t\r")
	if !libcxxTerminates.MatchString(line) {
		return 0, false
	}
	p.addRuntime(line, []string{line})
	return 0, true
}

func (p *parser) addRuntime(message string, raw []string) {
	p.add(domain.Diagnostic{
		Severity: domain.SeverityError,
		Message:  message,
		Source:   sourceRuntime,
	}, raw, false)
}

func (p *parser) lastIsRuntime() bool {
	last := len(p.diagnostics) - 1
	return last >= 0 && p.diagnostics[last].Source == sourceRuntime
}
