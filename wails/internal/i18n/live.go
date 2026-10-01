package i18n

import (
	goi18n "github.com/nicksnyder/go-i18n/v2/i18n"
)

// LiveTranslator translates into the language that is current at the moment of each call,
// so changing the language in the settings needs no restart.
type LiveTranslator struct {
	bundle   *goi18n.Bundle
	language func() string
}

// NewLiveTranslator creates a translator; language returns "en" or "es" on every call.
func NewLiveTranslator(language func() string) (*LiveTranslator, error) {
	bundle, err := NewBundle()
	if err != nil {
		return nil, err
	}
	return &LiveTranslator{bundle: bundle, language: language}, nil
}

// T returns the text of id in the current language with its placeholders filled in.
func (l *LiveTranslator) T(id string, data map[string]any) string {
	return NewTranslator(l.bundle, l.language()).T(id, data)
}
