package runner

import "runtime"

// Lines a terminal prints when a program dies without a panic. They are terminal text, never
// translated: the error catalog (errors/) recognises them. A panic exits with 101 and has
// already printed its message, so it gets no line here.
const (
	lineSegfault = "Segmentation fault"
	lineAborted  = "Aborted"
)

// crashLine is Job.Finished of the program stage: the line for an exit code that means a crash,
// "" for a normal exit and for a panic (101).
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
	case 0xC0000409, 3: // fast fail (abort) and the C runtime's abort()
		return lineAborted
	}
	return ""
}

// unixCrash maps a death by signal: the supervisor reports it as -signal, and a shell as
// 128+signal.
func unixCrash(exitCode int) string {
	const (
		sigabrt, sigbus, sigsegv = 6, 7, 11
		shell                    = 128
	)
	signal := exitCode - shell
	if exitCode < 0 {
		signal = -exitCode
	}
	switch signal {
	case sigsegv, sigbus:
		return lineSegfault
	case sigabrt:
		return lineAborted
	}
	return ""
}
