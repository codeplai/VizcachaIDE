package bridge

import (
	"context"

	"github.com/codeplai/VizcachaIDE/wails/internal/app"
	"github.com/codeplai/VizcachaIDE/wails/internal/domain"
)

// LanguageService gives code intelligence by delegating to app.LanguageServer (gopls).
// The adapter emits lsp:diagnostics and lsp:status through the EventSink.
type LanguageService struct {
	server app.LanguageServer
}

// NewLanguageService creates the service.
func NewLanguageService(server app.LanguageServer) *LanguageService {
	return &LanguageService{server: server}
}

// OpenDocument tells the language server about a file the user opened.
func (s *LanguageService) OpenDocument(path, text string) error {
	return s.server.OpenDocument(context.Background(), path, text)
}

// ChangeDocument sends the new text of an open file.
func (s *LanguageService) ChangeDocument(path, text string, version int) error {
	return s.server.ChangeDocument(context.Background(), path, text, version)
}

// CloseDocument tells the language server that a file was closed.
func (s *LanguageService) CloseDocument(path string) error {
	return s.server.CloseDocument(context.Background(), path)
}

// Completion returns code suggestions at a position.
func (s *LanguageService) Completion(at domain.SourceLocation) ([]domain.CompletionItem, error) {
	return s.server.Completion(context.Background(), at)
}

// Hover returns the documentation at a position.
func (s *LanguageService) Hover(at domain.SourceLocation) (string, error) {
	return s.server.Hover(context.Background(), at)
}

// Definition returns where the symbol at a position is declared.
func (s *LanguageService) Definition(at domain.SourceLocation) (*domain.SourceLocation, error) {
	return s.server.Definition(context.Background(), at)
}

// SignatureHelp returns the call tip at a position.
func (s *LanguageService) SignatureHelp(at domain.SourceLocation) (*domain.SignatureHelp, error) {
	return s.server.SignatureHelp(context.Background(), at)
}

// DocumentHighlights returns the occurrences of the symbol at a position.
func (s *LanguageService) DocumentHighlights(at domain.SourceLocation) ([]domain.SourceRange, error) {
	return s.server.DocumentHighlights(context.Background(), at)
}

// DocumentSymbols returns the declarations of a file for the Outline.
func (s *LanguageService) DocumentSymbols(path string) ([]domain.DocumentSymbol, error) {
	return s.server.DocumentSymbols(context.Background(), path)
}
