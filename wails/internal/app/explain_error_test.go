package app

import (
	"testing"

	"github.com/codeplai/VizcachaIDE/wails/internal/domain"
)

type stubExplainer struct{}

func (stubExplainer) Parse(string, string) []domain.Diagnostic {
	return []domain.Diagnostic{
		{Severity: domain.SeverityError, Message: "a"},
		{Severity: domain.SeverityError, Message: "a"},
		{Severity: domain.SeverityWarning, Message: "b"},
		{Severity: domain.SeverityInfo, Message: "c"},
		{Severity: domain.SeverityHint, Message: "d"},
	}
}

func (stubExplainer) Explain(d domain.Diagnostic, _ string) *domain.ErrorExplanation {
	if d.Message == "b" {
		return nil
	}
	return &domain.ErrorExplanation{ExplanationID: d.Message}
}

func TestExplainErrorKeepsOnlyUniqueErrorsAndWarnings(t *testing.T) {
	items := NewExplainError(stubExplainer{}).FromOutput("", "", "en")
	if len(items) != 2 {
		t.Fatalf("items = %+v, want 2", items)
	}
	if items[0].Explanation == nil || items[0].Explanation.ExplanationID != "a" {
		t.Errorf("first = %+v", items[0])
	}
	if items[1].Explanation != nil || items[1].Diagnostic.Message != "b" {
		t.Errorf("unrecognised diagnostics keep a nil explanation: %+v", items[1])
	}
}

func TestExplainErrorSeparatesSameMessageInDifferentPlaces(t *testing.T) {
	first, second := domain.SourceLocation{File: "a.go", Line: 1}, domain.SourceLocation{File: "a.go", Line: 2}
	items := NewExplainError(stubExplainer{}).FromDiagnostics([]domain.Diagnostic{
		{Severity: domain.SeverityError, Message: "x", Location: &first},
		{Severity: domain.SeverityError, Message: "x", Location: &second},
		{Severity: domain.SeverityError, Message: "x", Location: &first},
	}, "es")
	if len(items) != 2 {
		t.Errorf("items = %d, want 2", len(items))
	}
}
