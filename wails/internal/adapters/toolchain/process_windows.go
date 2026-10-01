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

const taskkillTimeout = 5 * time.Second

// hideConsole keeps a console window from flashing when the IDE starts a process.
func hideConsole(cmd *exec.Cmd) {
	cmd.SysProcAttr = &syscall.SysProcAttr{HideWindow: true, CreationFlags: createNoWindow}
}

// prepareTree is a no-op on Windows: taskkill /T follows the process tree by itself.
func prepareTree(cmd *exec.Cmd) { hideConsole(cmd) }

// killTree ends pid and all its descendants at once (taskkill /T /F).
func killTree(pid int) {
	ctx, cancel := context.WithTimeout(context.Background(), taskkillTimeout)
	defer cancel()
	cmd := exec.CommandContext(ctx, "taskkill", "/PID", strconv.Itoa(pid), "/T", "/F")
	hideConsole(cmd)
	_ = cmd.Run() // best effort: the caller also kills the direct child
}

// terminateTree stops the process tree. Windows has no graceful signal for console
// programs, so it is the same as killTree.
func terminateTree(pid int, _ <-chan struct{}) { killTree(pid) }
