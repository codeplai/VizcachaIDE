package bridge

import (
	"context"

	"github.com/codeplai/VizcachaIDE/wails/internal/app"
	"github.com/codeplai/VizcachaIDE/wails/internal/domain"
)

// CodeLanguagesService tells the frontend which programming languages exist and which of their
// tools were found. Not to be confused with LanguageService, which is the language server (LSP).
type CodeLanguagesService struct {
	registry *app.LanguageRegistry
}

// NewCodeLanguagesService creates the service.
func NewCodeLanguagesService(registry *app.LanguageRegistry) *CodeLanguagesService {
	return &CodeLanguagesService{registry: registry}
}

// Profiles returns the profile of every language (extensions, capabilities, tools).
func (s *CodeLanguagesService) Profiles() []domain.LanguageProfile {
	return s.registry.Profiles()
}

// Tools returns the status of every tool of every language.
func (s *CodeLanguagesService) Tools() []domain.ToolStatus {
	return toolStatuses(context.Background(), s.registry)
}

// toolStatuses asks every runner for its tools. The result is never nil.
func toolStatuses(ctx context.Context, registry *app.LanguageRegistry) []domain.ToolStatus {
	statuses := []domain.ToolStatus{}
	for _, support := range registry.All() {
		statuses = append(statuses, support.Runner.Tools(ctx)...)
	}
	return statuses
}
