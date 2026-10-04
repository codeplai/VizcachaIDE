package runner

import "runtime"

// Lines a terminal prints when a program dies. They are terminal text, never translated: the
// error catalog (errors/) recognises them.
const (
	lineSegfault      = "Segmentation fault"
	lineStackOverflow = "Stack overflow"
	lineFloatingPoint = "Floating point exception"
	lineAborted       = "Aborted"
)

// crashLine is Job.Finished of the program stage: the line for an exit code that means a crash,
// "" for a normal exit.
func crashLine(exitCode int) string { return crashLineFor(runtime.GOOS, exitCode) }

func crashLineFor(goos string, exitCode int) string {
	if goos == "windows" {
		return windowsCrash(uint32(exitCode))
	}
	return unixCrash(exitCode)
}

// windowsCrash maps the NTSTATUS exception codes Windows uses as exit code.
func windowsCrash(status uint32) string {
	switch status {
	case 0xC0000005: // access violation
		return lineSegfault
	case 0xC00000FD: // stack overflow
		return lineStackOverflow
	case 0xC0000094: // integer divide by zero
		return lineFloatingPoint
	case 0xC0000409, 3: // fast fail (terminate, abort) and the CRT's abort()
		return lineAborted
	}
	return ""
}

// unixCrash maps a death by signal: the supervisor reports it as -signal, and a shell as
// 128+signal. SIGSEGV with an exhausted stack is a segmentation fault too (docs/PLAN_CPP.md
// section 4.3).
func unixCrash(exitCode int) string {
	const (
		sigabrt, sigbus, sigfpe, sigsegv = 6, 7, 8, 11
		shell                            = 128
	)
	signal := exitCode - shell
	if exitCode < 0 {
		signal = -exitCode
	}
	switch signal {
	case sigsegv, sigbus:
		return lineSegfault
	case sigfpe:
		return lineFloatingPoint
	case sigabrt:
		return lineAborted
	}
	return ""
}
