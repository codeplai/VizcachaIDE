//go:build windows

package toolchain

import (
	"context"
	"os/exec"
	"strconv"
	"syscall"
	"time"
)

const createNoWindow = 0x08000000

const (
	taskkillTimeout = 5 * time.Second
	killGrace       = 2 * time.Second
)

// hideConsole keeps a console window from flashing when the IDE starts a process.
func hideConsole(cmd *exec.Cmd) {
	cmd.SysProcAttr = &syscall.SysProcAttr{HideWindow: true, CreationFlags: createNoWindow}
}

// prepareTree hides the console and starts a new process group, the target of CTRL_BREAK
// when the program is stopped gracefully. taskkill /T follows the tree by itself.
func prepareTree(cmd *exec.Cmd) {
	hideConsole(cmd)
	newProcessGroup(cmd.SysProcAttr)
}

// killTree ends pid and all its descendants at once (taskkill /T /F).
func killTree(pid int) {
	ctx, cancel := context.WithTimeout(context.Background(), taskkillTimeout)
	defer cancel()
	cmd := exec.CommandContext(ctx, "taskkill", "/PID", strconv.Itoa(pid), "/T", "/F")
	hideConsole(cmd)
	_ = cmd.Run() // best effort: the caller also kills the direct child
}

// terminateTree asks the program to stop (CTRL_BREAK, which Go delivers as os.Interrupt, so
// defers and signal handlers run) and kills the whole tree if it is still alive after two
// seconds. If the signal cannot be sent it kills at once. done is closed when the process ended.
func terminateTree(pid int, done <-chan struct{}) {
	go func() {
		if err := sendCtrlBreak(pid); err != nil {
			killTree(pid)
			return
		}
		select {
		case <-done:
		case <-time.After(killGrace):
			killTree(pid)
		}
	}()
}
