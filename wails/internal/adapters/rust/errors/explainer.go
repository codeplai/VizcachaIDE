// Package errors is the error explainer of Rust: a parser of the output of rustc, cargo and
// clippy (JSON diagnostics or human text) and of the stderr of the program (panics, stack
// overflows, crashes, an Err returned by main), plus a catalog of explanations, run by the
// shared engine in internal/protocol/errorcatalog.
//
// The catalog is data: data/catalog.rust.json is embedded in the binary and holds, for every
// stable id (RS-*, W-RS-*), the error codes of rustc (E0382...) or the regular expressions that
// recognise it and its EN/ES texts. The messages of the tools are never translated; only the
// explanations are.
package errors

import (
	_ "embed" // the catalog is embedded data
	"regexp"
	"strings"

	"github.com/codeplai/VizcachaIDE/wails/internal/app"
	"github.com/codeplai/VizcachaIDE/wails/internal/domain"
	"github.com/codeplai/VizcachaIDE/wails/internal/protocol/errorcatalog"
)

//go:embed data/catalog.rust.json
var catalogJSON []byte

// unfilled is a placeholder the message did not give a value for (an entry recognised by its
// error code only, when the message has another wording).
var unfilled = regexp.MustCompile(`\{\w+\}`)

// Explainer is the catalog engine plus the reading of the diagnostics rust-analyzer publishes.
type Explainer struct{ catalog *errorcatalog.Explainer }

var _ app.ErrorExplainer = (*Explainer)(nil)

// NewExplainer builds the app.ErrorExplainer of Rust. It fails only if the catalog is inconsistent.
func NewExplainer() (*Explainer, error) {
	catalog, err := errorcatalog.NewExplainer(catalogJSON, ParseOutput)
	if err != nil {
		return nil, err
	}
	return &Explainer{catalog: catalog}, nil
}

// Parse reads the output of rustc, cargo, clippy or the program.
func (e *Explainer) Parse(rawOutput, workingDir string) []domain.Diagnostic {
	return e.catalog.Parse(rawOutput, workingDir)
}

// Explain explains a diagnostic of the compiler, clippy or rust-analyzer. rust-analyzer
// publishes rustc's diagnostics with their code ("E0382") and message, so the code alone
// recognises them. The name of a clippy lint is the Code, not part of the message: it goes in
// front of the message so the generic entry of clippy can say it.
func (e *Explainer) Explain(diagnostic domain.Diagnostic, language string) *domain.ErrorExplanation {
	if strings.HasPrefix(diagnostic.Code, "clippy::") && !strings.HasPrefix(diagnostic.Message, diagnostic.Code+": ") {
		diagnostic.Message = diagnostic.Code + ": " + diagnostic.Message
	}
	explanation := e.catalog.Explain(diagnostic, language)
	if explanation == nil {
		return nil
	}
	explanation.Title = unfilled.ReplaceAllString(explanation.Title, "…")
	explanation.Body = unfilled.ReplaceAllString(explanation.Body, "…")
	explanation.FixHint = unfilled.ReplaceAllString(explanation.FixHint, "…")
	return explanation
}
