//go:build !windows

package process

import (
	"os/exec"
	"syscall"
	"time"
)

// prepareTree puts the process in its own group so the whole tree can be signalled.
func prepareTree(cmd *exec.Cmd) {
	cmd.SysProcAttr = &syscall.SysProcAttr{Setpgid: true}
}

// killTree ends the whole process group at once.
func killTree(pid int) {
	_ = syscall.Kill(-pid, syscall.SIGKILL)
}

// terminateTree asks the group to stop (INT, as Ctrl+C does, so defers and signal handlers run)
// and kills it if it is still alive after two seconds. ended is closed when the process has ended.
func terminateTree(pid int, ended <-chan struct{}) {
	_ = syscall.Kill(-pid, syscall.SIGINT)
	time.AfterFunc(killGrace, func() {
		select {
		case <-ended:
		default:
			killTree(pid)
		}
	})
}

// HideConsole is only needed on Windows.
func HideConsole(*exec.Cmd) {}
