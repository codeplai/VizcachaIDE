package errors

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/codeplai/VizcachaIDE/wails/internal/domain"
	"github.com/codeplai/VizcachaIDE/wails/internal/protocol/errorcatalog"
)

func newExplainer(t *testing.T) *errorcatalog.Explainer {
	t.Helper()
	explainer, err := NewExplainer()
	if err != nil {
		t.Fatal(err)
	}
	return explainer
}

func catalogIDs(t *testing.T) []string {
	t.Helper()
	var entries []struct {
		ID string `json:"id"`
	}
	if err := json.Unmarshal(catalogJSON, &entries); err != nil {
		t.Fatal(err)
	}
	ids := make([]string, 0, len(entries))
	for _, item := range entries {
		ids = append(ids, item.ID)
	}
	return ids
}

func TestCatalogHasTwentyFiveStableIDs(t *testing.T) {
	ids := catalogIDs(t)
	if len(ids) != 25 {
		t.Fatalf("catalog has %d entries, want 25", len(ids))
	}
	for _, id := range ids {
		if !strings.HasPrefix(id, "E-") && !strings.HasPrefix(id, "P-") && !strings.HasPrefix(id, "V-") {
			t.Errorf("unexpected id %q", id)
		}
	}
}

// Every real Go output recorded by the 1.0 produces exactly its id (25 of 25).
func TestFixturesProduceTheirID(t *testing.T) {
	explainer := newExplainer(t)
	workDir := t.TempDir()
	for _, id := range catalogIDs(t) {
		raw, err := os.ReadFile(filepath.Join("testdata", "go_output", id+".txt"))
		if err != nil {
			t.Fatal(err)
		}
		text := strings.ReplaceAll(string(raw), "$WORKDIR", filepath.ToSlash(workDir))
		diagnostics := explainer.Parse(text, workDir)
		if len(diagnostics) != 1 || diagnostics[0].Code != id {
			t.Errorf("%s: diagnostics = %+v", id, diagnostics)
			continue
		}
		explanation := explainer.Explain(diagnostics[0], "en")
		if explanation == nil || explanation.ExplanationID != id {
			t.Errorf("%s: explanation = %+v", id, explanation)
		}
		if location := diagnostics[0].Location; location != nil && filepath.Base(location.File) != id+".go" {
			t.Errorf("%s: location = %+v", id, location)
		}
	}
}

func TestPlaceholdersAreFilledInEnglishAndSpanish(t *testing.T) {
	explainer := newExplainer(t)
	diagnostic := domain.Diagnostic{Severity: domain.SeverityError, Message: "index out of range [7] with length 2"}

	en := explainer.Explain(diagnostic, "en")
	if en.Title != "Index 7 is out of range" || !strings.Contains(en.Body, "only has 2 element(s)") {
		t.Errorf("en = %+v", en)
	}
	es := explainer.Explain(diagnostic, "es")
	if !strings.Contains(es.Title, "7") || !strings.Contains(es.Body, "2") || strings.Contains(es.Body, "{") {
		t.Errorf("es = %+v", es)
	}
	if en.Placeholders["index"] != "7" || en.Placeholders["length"] != "2" {
		t.Errorf("placeholders = %v", en.Placeholders)
	}
}

func TestLiteralBracesAreKept(t *testing.T) {
	explainer := newExplainer(t)
	diagnostic := domain.Diagnostic{Message: "function main is undeclared in the main package"}
	if got := explainer.Explain(diagnostic, "en").FixHint; !strings.Contains(got, "func main() { ... }") {
		t.Errorf("fix = %q", got)
	}
}

func TestUnknownLanguageFallsBackToEnglish(t *testing.T) {
	explainer := newExplainer(t)
	diagnostic := domain.Diagnostic{Message: "declared and not used: n"}
	if got := explainer.Explain(diagnostic, "fr").Title; got != `Variable "n" is never used` {
		t.Errorf("title = %q", got)
	}
}

func TestExplainUsesTheMessageOfLanguageServerDiagnostics(t *testing.T) {
	explainer := newExplainer(t)
	diagnostic := domain.Diagnostic{Severity: domain.SeverityError, Message: "declared and not used: n", Source: "compiler"}
	explanation := explainer.Explain(diagnostic, "es")
	if explanation.ExplanationID != "E-UNUSED-VAR" || explanation.Placeholders["name"] != "n" {
		t.Errorf("explanation = %+v", explanation)
	}
}

func TestUnrecognisedMessageHasNoExplanation(t *testing.T) {
	explainer := newExplainer(t)
	if got := explainer.Explain(domain.Diagnostic{Message: "something Go never says"}, "en"); got != nil {
		t.Errorf("explanation = %+v, want nil", got)
	}
}
