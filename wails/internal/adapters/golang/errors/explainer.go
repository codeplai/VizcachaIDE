// Package errors is the error explainer of Go: a parser of the output of go build, go run and
// go vet (and of panics) plus a catalog of 25 explanations, run by the shared engine in
// internal/protocol/errorcatalog.
//
// The catalog is data: data/catalog.go.json is embedded in the binary and holds, for every
// stable id (E-*, P-*, V-*), its regular expressions and its EN/ES texts. The texts live here
// and not in internal/i18n because go-i18n parses "{{" as a template action and the fix hints
// contain Go code with braces ("func main() { ... }").
package errors

import (
	_ "embed" // the catalog is embedded data

	"github.com/codeplai/VizcachaIDE/wails/internal/protocol/errorcatalog"
)

//go:embed data/catalog.go.json
var catalogJSON []byte

// NewExplainer builds the app.ErrorExplainer of Go. It fails only if the catalog is inconsistent.
func NewExplainer() (*errorcatalog.Explainer, error) {
	return errorcatalog.NewExplainer(catalogJSON, ParseOutput)
}
