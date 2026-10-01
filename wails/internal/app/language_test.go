package app

import "testing"

func TestLanguageOfLocale(t *testing.T) {
	cases := map[string]string{
		"es": "es", "es-PE": "es", "es_ES.UTF-8": "es", "ES-mx": "es", " es-419 ": "es",
		"en-US": "en", "en": "en", "fr-FR": "en", "": "en", "esperanto": "en", "C": "en",
	}
	for tag, want := range cases {
		if got := LanguageOfLocale(tag); got != want {
			t.Errorf("LanguageOfLocale(%q) = %q, want %q", tag, got, want)
		}
	}
}

func TestResolveLanguage(t *testing.T) {
	spanish := func() string { return "es-PE" }
	english := func() string { return "en-US" }
	cases := []struct {
		setting string
		system  func() string
		want    string
	}{
		{"es", english, "es"}, {"en", spanish, "en"}, {"auto", spanish, "es"},
		{"auto", english, "en"}, {"auto", nil, "en"}, {"zz", spanish, "en"}, {"", spanish, "en"},
	}
	for _, c := range cases {
		if got := ResolveLanguage(c.setting, c.system); got != c.want {
			t.Errorf("ResolveLanguage(%q) = %q, want %q", c.setting, got, c.want)
		}
	}
}
