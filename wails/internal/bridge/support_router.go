package bridge

import (
	"fmt"

	"github.com/codeplai/VizcachaIDE/wails/internal/app"
	"github.com/codeplai/VizcachaIDE/wails/internal/domain"
)

// supportRouter is the repeated step of every service: find the language support of a file or
// of a language id, or fail with app.ErrUnknownCodeLanguage. It is embedded in the services.
// (LanguageResolver, in language_resolver.go, is something else: the UI language.)
type supportRouter struct {
	registry *app.LanguageRegistry
}

// supportFor finds the language of a file by its extension.
func (r supportRouter) supportFor(path string) (app.LanguageSupport, error) {
	support, ok := r.registry.ForPath(path)
	if !ok {
		return app.LanguageSupport{}, fmt.Errorf("%w: no language for %q", app.ErrUnknownCodeLanguage, path)
	}
	return support, nil
}

// supportOf finds a language by its id.
func (r supportRouter) supportOf(codeLanguage domain.CodeLanguage) (app.LanguageSupport, error) {
	support, ok := r.registry.ForID(codeLanguage)
	if !ok {
		return app.LanguageSupport{}, fmt.Errorf("%w: %q", app.ErrUnknownCodeLanguage, codeLanguage)
	}
	return support, nil
}

// supportForFile is supportFor for places where a missing or unknown file means "the default
// language" (a diagnostic without a location).
func (r supportRouter) supportForFile(path string) app.LanguageSupport {
	if support, ok := r.registry.ForPath(path); ok {
		return support
	}
	return r.registry.Default()
}
