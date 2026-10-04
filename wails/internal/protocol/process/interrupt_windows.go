//go:build windows

package process

import (
	"fmt"
	"sync"
	"syscall"
)

// ctrlBreakEvent is CTRL_BREAK_EVENT of GenerateConsoleCtrlEvent.
const ctrlBreakEvent = 1

var (
	kernel32         = syscall.NewLazyDLL("kernel32.dll")
	procAttach       = kernel32.NewProc("AttachConsole")
	procFree         = kernel32.NewProc("FreeConsole")
	procCtrlHandler  = kernel32.NewProc("SetConsoleCtrlHandler")
	procGenerate     = kernel32.NewProc("GenerateConsoleCtrlEvent")
	consoleOperation sync.Mutex
)

// sendCtrlBreak sends CTRL_BREAK to the process group that starts at pid, which is how a
// Go program with a signal handler (or defer) gets os.Interrupt on Windows.
//
// GenerateConsoleCtrlEvent only reaches processes on the caller's console. The IDE is a
// window application without a console and the program runs on its own hidden one, so the
// IDE attaches to that console for the instant it takes to send the event. The calls are
// serialized because a process has at most one console.
func sendCtrlBreak(pid int) error {
	consoleOperation.Lock()
	defer consoleOperation.Unlock()
	_, _, _ = procFree.Call() // leave the console we may have (for example when run from a terminal)
	if result, _, err := procAttach.Call(uintptr(pid)); result == 0 {
		return fmt.Errorf("attach to the console of process %d: %w", pid, err)
	}
	defer func() { _, _, _ = procFree.Call() }()
	// Ignore Ctrl+C in the IDE itself while attached; the group below is not ours anyway.
	_, _, _ = procCtrlHandler.Call(0, 1)
	defer func() { _, _, _ = procCtrlHandler.Call(0, 0) }()
	if result, _, err := procGenerate.Call(ctrlBreakEvent, uintptr(pid)); result == 0 {
		return fmt.Errorf("send CTRL_BREAK to process group %d: %w", pid, err)
	}
	return nil
}
