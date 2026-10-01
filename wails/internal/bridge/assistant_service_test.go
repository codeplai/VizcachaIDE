package bridge

import (
	"testing"

	"github.com/codeplai/VizcachaIDE/wails/internal/domain"
)

// fakeExplainer parses every line as a diagnostic and explains it as "<language>:<message>".
type fakeExplainer struct{}

func (fakeExplainer) Parse(rawOutput, _ string) []domain.Diagnostic {
	return []domain.Diagnostic{
		{Severity: domain.SeverityError, Message: rawOutput},
		{Severity: domain.SeverityError, Message: rawOutput},
		{Severity: domain.SeverityHint, Message: "hint"},
	}
}

func (fakeExplainer) Explain(d domain.Diagnostic, language string) *domain.ErrorExplanation {
	return &domain.ErrorExplanation{ExplanationID: "X", Title: language + ":" + d.Message}
}

type explainedSink struct {
	recordingSink
	items []domain.ExplainedDiagnostic
}

func (s *explainedSink) Explained(items []domain.ExplainedDiagnostic) { s.items = items }

func TestAssistantExplainsInTheSettingsLanguage(t *testing.T) {
	store := NewMemorySettingsStore()
	if err := store.Save(domain.Settings{Language: domain.LanguageES}); err != nil {
		t.Fatal(err)
	}
	sink := &explainedSink{}
	items, err := NewAssistantService(sink, fakeExplainer{}, NewLanguageResolver(store, nil)).Explain("boom", "")
	if err != nil {
		t.Fatal(err)
	}
	if len(items) != 1 || items[0].Explanation.Title != "es:boom" {
		t.Fatalf("items = %+v, want one deduplicated Spanish explanation", items)
	}
	if len(sink.items) != 1 {
		t.Errorf("assistant:explained was not emitted with the items: %+v", sink.items)
	}
}

func TestAssistantExplainsParsedDiagnostics(t *testing.T) {
	sink := &explainedSink{}
	store := NewMemorySettingsStore()
	if err := store.Save(domain.Settings{Language: domain.LanguageEN}); err != nil {
		t.Fatal(err)
	}
	service := NewAssistantService(sink, fakeExplainer{}, NewLanguageResolver(store, nil))
	items, err := service.ExplainDiagnostics([]domain.Diagnostic{
		{Severity: domain.SeverityWarning, Message: "w"},
		{Severity: domain.SeverityInfo, Message: "i"},
	})
	if err != nil {
		t.Fatal(err)
	}
	if len(items) != 1 || items[0].Explanation.Title != "en:w" || len(sink.items) != 1 {
		t.Errorf("items = %+v", items)
	}
}
