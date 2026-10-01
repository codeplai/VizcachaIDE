// Package app holds the use cases of VizcachaIDE and the ports (interfaces) they need
// from the outside world. It imports only the domain package.
package app

import (
	"context"

	"github.com/codeplai/VizcachaIDE/wails/internal/domain"
)

// PORTS: FROZEN CONTRACT (phase W0).
//
// The parallel tracks of W1 implement these interfaces; they do not edit them.
// A track that needs a change writes a "Contract change request" in its final
// report and the integrator decides.
//
// Adapters never know event names. They receive an EventSink in their
// constructor and call its typed methods; the bridge package turns each call
// into the event of the same row in bridge/events.go (and frontend/src/lib/events.ts).

// EventSink is how adapters and use cases tell the frontend that something happened.
// The bridge package implements it on top of the Wails runtime.
// Methods must not block and must be safe to call from any goroutine.
type EventSink interface {
	// RunOutput emits "run:output". stream is "stdout" or "stderr".
	RunOutput(stream, text string)
	// RunStarted emits "run:started".
	RunStarted(config domain.RunConfiguration)
	// RunFinished emits "run:finished". Also used after a build.
	RunFinished(exitCode int, durationMs int64)
	// DebugStopped emits "debug:stopped" every time the program pauses.
	DebugStopped(state domain.DebugState)
	// DebugVariables emits "debug:variables": the answer to Debugger.RequestVariables.
	DebugVariables(reference int, variables []domain.Variable)
	// DebugOutput emits "debug:output". category is "stdout", "stderr" or "console".
	DebugOutput(text, category string)
	// DebugTerminated emits "debug:terminated". exitCode is domain.TerminatedByUser (-1)
	// when the user stopped the session.
	DebugTerminated(exitCode int)
	// Diagnostics emits "lsp:diagnostics" for one file.
	Diagnostics(path string, diagnostics []domain.Diagnostic)
	// LanguageServerStatus emits "lsp:status".
	LanguageServerStatus(status domain.ServerStatus)
	// Explained emits "assistant:explained" with already translated texts.
	Explained(items []domain.ExplainedDiagnostic)
	// SettingsChanged emits "settings:changed".
	SettingsChanged(settings domain.Settings)
}

// Toolchain runs the user's program with the Go toolchain.
//
// Events (through the EventSink given to the adapter): run:started when a process
// starts, run:output for every stdout/stderr chunk, run:finished when it ends.
type Toolchain interface {
	// Environment returns the environment variables the Go tools run with.
	Environment() map[string]string
	// Info reports which tools were found (config, bundled, PATH) and their versions.
	Info(ctx context.Context) domain.ToolchainInfo
	// Run compiles and runs the configuration. Returns ErrBusy while another run is active.
	Run(ctx context.Context, config domain.RunConfiguration) error
	// Build compiles without running. Emits the same events as Run.
	Build(ctx context.Context, config domain.RunConfiguration) error
	// Stop kills the running program, if any.
	Stop() error
	// IsRunning reports whether a process started by Run, Build or RunGoCommand is alive.
	IsRunning() bool
	// WriteInput sends text to the program's stdin.
	WriteInput(text string) error
	// RunUntitled runs unsaved source from a temporary directory and returns the
	// configuration it used. Returns ErrBusy while another run is active.
	RunUntitled(ctx context.Context, source string, programArgs []string) (domain.RunConfiguration, error)
	// RunGoCommand runs "go <args>" (for example "mod tidy") emitting the same events.
	RunGoCommand(ctx context.Context, workingDir string, args []string) error
	// FormatSource returns gofmt-formatted text. Errors wrap ErrFormat or ErrToolNotFound.
	FormatSource(text string) (string, error)
}

// Debugger is the interactive debugger (Delve over DAP).
//
// Events: debug:stopped (DebugState) each time the program pauses, debug:variables
// (reference, variables) as the answer to RequestVariables, debug:output
// (text, category) for program output, debug:terminated (exitCode) at the end.
type Debugger interface {
	// Start compiles with debug info and launches the program with the given breakpoints.
	Start(ctx context.Context, config domain.RunConfiguration, breakpoints []domain.Breakpoint) error
	// SetBreakpoints replaces the breakpoints of one file (also while running).
	SetBreakpoints(file string, lines []int) error
	StepOver() error
	StepInto() error
	StepOut() error
	// Resume continues until the next breakpoint.
	Resume() error
	// RunTo runs until the given location.
	RunTo(location domain.SourceLocation) error
	// RequestVariables asks for the children of a variable with a non-zero Reference.
	// It never blocks: the answer arrives as debug:variables, always emitted (with an
	// empty list when there is no session or the request fails), except when the
	// program resumes first.
	RequestVariables(reference int) error
	// Stop ends the session. debug:terminated carries domain.TerminatedByUser.
	Stop() error
	// IsActive reports whether a session is alive.
	IsActive() bool
}

// LanguageServer gives code intelligence (gopls over LSP).
//
// Events: lsp:diagnostics (path, diagnostics) whenever the server publishes them,
// lsp:status ("starting", "ready", "unavailable"). If gopls is missing, status is
// "unavailable" and every query returns an empty result without an error.
type LanguageServer interface {
	OpenDocument(ctx context.Context, path, text string) error
	ChangeDocument(ctx context.Context, path, text string, version int) error
	CloseDocument(ctx context.Context, path string) error
	Completion(ctx context.Context, at domain.SourceLocation) ([]domain.CompletionItem, error)
	// Hover returns the documentation at a position, or "" if there is none.
	Hover(ctx context.Context, at domain.SourceLocation) (string, error)
	// Definition returns where the symbol is declared, or nil.
	Definition(ctx context.Context, at domain.SourceLocation) (*domain.SourceLocation, error)
	SignatureHelp(ctx context.Context, at domain.SourceLocation) (*domain.SignatureHelp, error)
	// DocumentHighlights returns the occurrences of the symbol at a position in the same file.
	DocumentHighlights(ctx context.Context, at domain.SourceLocation) ([]domain.SourceRange, error)
	// DocumentSymbols returns the nested declarations of an open document (for the Outline).
	DocumentSymbols(ctx context.Context, path string) ([]domain.DocumentSymbol, error)
	Shutdown(ctx context.Context) error
}

// ErrorExplainer turns raw Go output into diagnostics and beginner-friendly explanations.
//
// It is synchronous and emits no events; AssistantService emits "assistant:explained".
type ErrorExplainer interface {
	// Parse extracts diagnostics (compiler errors, vet warnings, panics) from raw output.
	Parse(rawOutput, workingDir string) []domain.Diagnostic
	// Explain returns the explanation of a diagnostic in the given language ("en" or "es"),
	// or nil when the error is not recognised.
	Explain(diagnostic domain.Diagnostic, language string) *domain.ErrorExplanation
}

// SettingsStore persists the user preferences.
//
// Event: SettingsService emits "settings:changed" after a successful Save.
type SettingsStore interface {
	// Load returns the saved settings, or domain.DefaultSettings when there are none.
	Load() (domain.Settings, error)
	Save(settings domain.Settings) error
}
