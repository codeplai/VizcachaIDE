//go:build !windows

package toolchain

import (
	"os/exec"
	"syscall"
	"time"
)

const killGrace = 2 * time.Second

// prepareTree puts the process in its own group so the whole tree can be signalled.
func prepareTree(cmd *exec.Cmd) {
	cmd.SysProcAttr = &syscall.SysProcAttr{Setpgid: true}
}

// killTree ends the whole process group at once.
func killTree(pid int) {
	_ = syscall.Kill(-pid, syscall.SIGKILL)
}

// terminateTree asks the group to stop (TERM) and kills it if it is still alive
// after two seconds. done is closed when the process has ended.
func terminateTree(pid int, done <-chan struct{}) {
	_ = syscall.Kill(-pid, syscall.SIGTERM)
	time.AfterFunc(killGrace, func() {
		select {
		case <-done:
		default:
			killTree(pid)
		}
	})
}

// hideConsole is only needed on Windows.
func hideConsole(*exec.Cmd) {}
