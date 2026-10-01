package main

import (
	"github.com/codeplai/VizcachaIDE/wails/internal/app"
	"github.com/codeplai/VizcachaIDE/wails/internal/domain"
	"github.com/codeplai/VizcachaIDE/wails/internal/i18n"
)

// settingsTranslator returns a function that translates an i18n key into the language
// chosen in the settings (English when it is "auto" or unknown). Adapters receive it so
// their user-facing messages follow the interface language.
func settingsTranslator(settings app.SettingsStore) func(key string) string {
	bundle, err := i18n.NewBundle()
	if err != nil {
		return nil
	}
	return func(key string) string {
		language := domain.LanguageEN
		if current, err := settings.Load(); err == nil && current.Language == domain.LanguageES {
			language = domain.LanguageES
		}
		return i18n.NewTranslator(bundle, language).T(key, nil)
	}
}

// firstBuildNotice returns the "preparing Go for the first time" text in the settings language.
func firstBuildNotice(settings app.SettingsStore) func() string {
	translate := settingsTranslator(settings)
	if translate == nil {
		return nil
	}
	return func() string { return translate("run.firstBuild") }
}
