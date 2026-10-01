package errorcatalog

import (
	"strings"

	"github.com/codeplai/VizcachaIDE/wails/internal/app"
	"github.com/codeplai/VizcachaIDE/wails/internal/domain"
)

// Explainer implements app.ErrorExplainer on top of the embedded catalog.
type Explainer struct {
	catalog *catalog
	parser  parser
}

var _ app.ErrorExplainer = (*Explainer)(nil)

// NewExplainer loads the embedded catalog. It fails only if data/catalog.json is inconsistent.
func NewExplainer() (*Explainer, error) {
	loaded, err := loadCatalog()
	if err != nil {
		return nil, err
	}
	identify := func(message string) string {
		if found := loaded.find(message); found != nil {
			return found.entry.id
		}
		return ""
	}
	return &Explainer{catalog: loaded, parser: parser{identify: identify}}, nil
}

// Parse extracts diagnostics (compiler errors, vet warnings, panics) from raw output.
func (e *Explainer) Parse(rawOutput, workingDir string) []domain.Diagnostic {
	return e.parser.parse(rawOutput, workingDir)
}

// Explain returns the explanation in "en" or "es" (anything else is English),
// or nil when the message is not in the catalog.
func (e *Explainer) Explain(diagnostic domain.Diagnostic, language string) *domain.ErrorExplanation {
	found := e.catalog.find(diagnostic.Message)
	if found == nil {
		return nil
	}
	chosen := found.entry.texts["en"]
	if language == domain.LanguageES {
		chosen = found.entry.texts["es"]
	}
	return &domain.ErrorExplanation{
		ExplanationID: found.entry.id,
		Title:         fill(chosen.Title, found.placeholders),
		Body:          fill(chosen.Body, found.placeholders),
		FixHint:       fill(chosen.Fix, found.placeholders),
		Placeholders:  found.placeholders,
	}
}

// fill replaces every {name} with its value in a single pass; other braces stay as they are.
func fill(text string, placeholders map[string]string) string {
	pairs := make([]string, 0, 2*len(placeholders))
	for name, value := range placeholders {
		pairs = append(pairs, "{"+name+"}", value)
	}
	return strings.NewReplacer(pairs...).Replace(text)
}
