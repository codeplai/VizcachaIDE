package bridge

import (
	"github.com/codeplai/VizcachaIDE/wails/internal/app"
	"github.com/codeplai/VizcachaIDE/wails/internal/domain"
)

// LanguageService gives code intelligence. W0 STUB: owned by track G3, which
// replaces the body of each method with calls to app.LanguageServer. Emits
// lsp:diagnostics and lsp:status.
type LanguageService struct {
	sink app.EventSink
}

// NewLanguageService creates the service.
func NewLanguageService(sink app.EventSink) *LanguageService { return &LanguageService{sink: sink} }

// OpenDocument tells the language server about a file the user opened.
func (s *LanguageService) OpenDocument(path, text string) error {
	s.sink.LanguageServerStatus(domain.ServerReady)
	return nil
}

// ChangeDocument sends the new text of an open file.
func (s *LanguageService) ChangeDocument(path, text string, version int) error { return nil }

// CloseDocument tells the language server that a file was closed.
func (s *LanguageService) CloseDocument(path string) error { return nil }

// Completion returns code suggestions at a position.
func (s *LanguageService) Completion(at domain.SourceLocation) ([]domain.CompletionItem, error) {
	return []domain.CompletionItem{
		{Label: "Println", Kind: domain.CompletionFunction, Detail: "func(a ...any) (n int, err error)"},
		{Label: "Printf", Kind: domain.CompletionFunction, Detail: "func(format string, a ...any) (n int, err error)"},
	}, nil
}

// Hover returns the documentation at a position.
func (s *LanguageService) Hover(at domain.SourceLocation) (string, error) {
	return "func sumar(a, b int) int", nil
}

// Definition returns where the symbol at a position is declared.
func (s *LanguageService) Definition(at domain.SourceLocation) (*domain.SourceLocation, error) {
	loc := sampleLocation(5, 6)
	return &loc, nil
}

// SignatureHelp returns the call tip at a position.
func (s *LanguageService) SignatureHelp(at domain.SourceLocation) (*domain.SignatureHelp, error) {
	return &domain.SignatureHelp{Label: "sumar(a, b int) int", Parameters: []string{"a", "b int"}}, nil
}

// DocumentHighlights returns the occurrences of the symbol at a position.
func (s *LanguageService) DocumentHighlights(at domain.SourceLocation) ([]domain.SourceRange, error) {
	return []domain.SourceRange{}, nil
}

// DocumentSymbols returns the declarations of a file for the Outline.
func (s *LanguageService) DocumentSymbols(path string) ([]domain.DocumentSymbol, error) {
	return sampleSymbols(), nil
}
