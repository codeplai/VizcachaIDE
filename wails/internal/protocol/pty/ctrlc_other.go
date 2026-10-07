//go:build !windows

package pty

// acceptCtrlC has nothing to do outside Windows: a Unix terminal delivers SIGINT itself.
func acceptCtrlC() {}
