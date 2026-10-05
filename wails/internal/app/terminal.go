package app

import (
	"context"
	"slices"
)

// TerminalSink is told what the integrated terminals print and when one ends. The bridge turns
// each call into "terminal:output" and "terminal:exit". It must not block.
type TerminalSink interface {
	// TerminalOutput carries raw terminal text (escape sequences included, never a split rune).
	TerminalOutput(id, data string)
	// TerminalExit tells that the shell of a session ended.
	TerminalExit(id string, exitCode int)
}

// TerminalHost runs the shells of the integrated terminal. Sessions are independent of the
// process supervisor, so the "one program at a time" rule does not apply to them.
type TerminalHost interface {
	// Start opens a shell in dir (the user's home when dir is not a folder) and returns its id.
	Start(dir string, cols, rows int) (string, error)
	Write(id, data string) error
	Resize(id string, cols, rows int) error
	// Close ends the shell and everything it started. Closing an unknown id is not an error.
	Close(id string) error
	// CloseAll closes every session (the application is shutting down).
	CloseAll()
}

// ShellPaths is an optional port of a language: the folders of its tools that the integrated
// terminal puts first in PATH, so the commands work there as they do when the IDE runs a program.
type ShellPaths interface {
	ShellPaths(ctx context.Context) []string
}

// ShellVariables is an optional port of a language: environment variables ("NAME=value") of its
// tools that the integrated terminal starts with (VCPKG_ROOT for "cmake --preset debug").
type ShellVariables interface {
	ShellVariables(ctx context.Context) []string
}

// ShellVariables gathers the variables of every language that offers them.
func (r *LanguageRegistry) ShellVariables(ctx context.Context) []string {
	var variables []string
	for _, support := range r.supports {
		if provider, ok := support.Shell.(ShellVariables); ok {
			variables = append(variables, provider.ShellVariables(ctx)...)
		}
	}
	return variables
}

// ShellPaths gathers the folders of every language that offers them, without repeats, in the
// order of the languages.
func (r *LanguageRegistry) ShellPaths(ctx context.Context) []string {
	var paths []string
	for _, support := range r.supports {
		if support.Shell == nil {
			continue
		}
		for _, dir := range support.Shell.ShellPaths(ctx) {
			if dir != "" && !slices.Contains(paths, dir) {
				paths = append(paths, dir)
			}
		}
	}
	return paths
}
