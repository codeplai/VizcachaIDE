package main

import (
	"reflect"
	"strings"
	"testing"
)

const samplePO = `msgid ""
msgstr ""
"Language: es\n"

#, python-brace-format
msgid ""
"Hello {name}\n"
"again"
msgstr ""
"Hola {name}\n"
"otra vez"

msgid "{count} match"
msgid_plural "{count} matches"
msgstr[0] "{count} coincidencia"
msgstr[1] "{count} coincidencias"
`

func TestParsePO(t *testing.T) {
	entries, err := parsePO(strings.NewReader(samplePO))
	if err != nil {
		t.Fatal(err)
	}
	if len(entries) != 2 {
		t.Fatalf("got %d entries, want 2 (header skipped)", len(entries))
	}
	if entries[0].ID != "Hello {name}\nagain" || entries[0].Strs[0] != "Hola {name}\notra vez" {
		t.Errorf("entry 0 = %+v", entries[0])
	}
	if entries[1].Plural != "{count} matches" || len(entries[1].Strs) != 2 {
		t.Errorf("entry 1 = %+v", entries[1])
	}
}

func TestPluralBecomesICUAndGoForms(t *testing.T) {
	entries, err := parsePO(strings.NewReader(samplePO))
	if err != nil {
		t.Fatal(err)
	}
	icu := poText(entries[1])
	want := "{count, plural, one {{count} coincidencia} other {{count} coincidencias}}"
	if icu != want {
		t.Fatalf("icu = %q, want %q", icu, want)
	}
	forms, ok := goTemplate(icu).(map[string]string)
	if !ok || forms["one"] != "{{.count}} coincidencia" || forms["other"] != "{{.count}} coincidencias" {
		t.Errorf("go forms = %#v", goTemplate(icu))
	}
}

func TestGoTemplateWithHashPlural(t *testing.T) {
	icu := "Go found {count, plural, one {# problem} other {# problems}}. Done {file}."
	want := map[string]string{
		"one":   "Go found {{.count}} problem. Done {{.file}}.",
		"other": "Go found {{.count}} problems. Done {{.file}}.",
	}
	if got := goTemplate(icu); !reflect.DeepEqual(got, want) {
		t.Errorf("got %#v, want %#v", got, want)
	}
}

func TestIcuEscapeKeepsPlaceholdersAndQuotesBraces(t *testing.T) {
	cases := map[string]string{
		"Hello {name}":     "Hello {name}",
		"func() { x }":     "func() '{' x '}'",
		"'{name}'":         "''{name}'",
		"a literal } here": "a literal '}' here",
	}
	for input, want := range cases {
		if got := icuEscape(input); got != want {
			t.Errorf("icuEscape(%q) = %q, want %q", input, got, want)
		}
	}
	if got := goPlaceholders(icuEscape("func() { {name} }")); got != "func() { {{.name}} }" {
		t.Errorf("round trip = %q", got)
	}
}

func TestLegacyKeyIsStable(t *testing.T) {
	entry := poEntry{ID: "The program arguments have an unclosed quote."}
	first, second := legacyKey(entry), legacyKey(entry)
	if first != second || !strings.HasPrefix(first, "legacy.the_program_arguments_have_an_") {
		t.Errorf("key = %q", first)
	}
}
