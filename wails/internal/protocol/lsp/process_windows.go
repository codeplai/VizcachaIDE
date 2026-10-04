//go:build windows

package lsp

import (
	"os/exec"
	"syscall"
)

// hideWindow keeps the server from opening a console window.
func hideWindow(cmd *exec.Cmd) {
	cmd.SysProcAttr = &syscall.SysProcAttr{HideWindow: true, CreationFlags: 0x08000000}
}
