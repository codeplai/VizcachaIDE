package errors

import (
	"encoding/json"
	"strings"
	"testing"

	"github.com/codeplai/VizcachaIDE/wails/internal/domain"
)

// fixtureOf names the real output (testdata/rust_output, rustc 1.99) that each catalog id
// recognises, as "folder/case". The compile errors also have the same case in text/.
var fixtureOf = map[string]string{
	"W-RS-CLIPPY":           "clippy/needless_range_loop",
	"RS-MOVED":              "json/moved",
	"RS-BORROW-MUT-TWICE":   "json/borrow_mut_twice",
	"RS-BORROW-CONFLICT":    "json/borrow_conflict",
	"RS-ASSIGN-BORROWED":    "json/assign_borrowed",
	"RS-MOVE-BORROWED":      "json/move_borrowed",
	"RS-MOVE-OUT-OF-REF":    "json/move_out_of_ref",
	"RS-DANGLING-REF":       "json/dangling_ref",
	"RS-RETURN-LOCAL-REF":   "json/return_local_ref",
	"RS-MISSING-LIFETIME":   "json/missing_lifetime",
	"RS-MISMATCHED-TYPES":   "json/mismatched_types",
	"RS-FAILED-RESOLVE":     "json/failed_resolve",
	"RS-UNKNOWN-MACRO":      "json/unknown_macro",
	"RS-UNRESOLVED-NAME":    "json/unresolved_name",
	"RS-UNRESOLVED-IMPORT":  "json/unresolved_import",
	"RS-NO-METHOD":          "json/no_method",
	"RS-NO-FIELD":           "json/no_field",
	"RS-ASSIGN-TWICE":       "json/assign_twice",
	"RS-NOT-MUTABLE":        "json/not_mutable",
	"RS-TRAIT-BOUND":        "json/trait_bound",
	"RS-WRONG-ARGS":         "json/wrong_args",
	"RS-BINARY-OP":          "json/binary_op",
	"RS-DEREF":              "json/deref",
	"RS-NON-EXHAUSTIVE":     "json/non_exhaustive",
	"RS-PRIVATE":            "json/private",
	"RS-TYPE-ANNOTATIONS":   "json/type_annotations",
	"RS-MISSING-ITEMS":      "json/missing_items",
	"RS-NO-MAIN":            "json/no_main",
	"RS-EXPECTED-TOKEN":     "json/expected_token",
	"RS-UNCLOSED-DELIMITER": "json/unclosed_delimiter",
	"W-RS-UNUSED-VAR":       "json/warnings",
	"W-RS-UNUSED-MUT":       "json/warnings",
	"W-RS-UNUSED-IMPORT":    "json/warnings",
	"W-RS-DEAD-CODE":        "json/warnings",
	"RS-LINKER-MISSING":     "json/linker_missing",
	"RS-NETWORK":            "cargo/network",
	"RS-PANIC-INDEX":        "runtime/panic_index",
	"RS-PANIC-UNWRAP-NONE":  "runtime/panic_unwrap_none",
	"RS-PANIC-UNWRAP-ERR":   "runtime/panic_unwrap_err",
	"RS-PANIC-EXPECT":       "runtime/panic_expect",
	"RS-PANIC-OVERFLOW":     "runtime/panic_overflow",
	"RS-PANIC-DIV-ZERO":     "runtime/panic_div_zero",
	"RS-PANIC-STR-BOUNDARY": "runtime/panic_str_boundary",
	"RS-PANIC-REFCELL":      "runtime/panic_refcell",
	"RS-PANIC-EXPLICIT":     "runtime/panic_explicit",
	"RS-STACK-OVERFLOW":     "runtime/stack_overflow",
	"RS-MAIN-ERR":           "runtime/main_err",
	"RS-CRASH":              "", // no program of the fixtures crashes: see TestRunnerCrashLines
}

