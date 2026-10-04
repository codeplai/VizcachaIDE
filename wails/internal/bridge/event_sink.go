package bridge

import (
	"context"
	"sync"

	"github.com/codeplai/VizcachaIDE/wails/internal/app"
	"github.com/codeplai/VizcachaIDE/wails/internal/domain"
	"github.com/wailsapp/wails/v2/pkg/runtime"
)

// WailsEventSink implements app.EventSink on top of the Wails runtime.
// It is the only place where domain values become frontend events.
type WailsEventSink struct {
	mu  sync.RWMutex
	ctx context.Context
}

var (
	_ app.EventSink      = (*WailsEventSink)(nil)
	_ app.FileChangeSink = (*WailsEventSink)(nil)
	_ app.UpdateSink     = (*WailsEventSink)(nil)
)

// UpdateState implements app.UpdateSink.
func (s *WailsEventSink) UpdateState(state domain.UpdateState) { s.emit(EventUpdateState, state) }

// Quit closes the application (after the update installer started).
func (s *WailsEventSink) Quit() {
	if ctx := s.Context(); ctx != nil {
		runtime.Quit(ctx)
	}
}

// NewWailsEventSink creates a sink that drops events until SetContext is called.
func NewWailsEventSink() *WailsEventSink { return &WailsEventSink{} }

// SetContext gives the sink the Wails application context (call it from OnStartup).
func (s *WailsEventSink) SetContext(ctx context.Context) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.ctx = ctx
}

// Context returns the Wails application context, or nil before OnStartup.
func (s *WailsEventSink) Context() context.Context {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.ctx
}

func (s *WailsEventSink) emit(name string, payload any) {
	s.mu.RLock()
	ctx := s.ctx
	s.mu.RUnlock()
	if ctx == nil {
		return
	}
	runtime.EventsEmit(ctx, name, payload)
}

// RunOutput implements app.EventSink.
func (s *WailsEventSink) RunOutput(stream, text string) {
	s.emit(EventRunOutput, RunOutputPayload{Stream: stream, Text: text})
}

// RunStarted implements app.EventSink.
func (s *WailsEventSink) RunStarted(config domain.RunConfiguration) {
	s.emit(EventRunStarted, config)
}

// RunFinished implements app.EventSink.
func (s *WailsEventSink) RunFinished(exitCode int, durationMs int64) {
	s.emit(EventRunFinished, RunFinishedPayload{ExitCode: exitCode, DurationMs: durationMs})
}

// DebugStopped implements app.EventSink.
func (s *WailsEventSink) DebugStopped(state domain.DebugState) {
	s.emit(EventDebugStopped, state)
}

// DebugVariables implements app.EventSink.
func (s *WailsEventSink) DebugVariables(reference int, variables []domain.Variable) {
	s.emit(EventDebugVariables, DebugVariablesPayload{Reference: reference, Variables: variables})
}

// DebugOutput implements app.EventSink.
func (s *WailsEventSink) DebugOutput(text, category string) {
	s.emit(EventDebugOutput, DebugOutputPayload{Text: text, Category: category})
}

// DebugTerminated implements app.EventSink.
func (s *WailsEventSink) DebugTerminated(exitCode int) {
	s.emit(EventDebugTerminated, DebugTerminatedPayload{ExitCode: exitCode})
}

// Diagnostics implements app.EventSink.
func (s *WailsEventSink) Diagnostics(path string, diagnostics []domain.Diagnostic) {
	s.emit(EventLspDiagnostics, DiagnosticsPayload{Path: path, Diagnostics: diagnostics})
}

// LanguageServerStatus implements app.EventSink.
func (s *WailsEventSink) LanguageServerStatus(codeLanguage domain.CodeLanguage, status domain.ServerStatus) {
	s.emit(EventLspStatus, LspStatusPayload{CodeLanguage: codeLanguage, Status: status})
}

// Explained implements app.EventSink.
func (s *WailsEventSink) Explained(items []domain.ExplainedDiagnostic) {
	s.emit(EventAssistantExplained, items)
}

// SettingsChanged implements app.EventSink.
func (s *WailsEventSink) SettingsChanged(settings domain.Settings) {
	s.emit(EventSettingsChanged, settings)
}

// FileChanged implements app.FileChangeSink.
func (s *WailsEventSink) FileChanged(path string) {
	s.emit(EventFileChanged, FileChangedPayload{Path: path})
}
