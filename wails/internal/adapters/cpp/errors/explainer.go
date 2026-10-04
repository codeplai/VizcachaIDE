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
	"strings"
	"unicode"
	"unicode/utf8"

	"github.com/codeplai/VizcachaIDE/wails/internal/app"
	"github.com/codeplai/VizcachaIDE/wails/internal/domain"
	"github.com/codeplai/VizcachaIDE/wails/internal/protocol/errorcatalog"
)

// clangdSource is the Source of the diagnostics clangd publishes (adapters/cpp/clangd).
const clangdSource = "clangd"

//go:embed data/catalog.cpp.json
var catalogJSON []byte

// Explainer is the catalog engine plus the reading of clangd's live messages.
type Explainer struct{ catalog *errorcatalog.Explainer }

var _ app.ErrorExplainer = (*Explainer)(nil)

// NewExplainer builds the app.ErrorExplainer of C++. It fails only if the catalog is inconsistent.
func NewExplainer() (*Explainer, error) {
	catalog, err := errorcatalog.NewExplainer(catalogJSON, ParseOutput)
	if err != nil {
		return nil, err
	}
	return &Explainer{catalog: catalog}, nil
}

// Parse reads the output of the compiler, the linker or the program.
func (e *Explainer) Parse(rawOutput, workingDir string) []domain.Diagnostic {
	return e.catalog.Parse(rawOutput, workingDir)
}

// Explain explains a diagnostic of the compiler or of clangd. clangd writes clang's message
// with a capital letter and may append " (fix available)"; the catalog is written for the
// compiler's own form, so that is restored first.
func (e *Explainer) Explain(diagnostic domain.Diagnostic, language string) *domain.ErrorExplanation {
	if diagnostic.Source == clangdSource {
		diagnostic.Message = compilerForm(diagnostic.Message)
	}
	return e.catalog.Explain(diagnostic, language)
}

func compilerForm(message string) string {
	message = strings.TrimSuffix(message, " (fix available)")
	first, size := utf8.DecodeRuneInString(message)
	if size == 0 || !unicode.IsUpper(first) {
		return message
	}
	return string(unicode.ToLower(first)) + message[size:]
}
