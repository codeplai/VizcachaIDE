//go:build windows

package gopls

import (
	"os/exec"
	"syscall"
)

// hideWindow keeps gopls from opening a console window.
func hideWindow(cmd *exec.Cmd) {
	cmd.SysProcAttr = &syscall.SysProcAttr{HideWindow: true, CreationFlags: 0x08000000}
}
