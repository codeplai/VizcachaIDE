//go:build !windows

package delve

import "os/exec"

func hideConsoleWindow(*exec.Cmd) {}
