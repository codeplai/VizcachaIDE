//go:build !windows

package gopls

import "os/exec"

// hideWindow does nothing outside Windows.
func hideWindow(*exec.Cmd) {}
