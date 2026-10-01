package i18n

import (
	"encoding/json"
	"slices"
	"testing"
)

func TestCatalogsHaveTheSameKeys(t *testing.T) {
	keys := map[string][]string{}
	for _, lang := range Supported {
		raw, err := localeFiles.ReadFile("locales/" + lang + ".json")
		if err != nil {
			t.Fatal(err)
		}
		var catalog map[string]json.RawMessage
		if err := json.Unmarshal(raw, &catalog); err != nil {
			t.Fatal(err)
		}
		for key := range catalog {
			keys[lang] = append(keys[lang], key)
		}
		slices.Sort(keys[lang])
	}
	if !slices.Equal(keys["en"], keys["es"]) {
		t.Error("en.json and es.json have different keys")
	}
	if len(keys["en"]) == 0 {
		t.Error("the catalogs are empty: run go run ./tools/po2json")
	}
}

func TestTranslatesWithPlaceholders(t *testing.T) {
	bundle, err := NewBundle()
	if err != nil {
		t.Fatal(err)
	}
	es := NewTranslator(bundle, "es")
	got := es.T("run.starting", map[string]any{"file": "main.go"})
	if got != "▶ Ejecutando main.go…" {
		t.Errorf("es run.starting = %q", got)
	}
	en := NewTranslator(bundle, "en")
	if got := en.T("actions.run", nil); got != "Run" {
		t.Errorf("en actions.run = %q", got)
	}
}

func TestPlurals(t *testing.T) {
	bundle, err := NewBundle()
	if err != nil {
		t.Fatal(err)
	}
	en := NewTranslator(bundle, "en")
	one := en.Plural("run.compileFailed", 1, nil)
	many := en.Plural("run.compileFailed", 3, nil)
	if one != "✗ Your program didn't run: Go found 1 problem. See the explanation on the right." {
		t.Errorf("one = %q", one)
	}
	if many != "✗ Your program didn't run: Go found 3 problems. See the explanation on the right." {
		t.Errorf("many = %q", many)
	}
}

func TestMissingIDReturnsTheID(t *testing.T) {
	bundle, err := NewBundle()
	if err != nil {
		t.Fatal(err)
	}
	if got := NewTranslator(bundle, "es").T("does.not.exist", nil); got != "does.not.exist" {
		t.Errorf("got %q", got)
	}
}
