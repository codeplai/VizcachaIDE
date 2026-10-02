package bridge

import "github.com/codeplai/VizcachaIDE/wails/internal/domain"

// RunOutputPayload is the payload of EventRunOutput. Stream is "stdout" or "stderr".
type RunOutputPayload struct {
	Stream string `json:"stream"`
	Text   string `json:"text"`
}

// RunFinishedPayload is the payload of EventRunFinished.
type RunFinishedPayload struct {
	ExitCode   int   `json:"exitCode"`
	DurationMs int64 `json:"durationMs"`
}

// DebugVariablesPayload is the payload of EventDebugVariables.
type DebugVariablesPayload struct {
	Reference int               `json:"reference"`
	Variables []domain.Variable `json:"variables"`
}

// DebugOutputPayload is the payload of EventDebugOutput. Category is
// "stdout", "stderr" or "console".
type DebugOutputPayload struct {
	Text     string `json:"text"`
	Category string `json:"category"`
}

// DebugTerminatedPayload is the payload of EventDebugTerminated.
type DebugTerminatedPayload struct {
	ExitCode int `json:"exitCode"`
}

// DiagnosticsPayload is the payload of EventLspDiagnostics.
type DiagnosticsPayload struct {
	Path        string              `json:"path"`
	Diagnostics []domain.Diagnostic `json:"diagnostics"`
}

// FileChangedPayload is the payload of EventFileChanged.
type FileChangedPayload struct {
	Path string `json:"path"`
}
