package bridge

import (
	"github.com/codeplai/VizcachaIDE/wails/internal/app"
	"github.com/codeplai/VizcachaIDE/wails/internal/domain"
)

// AssistantService explains Go errors to beginners. It emits assistant:explained with
// texts already translated to the language of the settings.
type AssistantService struct {
	sink     app.EventSink
	explain  *app.ExplainError
	settings app.SettingsStore
}

// NewAssistantService creates the service.
func NewAssistantService(sink app.EventSink, explainer app.ErrorExplainer, settings app.SettingsStore) *AssistantService {
	return &AssistantService{sink: sink, explain: app.NewExplainError(explainer), settings: settings}
}

// Explain parses raw go build / go run / go vet output (or a panic) and emits the
// explained diagnostics. Only errors and warnings are included, without duplicates.
func (s *AssistantService) Explain(rawOutput, workingDir string) ([]domain.ExplainedDiagnostic, error) {
	items := s.explain.FromOutput(rawOutput, workingDir, s.language())
	s.sink.Explained(items)
	return items, nil
}

// ExplainDiagnostics explains diagnostics that are already parsed (for example the
// ones the language server published) and emits them like Explain does.
func (s *AssistantService) ExplainDiagnostics(diagnostics []domain.Diagnostic) ([]domain.ExplainedDiagnostic, error) {
	items := s.explain.FromDiagnostics(diagnostics, s.language())
	s.sink.Explained(items)
	return items, nil
}

// language is "es" when the user chose Spanish and "en" otherwise.
func (s *AssistantService) language() string {
	current, err := s.settings.Load()
	if err != nil || current.Language != domain.LanguageES {
		return domain.LanguageEN
	}
	return domain.LanguageES
}
