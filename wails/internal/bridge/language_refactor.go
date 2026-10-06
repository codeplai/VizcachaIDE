package bridge

import (
	"context"

	"github.com/codeplai/VizcachaIDE/wails/internal/domain"
)

// PrepareRename tells whether the symbol at a position can be renamed, and which text is the name.
func (s *LanguageService) PrepareRename(at domain.SourceLocation) (domain.RenameTarget, error) {
	server := s.serverFor(at.File)
	if server == nil {
		return domain.RenameTarget{Refusal: domain.RenameUnsupported}, nil
	}
	return server.PrepareRename(context.Background(), at)
}

// Rename returns the edits that rename the symbol at a position in every affected file. It
// writes nothing: open files are edited by the editor, the others with FilesService.ApplyTextEdits.
func (s *LanguageService) Rename(at domain.SourceLocation, newName string) (domain.RenameResult, error) {
	server := s.serverFor(at.File)
	if server == nil {
		return domain.RenameResult{Refusal: domain.RenameUnsupported, Files: []domain.FileEdit{}}, nil
	}
	return server.Rename(context.Background(), at, newName)
}

// References returns every use of the symbol at a position, declaration included.
func (s *LanguageService) References(at domain.SourceLocation) ([]domain.Reference, error) {
	server := s.serverFor(at.File)
	if server == nil {
		return []domain.Reference{}, nil
	}
	return server.References(context.Background(), at)
}
