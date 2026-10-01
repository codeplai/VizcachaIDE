package i18n

import (
	locale "github.com/jeandeaual/go-locale"

	"github.com/codeplai/VizcachaIDE/wails/internal/app"
)

// SystemLocale returns the locale tag of the operating system ("es-PE", "en-US"...),
// or "" when it cannot be detected. It needs no cgo on Windows, macOS or Linux.
func SystemLocale() string {
	tag, err := locale.GetLocale()
	if err != nil {
		return ""
	}
	return tag
}

// SystemLanguage returns the language of the operating system: "es" for any Spanish
// locale and "en" for everything else (including when it cannot be detected).
func SystemLanguage() string { return app.LanguageOfLocale(SystemLocale()) }
