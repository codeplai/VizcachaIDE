package app

import (
	"strings"

	"github.com/codeplai/VizcachaIDE/wails/internal/domain"
)

// LanguageOfLocale maps a locale tag such as "es-PE", "es_ES.UTF-8" or "en-US" to
// domain.LanguageES for any Spanish locale and domain.LanguageEN for everything else.
func LanguageOfLocale(tag string) string {
	lower := strings.ToLower(strings.TrimSpace(tag))
	if lower == "es" || strings.HasPrefix(lower, "es-") || strings.HasPrefix(lower, "es_") {
		return domain.LanguageES
	}
	return domain.LanguageEN
}

// ResolveLanguage turns the language setting ("auto", "en", "es" or anything else) into
// "en" or "es". "auto" asks system for the locale of the operating system, and only then.
func ResolveLanguage(setting string, system func() string) string {
	switch setting {
	case domain.LanguageES:
		return domain.LanguageES
	case domain.LanguageAuto:
		if system != nil {
			return LanguageOfLocale(system())
		}
	}
	return domain.LanguageEN
}
