package main

import (
	"fmt"

	"github.com/codeplai/VizcachaIDE/wails/internal/i18n"
)

// backendTexts translates the texts the backend itself produces (first-build notice,
// debugger messages, dialog titles) into the interface language. The language is read
// on every call: "auto" follows the operating system and a change in Settings applies
// at once.
type backendTexts struct {
	live *i18n.LiveTranslator
}

func newBackendTexts(currentLanguage func() string) (*backendTexts, error) {
	live, err := i18n.NewLiveTranslator(currentLanguage)
	if err != nil {
		return nil, fmt.Errorf("load backend texts: %w", err)
	}
	return &backendTexts{live: live}, nil
}

// text translates a key without placeholders (the shape adapters expect).
func (b *backendTexts) text(key string) string { return b.live.T(key, nil) }

// withData translates a key that has placeholders.
func (b *backendTexts) withData(key string, data map[string]any) string {
	return b.live.T(key, data)
}
