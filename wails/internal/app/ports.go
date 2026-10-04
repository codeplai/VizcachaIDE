// Package app holds the use cases of VizcachaIDE and the ports (interfaces) they need
// from the outside world. It imports only the domain package.
package app

import (
	"context"

	"github.com/codeplai/VizcachaIDE/wails/internal/domain"
)

// PORTS: FROZEN CONTRACT (phase W0, revised once as contract v3 in M0).
//
// The language ports are in language_ports.go. The parallel tracks implement these
// interfaces; they do not edit them.
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

// Debugger is the interactive debugger of one language (Delve, debugpy, lldb-dap over DAP).
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
	// FrameVariables returns the arguments and locals of one frame of the paused
	// stack (the Calls view). Empty lists when there is no session or it is running.
	FrameVariables(frameID int) (domain.FrameVariables, error)
	// Stop ends the session. debug:terminated carries domain.TerminatedByUser.
	Stop() error
	// IsActive reports whether a session is alive.
	IsActive() bool
}

// LanguageServer gives code intelligence for one language (gopls, pylsp, clangd over LSP).
//
// Events: lsp:diagnostics (path, diagnostics) whenever the server publishes them,
// lsp:status ("starting", "ready", "unavailable"). If the server is missing, status is
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

// ErrorExplainer turns raw tool output of one language into diagnostics and
// beginner-friendly explanations.
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

// FileChangeSink is told when a watched file changed on disk outside the IDE.
// The bridge turns each call into "file:changed". It must not block.
type FileChangeSink interface {
	// FileChanged emits "file:changed" for the path of the file.
	FileChanged(path string)
}

// FileWatcher notices changes made to open files by other programs.
// Changes made by the IDE itself are reported through Remember and ignored.
type FileWatcher interface {
	// Watch replaces the set of watched files. Bursts of events are merged.
	Watch(paths []string) error
	// Remember tells the watcher the text the IDE just saved, so it is not reported.
	Remember(path, text string)
	// Close stops watching.
	Close() error
}

// Console is the interactive console of a language (the "Shell"): one interpreter session that
// remembers the variables and functions defined by earlier snippets.
//
// It is synchronous and emits no events.
type Console interface {
	// Eval runs a snippet. Failures of the snippet itself (compile errors, panics,
	// timeouts) are reported inside the result, not as a Go error.
	Eval(code string) domain.ConsoleResult
	// Reset throws the session away so the next Eval starts clean.
	Reset()
}
