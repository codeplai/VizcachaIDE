package app

import (
	"context"
	"path/filepath"

	"github.com/codeplai/VizcachaIDE/wails/internal/domain"
)

// UnavailableSupport is the support of a language whose adapters do not exist yet (Python and
// C++ in 2.1). Its profile keeps the real capabilities so menus exist, but running, debugging
// and building answer ErrUnsupported, the language server has nothing to say and every optional
// port is nil. NewLanguageRegistry skips the capability check for it.
func UnavailableSupport(profile domain.LanguageProfile) LanguageSupport {
	return LanguageSupport{
		Profile:        profile,
		Runner:         unavailableRunner{language: profile.ID},
		Debugger:       unavailableDebugger{},
		LanguageServer: unavailableServer{},
		Explainer:      unavailableExplainer{},
		unavailable:    true,
	}
}

type unavailableRunner struct{ language domain.CodeLanguage }

func (r unavailableRunner) Configure(path string, programArgs []string) domain.RunConfiguration {
	if absolute, err := filepath.Abs(path); err == nil {
		path = absolute
	}
	return domain.NewFileRunConfiguration(r.language, path, programArgs)
}

func (unavailableRunner) Run(context.Context, domain.RunConfiguration) error   { return ErrUnsupported }
func (unavailableRunner) Build(context.Context, domain.RunConfiguration) error { return ErrUnsupported }
func (unavailableRunner) RunUntitled(context.Context, string, string, []string) (domain.RunConfiguration, error) {
	return domain.RunConfiguration{}, ErrUnsupported
}

// Stop and WriteInput have nothing to act on: this runner never starts a process.
func (unavailableRunner) Stop() error                               { return nil }
func (unavailableRunner) IsRunning() bool                           { return false }
func (unavailableRunner) WriteInput(string) error                   { return nil }
func (unavailableRunner) Tools(context.Context) []domain.ToolStatus { return nil }
func (unavailableRunner) Environment() map[string]string            { return map[string]string{} }

type unavailableDebugger struct{}

func (unavailableDebugger) Start(context.Context, domain.RunConfiguration, []domain.Breakpoint) error {
	return ErrUnsupported
}
func (unavailableDebugger) SetBreakpoints(string, []int) error { return ErrUnsupported }
func (unavailableDebugger) StepOver() error                    { return ErrUnsupported }
func (unavailableDebugger) StepInto() error                    { return ErrUnsupported }
func (unavailableDebugger) StepOut() error                     { return ErrUnsupported }
func (unavailableDebugger) Resume() error                      { return ErrUnsupported }
func (unavailableDebugger) RunTo(domain.SourceLocation) error  { return ErrUnsupported }
func (unavailableDebugger) RequestVariables(int) error         { return ErrUnsupported }
func (unavailableDebugger) Stop() error                        { return ErrUnsupported }
func (unavailableDebugger) IsActive() bool                     { return false }
func (unavailableDebugger) FrameVariables(int) (domain.FrameVariables, error) {
	return domain.FrameVariables{}, ErrUnsupported
}

// unavailableServer follows the LanguageServer contract for a missing server: every query is
// empty and nothing fails.
type unavailableServer struct{}

func (unavailableServer) OpenDocument(context.Context, string, string) error        { return nil }
func (unavailableServer) ChangeDocument(context.Context, string, string, int) error { return nil }
func (unavailableServer) CloseDocument(context.Context, string) error               { return nil }
func (unavailableServer) Completion(context.Context, domain.SourceLocation) ([]domain.CompletionItem, error) {
	return []domain.CompletionItem{}, nil
}
func (unavailableServer) Hover(context.Context, domain.SourceLocation) (string, error) {
	return "", nil
}
func (unavailableServer) Definition(context.Context, domain.SourceLocation) (*domain.SourceLocation, error) {
	return nil, nil
}
func (unavailableServer) SignatureHelp(context.Context, domain.SourceLocation) (*domain.SignatureHelp, error) {
	return nil, nil
}
func (unavailableServer) DocumentHighlights(context.Context, domain.SourceLocation) ([]domain.SourceRange, error) {
	return []domain.SourceRange{}, nil
}
func (unavailableServer) DocumentSymbols(context.Context, string) ([]domain.DocumentSymbol, error) {
	return []domain.DocumentSymbol{}, nil
}
func (unavailableServer) InlayHints(context.Context, domain.SourceRange) ([]domain.InlayHint, error) {
	return []domain.InlayHint{}, nil
}
func (unavailableServer) Shutdown(context.Context) error { return nil }

type unavailableExplainer struct{}

func (unavailableExplainer) Parse(string, string) []domain.Diagnostic { return nil }
func (unavailableExplainer) Explain(domain.Diagnostic, string) *domain.ErrorExplanation {
	return nil
}
