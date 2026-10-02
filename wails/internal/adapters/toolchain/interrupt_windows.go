//go:build windows

package toolchain

import (
	"fmt"
	"sync"
	"syscall"

	"golang.org/x/sys/windows"
)

var (
	kernel32         = windows.NewLazySystemDLL("kernel32.dll")
	procAttach       = kernel32.NewProc("AttachConsole")
	procFree         = kernel32.NewProc("FreeConsole")
	procCtrlHandler  = kernel32.NewProc("SetConsoleCtrlHandler")
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
	if err := windows.GenerateConsoleCtrlEvent(windows.CTRL_BREAK_EVENT, uint32(pid)); err != nil {
		return fmt.Errorf("send CTRL_BREAK to process group %d: %w", pid, err)
	}
	return nil
}

// newProcessGroup makes the process the root of its own group, so CTRL_BREAK reaches it
// and its children (go run starts the program as a child) and nothing else.
func newProcessGroup(attr *syscall.SysProcAttr) {
	attr.CreationFlags |= windows.CREATE_NEW_PROCESS_GROUP
}
