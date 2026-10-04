package errorcatalog_test

import (
	"errors"
	"testing"

	"github.com/codeplai/VizcachaIDE/wails/internal/domain"
	"github.com/codeplai/VizcachaIDE/wails/internal/protocol/errorcatalog"
)

// Rust's catalog keys on rustc's stable codes; a pattern, when it matches, fills placeholders.
const codedCatalog = `[
 {"id": "RS-MOVED", "codes": ["E0382"], "patterns": ["of moved value: ` + "`" + `(?P<name>[^` + "`" + `]+)` + "`" + `"],
  "en": {"title": "{name} was moved", "body": "b", "fix": "f"},
  "es": {"title": "{name} se movió", "body": "b", "fix": "f"}},
 {"id": "W-RS-UNUSED-VAR", "codes": ["unused_variables"],
  "en": {"title": "Unused variable", "body": "b", "fix": "f"},
  "es": {"title": "Variable sin usar", "body": "b", "fix": "f"}}
]`

func TestEntriesAreRecognisedByTheirStableCode(t *testing.T) {
	explainer, err := errorcatalog.NewExplainer([]byte(codedCatalog), sampleParser)
	if err != nil {
		t.Fatal(err)
	}
	moved := explainer.Explain(domain.Diagnostic{Code: "E0382", Message: "borrow of moved value: `s`"}, "es")
	if moved == nil || moved.ExplanationID != "RS-MOVED" || moved.Title != "s se movió" {
		t.Errorf("moved = %+v", moved)
	}
	unused := explainer.Explain(domain.Diagnostic{Code: "unused_variables", Message: "unused variable: `x`"}, "en")
	if unused == nil || unused.ExplanationID != "W-RS-UNUSED-VAR" {
		t.Errorf("unused = %+v", unused)
	}
	if other := explainer.Explain(domain.Diagnostic{Code: "E9999", Message: "something else"}, "en"); other != nil {
		t.Errorf("other = %+v", other)
	}
}

func TestAnEntryWithoutPatternsCannotUsePlaceholders(t *testing.T) {
	bad := `[{"id": "X", "codes": ["E1"], "en": {"title": "{name}", "body": "b", "fix": "f"}, "es": {"title": "t", "body": "b", "fix": "f"}}]`
	if _, err := errorcatalog.NewExplainer([]byte(bad), sampleParser); !errors.Is(err, errorcatalog.ErrInvalidCatalog) {
		t.Errorf("err = %v", err)
	}
}
