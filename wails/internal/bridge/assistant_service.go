package bridge

import (
	"fmt"

	"github.com/codeplai/VizcachaIDE/wails/internal/app"
	"github.com/codeplai/VizcachaIDE/wails/internal/domain"
)

// AssistantService explains compiler and runtime errors to beginners, with the explainer of
// each language. It emits assistant:explained with texts already translated to the language of
// the settings.
type AssistantService struct {
	supportRouter
	sink     app.EventSink
	language *LanguageResolver
}

// NewAssistantService creates the service.
func NewAssistantService(sink app.EventSink, registry *app.LanguageRegistry, language *LanguageResolver) *AssistantService {
	return &AssistantService{supportRouter: supportRouter{registry: registry}, sink: sink, language: language}
}

// Explain parses the raw output of a language's tools (go build / go run / go vet, or a panic)
// and emits the explained diagnostics. Only errors and warnings are included, without duplicates.
func (s *AssistantService) Explain(codeLanguage domain.CodeLanguage, rawOutput, workingDir string) ([]domain.ExplainedDiagnostic, error) {
	support, err := s.supportOf(codeLanguage)
	if err != nil {
		return nil, fmt.Errorf("explain: %w", err)
	}
	items := app.NewExplainError(support.Explainer).FromOutput(rawOutput, workingDir, s.language.Current())
	s.sink.Explained(items)
	return items, nil
}

// ExplainDiagnostics explains diagnostics that are already parsed (for example the ones the
// language server published) and emits them like Explain does. Each one goes to the explainer
// of the language of its file; one without a file (a crash line, a missing main) goes to
// codeLanguage, the language of the last run, or to the default language when it is "".
func (s *AssistantService) ExplainDiagnostics(codeLanguage domain.CodeLanguage, diagnostics []domain.Diagnostic) ([]domain.ExplainedDiagnostic, error) {
	language := s.language.Current()
	items := []domain.ExplainedDiagnostic{}
	for _, group := range s.groupByLanguage(codeLanguage, diagnostics) {
		items = append(items, app.NewExplainError(group.support.Explainer).FromDiagnostics(group.diagnostics, language)...)
	}
	s.sink.Explained(items)
	return items, nil
}

type diagnosticGroup struct {
	support     app.LanguageSupport
	diagnostics []domain.Diagnostic
}

// groupByLanguage splits the diagnostics by language, in order of first appearance.
func (s *AssistantService) groupByLanguage(fallback domain.CodeLanguage, diagnostics []domain.Diagnostic) []*diagnosticGroup {
	var groups []*diagnosticGroup
	index := map[domain.CodeLanguage]*diagnosticGroup{}
	for _, diagnostic := range diagnostics {
		file := ""
		if diagnostic.Location != nil {
			file = diagnostic.Location.File
		}
		support := s.supportOfDiagnostic(file, fallback)
		group, ok := index[support.Profile.ID]
		if !ok {
			group = &diagnosticGroup{support: support}
			index[support.Profile.ID] = group
			groups = append(groups, group)
		}
		group.diagnostics = append(group.diagnostics, diagnostic)
	}
	return groups
}

// supportOfDiagnostic is the language of the file, or fallback for a diagnostic without one.
func (s *AssistantService) supportOfDiagnostic(file string, fallback domain.CodeLanguage) app.LanguageSupport {
	if file != "" || fallback == "" {
		return s.supportForFile(file)
	}
	if support, err := s.supportOf(fallback); err == nil {
		return support
	}
	return s.supportForFile(file)
}
