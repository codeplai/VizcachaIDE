package bridge

import (
	"github.com/codeplai/VizcachaIDE/wails/internal/app"
	"github.com/codeplai/VizcachaIDE/wails/internal/domain"
)

// AssistantService explains Go errors to beginners. W0 STUB: owned by track G4,
// which replaces the body with app.ErrorExplainer. Emits assistant:explained.
type AssistantService struct {
	sink     app.EventSink
	settings app.SettingsStore
}

// NewAssistantService creates the service.
func NewAssistantService(sink app.EventSink, settings app.SettingsStore) *AssistantService {
	return &AssistantService{sink: sink, settings: settings}
}

// Explain parses raw Go output and emits the explained diagnostics.
func (s *AssistantService) Explain(rawOutput, workingDir string) ([]domain.ExplainedDiagnostic, error) {
	language := domain.LanguageEN
	if current, err := s.settings.Load(); err == nil && current.Language == domain.LanguageES {
		language = domain.LanguageES
	}
	items := []domain.ExplainedDiagnostic{
		{Diagnostic: sampleDiagnostic(), Explanation: sampleExplanation(language)},
	}
	s.sink.Explained(items)
	return items, nil
}
