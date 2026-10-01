package app

import (
	"fmt"

	"github.com/codeplai/VizcachaIDE/wails/internal/domain"
)

// ExplainError is the use case "explain the problems found in Go output or reported
// by a language server". Hints and infos are skipped; duplicates are dropped.
type ExplainError struct {
	explainer ErrorExplainer
}

// NewExplainError creates the use case.
func NewExplainError(explainer ErrorExplainer) *ExplainError {
	return &ExplainError{explainer: explainer}
}

// FromOutput parses raw go build / go run / go vet output and explains each problem.
func (u *ExplainError) FromOutput(rawOutput, workingDir, language string) []domain.ExplainedDiagnostic {
	return u.FromDiagnostics(u.explainer.Parse(rawOutput, workingDir), language)
}

// FromDiagnostics explains errors and warnings. Problems the catalog does not know keep
// a nil Explanation: the frontend shows Go's own text and the search action.
func (u *ExplainError) FromDiagnostics(diagnostics []domain.Diagnostic, language string) []domain.ExplainedDiagnostic {
	seen := map[string]bool{}
	items := []domain.ExplainedDiagnostic{}
	for _, diagnostic := range diagnostics {
		if !isProblem(diagnostic) {
			continue
		}
		key := diagnosticKey(diagnostic)
		if seen[key] {
			continue
		}
		seen[key] = true
		items = append(items, domain.ExplainedDiagnostic{
			Diagnostic:  diagnostic,
			Explanation: u.explainer.Explain(diagnostic, language),
		})
	}
	return items
}

func isProblem(diagnostic domain.Diagnostic) bool {
	return diagnostic.Severity == domain.SeverityError || diagnostic.Severity == domain.SeverityWarning
}

// diagnosticKey identifies a diagnostic by value (Location and End are pointers).
func diagnosticKey(d domain.Diagnostic) string {
	return fmt.Sprintf("%s|%s|%s|%s|%s|%s|%s",
		locationKey(d.Location), locationKey(d.End), d.Severity, d.Message, d.RawText, d.Source, d.Code)
}

func locationKey(location *domain.SourceLocation) string {
	if location == nil {
		return "-"
	}
	return fmt.Sprintf("%s:%d:%d", location.File, location.Line, location.Column)
}
