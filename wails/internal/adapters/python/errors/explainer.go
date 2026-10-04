// Package errors is the error explainer of Python: a parser of the output of python (tracebacks,
// syntax errors) and of ruff check (JSON), plus a catalog of explanations, run by the shared
// engine in internal/protocol/errorcatalog.
//
// The catalog is data: data/catalog.python.json is embedded in the binary and holds, for every
// stable id (PY-*, W-PY-*), its regular expressions and its EN/ES texts. The messages of Python
// are never translated; only the explanations are.
package errors

import (
	_ "embed" // the catalog is embedded data

	"github.com/codeplai/VizcachaIDE/wails/internal/protocol/errorcatalog"
)

//go:embed data/catalog.python.json
var catalogJSON []byte

// NewExplainer builds the app.ErrorExplainer of Python. It fails only if the catalog is inconsistent.
func NewExplainer() (*errorcatalog.Explainer, error) {
	return errorcatalog.NewExplainer(catalogJSON, ParseOutput)
}
