package bridge

import (
	"testing"

	"github.com/codeplai/VizcachaIDE/wails/internal/domain"
)

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
	items, err := NewAssistantService(sink, newTestRegistry(t), NewLanguageResolver(store, nil)).Explain("go", "boom", "")
	if err != nil {
		t.Fatal(err)
	}
	if len(items) != 1 || items[0].Explanation.Title != "go:es:boom" {
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
	service := NewAssistantService(sink, newTestRegistry(t), NewLanguageResolver(store, nil))
	items, err := service.ExplainDiagnostics([]domain.Diagnostic{
		{Severity: domain.SeverityWarning, Message: "w"},
		{Severity: domain.SeverityInfo, Message: "i"},
	})
	if err != nil {
		t.Fatal(err)
	}
	if len(items) != 1 || items[0].Explanation.Title != "go:en:w" || len(sink.items) != 1 {
		t.Errorf("items = %+v", items)
	}
}
