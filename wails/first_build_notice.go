package main

import (
	"github.com/codeplai/VizcachaIDE/wails/internal/app"
	"github.com/codeplai/VizcachaIDE/wails/internal/domain"
	"github.com/codeplai/VizcachaIDE/wails/internal/i18n"
)

// firstBuildNotice returns the "preparing Go for the first time" text in the
// language chosen in the settings (English when it is "auto" or unknown).
func firstBuildNotice(settings app.SettingsStore) func() string {
	bundle, err := i18n.NewBundle()
	if err != nil {
		return nil
	}
	return func() string {
		language := domain.LanguageEN
		if current, err := settings.Load(); err == nil && current.Language == domain.LanguageES {
			language = domain.LanguageES
		}
		return i18n.NewTranslator(bundle, language).T("run.firstBuild", nil)
	}
}
