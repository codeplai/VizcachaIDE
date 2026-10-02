// Package bridge holds the thin services exposed to the frontend through Wails and the
// event contract. It is the only package that may import the Wails runtime.
package bridge

// Event names: the contract between backend and frontend (PLAN_WAILS.md section 2.2).
// FROZEN: frontend/src/lib/events.ts mirrors these constants and a test keeps them equal.
// Payloads are documented next to each constant; the frontend types are in events.ts.
const (
	// EventRunOutput carries RunOutputPayload.
	EventRunOutput = "run:output"
	// EventRunStarted carries a domain.RunConfiguration.
	EventRunStarted = "run:started"
	// EventRunFinished carries RunFinishedPayload.
	EventRunFinished = "run:finished"
	// EventDebugStopped carries a domain.DebugState.
	EventDebugStopped = "debug:stopped"
	// EventDebugVariables carries DebugVariablesPayload (asynchronous expansion).
	EventDebugVariables = "debug:variables"
	// EventDebugOutput carries DebugOutputPayload.
	EventDebugOutput = "debug:output"
	// EventDebugTerminated carries DebugTerminatedPayload. ExitCode -1 means stopped by the user.
	EventDebugTerminated = "debug:terminated"
	// EventLspDiagnostics carries DiagnosticsPayload.
	EventLspDiagnostics = "lsp:diagnostics"
	// EventLspStatus carries a domain.ServerStatus string.
	EventLspStatus = "lsp:status"
	// EventAssistantExplained carries []domain.ExplainedDiagnostic with translated texts.
	EventAssistantExplained = "assistant:explained"
	// EventSettingsChanged carries a domain.Settings.
	EventSettingsChanged = "settings:changed"
	// EventFileChanged carries FileChangedPayload: an open file changed outside the IDE.
	EventFileChanged = "file:changed"
)

// AllEvents lists every event name, in declaration order.
var AllEvents = []string{
	EventRunOutput,
	EventRunStarted,
	EventRunFinished,
	EventDebugStopped,
	EventDebugVariables,
	EventDebugOutput,
	EventDebugTerminated,
	EventLspDiagnostics,
	EventLspStatus,
	EventAssistantExplained,
	EventSettingsChanged,
	EventFileChanged,
}
