package app

import "errors"

// Sentinel errors shared by ports and use cases. Wrap them with %w and test with errors.Is.
var (
	// ErrBusy means another process is already running.
	ErrBusy = errors.New("another program is already running")
	// ErrToolNotFound means Go, Delve or gopls could not be located.
	ErrToolNotFound = errors.New("tool not found")
	// ErrFormat means gofmt could not parse the source.
	ErrFormat = errors.New("source cannot be formatted")
	// ErrNoSession means there is no active debug session.
	ErrNoSession = errors.New("no active debug session")
)
