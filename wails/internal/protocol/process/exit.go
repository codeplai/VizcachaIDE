package process

import (
	"os"
	"syscall"
)

// exitCodeOf is the exit code of a finished process. A Unix process killed by a signal has no
// exit code (os/exec says -1): it is reported as -signal, so the languages can name the crash
// (Job.Finished). Windows never reports a signal: its crashes are NTSTATUS exit codes.
func exitCodeOf(state *os.ProcessState) int {
	if status, ok := state.Sys().(syscall.WaitStatus); ok && status.Signaled() {
		return -int(status.Signal())
	}
	return state.ExitCode()
}
