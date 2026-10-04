package errorcatalog

import (
	"strings"

	"github.com/codeplai/VizcachaIDE/wails/internal/app"
	"github.com/codeplai/VizcachaIDE/wails/internal/domain"
)

// OutputParser turns raw tool output into diagnostics. identify returns the catalog id of a
// message ("" when the catalog does not know it).
type OutputParser func(rawOutput, workingDir string, identify func(message string) string) []domain.Diagnostic

// Explainer implements app.ErrorExplainer on top of the catalog of one language and its parser.
type Explainer struct {
	catalog *catalog
	parse   OutputParser
}

var _ app.ErrorExplainer = (*Explainer)(nil)

// NewExplainer loads the catalog the language embeds. It fails only if the JSON is inconsistent.
func NewExplainer(catalogJSON []byte, parse OutputParser) (*Explainer, error) {
	loaded, err := loadCatalog(catalogJSON)
	if err != nil {
		return nil, err
	}
	return &Explainer{catalog: loaded, parse: parse}, nil
}

// identify returns the catalog id of a message, or "".
func (e *Explainer) identify(message string) string {
	if found := e.catalog.find(message); found != nil {
		return found.entry.id
	}
	return ""
}

// Parse extracts diagnostics (compiler errors, warnings, panics) from raw output.
func (e *Explainer) Parse(rawOutput, workingDir string) []domain.Diagnostic {
	return e.parse(rawOutput, workingDir, e.identify)
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
