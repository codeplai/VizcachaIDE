// Package errors is the error explainer of C++: a parser of the output of g++ and clang++
// (compiler diagnostics, linker errors, crashes and uncaught exceptions of the program) plus a
// catalog of explanations, run by the shared engine in internal/protocol/errorcatalog.
//
// The catalog is data: data/catalog.cpp.json is embedded in the binary and holds, for every
// stable id (CPP-*), its regular expressions for the GNU and the LLVM family and its EN/ES
// texts. The messages of the compilers are never translated; only the explanations are.
package errors

import (
	_ "embed" // the catalog is embedded data

	"github.com/codeplai/VizcachaIDE/wails/internal/protocol/errorcatalog"
)

//go:embed data/catalog.cpp.json
var catalogJSON []byte

// NewExplainer builds the app.ErrorExplainer of C++. It fails only if the catalog is inconsistent.
func NewExplainer() (*errorcatalog.Explainer, error) {
	return errorcatalog.NewExplainer(catalogJSON, ParseOutput)
}