func TestEveryCatalogIDHasAFixture(t *testing.T) {
	ids := catalogIDs(t)
	if len(ids) != len(fixtureOf) {
		t.Errorf("catalog has %d entries, fixtures cover %d", len(ids), len(fixtureOf))
	}
	for _, id := range ids {
		if _, ok := fixtureOf[id]; !ok || !strings.HasPrefix(id, "RS-") && !strings.HasPrefix(id, "W-RS-") {
			t.Errorf("id %q has no fixture or a wrong prefix", id)
		}
	}
}

func TestFixturesAreExplainedByTheirID(t *testing.T) {
	explainer := newExplainer(t)
	for id, fixture := range fixtureOf {
		if fixture == "" {
			continue
		}
		folder, name, _ := strings.Cut(fixture, "/")
		folders := []string{folder}
		if folder == "json" && !strings.Contains(name, "warnings") {
			folders = append(folders, "text") // the same mistake in human text
		}
		for _, each := range folders {
			diagnostics, _ := parseFixture(t, each, name)
			for _, language := range []string{"en", "es"} {
				explanation := explainedAs(explainer, diagnostics, id, language)
				if explanation == nil {
					t.Errorf("%s (%s/%s/%s): codes = %v", id, each, name, language, codes(diagnostics))
					continue
				}
				all := explanation.Title + explanation.Body + explanation.FixHint
				if unfilledRe.MatchString(all) || strings.Contains(explanation.Title, "…") {
					t.Errorf("%s/%s/%s: placeholders not filled: %+v", id, each, language, explanation)
				}
			}
		}
	}
}

func TestSpanishUsesTheSamePlaceholdersAsEnglish(t *testing.T) {
	var entries []struct {
		ID string `json:"id"`
		EN struct{ Title, Body, Fix string }
		ES struct{ Title, Body, Fix string }
	}
	if err := json.Unmarshal(catalogJSON, &entries); err != nil {
		t.Fatal(err)
	}
	names := func(texts ...string) string {
		return strings.Join(unfilledRe.FindAllString(strings.Join(texts, " "), -1), ",")
	}
	for _, entry := range entries {
		sorted := func(joined string) map[string]bool {
			set := map[string]bool{}
			for _, name := range strings.Split(joined, ",") {
				set[name] = true
			}
			return set
		}
		en, es := sorted(names(entry.EN.Title, entry.EN.Body, entry.EN.Fix)), sorted(names(entry.ES.Title, entry.ES.Body, entry.ES.Fix))
		for name := range en {
			if !es[name] {
				t.Errorf("%s: English uses %s and Spanish does not", entry.ID, name)
			}
		}
	}
}

func TestOwnershipEntriesHaveAnExample(t *testing.T) {
	explainer := newExplainer(t)
	for _, id := range []string{"RS-MOVED", "RS-BORROW-MUT-TWICE", "RS-BORROW-CONFLICT", "RS-DANGLING-REF", "RS-MISSING-LIFETIME"} {
		folder, name, _ := strings.Cut(fixtureOf[id], "/")
		diagnostics, _ := parseFixture(t, folder, name)
		for _, language := range []string{"en", "es"} {
			if !strings.Contains(explainedAs(explainer, diagnostics, id, language).Body, "let ") && id != "RS-MISSING-LIFETIME" {
				t.Errorf("%s/%s: no example in the body", id, language)
			}
		}
	}
}

func TestPlaceholdersAreFilledInEnglishAndSpanish(t *testing.T) {
	explainer := newExplainer(t)
	diagnostic := domain.Diagnostic{Code: "E0382", Message: "borrow of moved value: `texto`"}

	en := explainer.Explain(diagnostic, "en")
	if en.ExplanationID != "RS-MOVED" || en.Placeholders["name"] != "texto" || !strings.Contains(en.Title, `"texto"`) {
		t.Errorf("en = %+v", en)
	}
	if es := explainer.Explain(diagnostic, "es"); !strings.Contains(es.Title, `"texto"`) || strings.Contains(es.Body, "{name}") {
		t.Errorf("es = %+v", es)
	}
	if fr := explainer.Explain(diagnostic, "fr"); fr.Title != en.Title {
		t.Errorf("unknown language did not fall back to English: %+v", fr)
	}
}
