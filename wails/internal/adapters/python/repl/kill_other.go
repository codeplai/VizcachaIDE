//go:build !windows

package repl

// killTree does nothing: on macOS and Linux the interpreter is the process itself, which kill
// ends directly.
func killTree(int) {}
