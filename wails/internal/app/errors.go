package app

import (
	"errors"
	"fmt"
)

// Sentinel errors shared by ports and use cases. Wrap them with %w and test with errors.Is.
var (
	// ErrBusy means another process is already running.
	ErrBusy = errors.New("another program is already running")
	// ErrToolNotFound means a tool of a language (go, dlv, python, clangd...) could not be
	// located. Return it through MissingTool so the frontend knows which tool is missing.
	ErrToolNotFound = errors.New("tool not found")
	// ErrFormat means the formatter could not parse the source.
	ErrFormat = errors.New("source cannot be formatted")
	// ErrNoSession means there is no active debug session.
	ErrNoSession = errors.New("no active debug session")
	// ErrUnsupported means the language of the file does not offer the action (Build in
	// Python, a console in C++...).
	ErrUnsupported = errors.New("this language does not support the action")
	// ErrUnknownCodeLanguage means no registered language matches a file or an id.
	ErrUnknownCodeLanguage = errors.New("unknown code language")
	// ErrInconsistentProfile means a language profile and its support disagree (a capability
	// without its port, two languages with the same extension...). The IDE fails at start-up.
	ErrInconsistentProfile = errors.New("inconsistent language profile")
)

// MissingTool wraps ErrToolNotFound with the ToolSpec.ID of the tool, in the form the
// frontend reads: `tool "<id>": tool not found`.
func MissingTool(toolID string) error {
	return fmt.Errorf("tool %q: %w", toolID, ErrToolNotFound)
}
