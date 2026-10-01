package i18n

import "testing"

// The errors of the IDE itself (docs/wails/UX_COPY.md, "Errores del IDE") must exist in both languages.
var ideErrorKeys = map[string]map[string]any{
	"errors.goNotFound":        nil,
	"errors.goNotFoundInstall": nil,
	"errors.goNotFoundChoose":  nil,
	"errors.delveNotFound":     nil,
	"errors.delveCopyCommand":  nil,
	"errors.goplsNotFound":     nil,
	"errors.saveFailed":        {"file": "main.go", "reason": "Disk full."},
	"errors.saveRetry":         nil,
	"errors.formatRejected":    {"line": 7},
	"errors.invalidArgs":       nil,
}

func TestIDEErrorMessagesExistInBothLanguages(t *testing.T) {
	bundle, err := NewBundle()
	if err != nil {
		t.Fatal(err)
	}
	for _, lang := range Supported {
		translator := NewTranslator(bundle, lang)
		for key, data := range ideErrorKeys {
			if got := translator.T(key, data); got == key || got == "" {
				t.Errorf("%s: %s is missing", lang, key)
			}
		}
	}
}

func TestIDEErrorPlaceholdersAreFilled(t *testing.T) {
	bundle, err := NewBundle()
	if err != nil {
		t.Fatal(err)
	}
	got := NewTranslator(bundle, "es").T("errors.saveFailed", ideErrorKeys["errors.saveFailed"])
	want := "No se pudo guardar main.go. Disk full. Revisa que la carpeta exista y que puedas escribir en ella."
	if got != want {
		t.Errorf("got %q", got)
	}
}
