//go:build windows

package pty

import "syscall"

var procCtrlHandler = syscall.NewLazyDLL("kernel32.dll").NewProc("SetConsoleCtrlHandler")

// acceptCtrlC makes the programs started from now on react to Ctrl+C. Windows passes the
// "ignore Ctrl+C" flag of a process on to every process it starts, and the IDE often inherits
// it from whoever launched it (a process started in a new process group gets it). ConPTY turns
// the ETX typed in the terminal into a CTRL_C_EVENT, which a program carrying that flag
// ignores: an endless loop could not be stopped with Ctrl+C. Clearing the flag in the IDE
// before each start fixes it; the IDE itself has no console, so it never receives the event.
func acceptCtrlC() {
	_, _, _ = procCtrlHandler.Call(0, 0)
}
