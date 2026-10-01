package i18n

import "testing"

func TestSystemLanguageIsSupported(t *testing.T) {
	if got := SystemLanguage(); got != "en" && got != "es" {
		t.Errorf("SystemLanguage() = %q", got)
	}
}
