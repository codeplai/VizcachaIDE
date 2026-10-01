package bridge

import (
	"github.com/codeplai/VizcachaIDE/wails/internal/app"
	"github.com/codeplai/VizcachaIDE/wails/internal/domain"
)

// LanguageResolver says which language ("en" or "es") the backend speaks right now: the
// one chosen in the settings, or the operating system's when the setting is "auto".
type LanguageResolver struct {
	store  app.SettingsStore
	system func() string
}

// NewLanguageResolver creates the resolver. system returns the locale tag of the
// operating system; nil means that "auto" resolves to English.
func NewLanguageResolver(store app.SettingsStore, system func() string) *LanguageResolver {
	return &LanguageResolver{store: store, system: system}
}

// Current reads the settings every time, so a change applies without a restart.
func (r *LanguageResolver) Current() string {
	settings, err := r.store.Load()
	if err != nil {
		return domain.LanguageEN
	}
	return app.ResolveLanguage(settings.Language, r.system)
}
