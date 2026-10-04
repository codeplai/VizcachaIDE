package bridge

import (
	"context"

	"github.com/codeplai/VizcachaIDE/wails/internal/app"
	"github.com/codeplai/VizcachaIDE/wails/internal/domain"
)

// LanguageService gives code intelligence by delegating to the app.LanguageServer of the
// language of the file (gopls for Go). It routes by path or by at.File. A file whose extension
// no language claims has no code intelligence: the answers are empty, never errors. The adapters
// emit lsp:diagnostics and lsp:status through the EventSink.
type LanguageService struct {
	supportRouter
}

// NewLanguageService creates the service.
func NewLanguageService(registry *app.LanguageRegistry) *LanguageService {
	return &LanguageService{supportRouter{registry: registry}}
}

// serverFor returns the language server of a file, or nil when no language claims it.
func (s *LanguageService) serverFor(path string) app.LanguageServer {
	support, ok := s.registry.ForPath(path)
	if !ok {
		return nil
	}
	return support.LanguageServer
}

// OpenDocument tells the language server about a file the user opened.
func (s *LanguageService) OpenDocument(path, text string) error {
	server := s.serverFor(path)
	if server == nil {
		return nil
	}
	return server.OpenDocument(context.Background(), path, text)
}

// ChangeDocument sends the new text of an open file.
func (s *LanguageService) ChangeDocument(path, text string, version int) error {
	server := s.serverFor(path)
	if server == nil {
		return nil
	}
	return server.ChangeDocument(context.Background(), path, text, version)
}

// CloseDocument tells the language server that a file was closed.
func (s *LanguageService) CloseDocument(path string) error {
	server := s.serverFor(path)
	if server == nil {
		return nil
	}
	return server.CloseDocument(context.Background(), path)
}

// Completion returns code suggestions at a position.
func (s *LanguageService) Completion(at domain.SourceLocation) ([]domain.CompletionItem, error) {
	server := s.serverFor(at.File)
	if server == nil {
		return []domain.CompletionItem{}, nil
	}
	return server.Completion(context.Background(), at)
}

// Hover returns the documentation at a position.
func (s *LanguageService) Hover(at domain.SourceLocation) (string, error) {
	server := s.serverFor(at.File)
	if server == nil {
		return "", nil
	}
	return server.Hover(context.Background(), at)
}

// Definition returns where the symbol at a position is declared.
func (s *LanguageService) Definition(at domain.SourceLocation) (*domain.SourceLocation, error) {
	server := s.serverFor(at.File)
	if server == nil {
		return nil, nil
	}
	return server.Definition(context.Background(), at)
}

// SignatureHelp returns the call tip at a position.
func (s *LanguageService) SignatureHelp(at domain.SourceLocation) (*domain.SignatureHelp, error) {
	server := s.serverFor(at.File)
	if server == nil {
		return nil, nil
	}
	return server.SignatureHelp(context.Background(), at)
}

// DocumentHighlights returns the occurrences of the symbol at a position.
func (s *LanguageService) DocumentHighlights(at domain.SourceLocation) ([]domain.SourceRange, error) {
	server := s.serverFor(at.File)
	if server == nil {
		return []domain.SourceRange{}, nil
	}
	return server.DocumentHighlights(context.Background(), at)
}

// DocumentSymbols returns the declarations of a file for the Outline.
func (s *LanguageService) DocumentSymbols(path string) ([]domain.DocumentSymbol, error) {
	server := s.serverFor(path)
	if server == nil {
		return []domain.DocumentSymbol{}, nil
	}
	return server.DocumentSymbols(context.Background(), path)
}
