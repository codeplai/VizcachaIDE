package errors

import (
	"encoding/json"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"testing"

	"github.com/codeplai/VizcachaIDE/wails/internal/domain"
	"github.com/codeplai/VizcachaIDE/wails/internal/protocol/errorcatalog"
)

// fixtureOf names the real CPython 3.12.14 output (testdata/python_output) that each catalog id
// recognises. W-PY-* ids come from ruff_check.json (see TestRuffFixtureProducesTheWarningIDs).
var fixtureOf = map[string]string{
	"PY-NAME-ERROR":             "name_error",
	"PY-UNBOUND-LOCAL":          "unbound_local",
	"PY-INDENTATION":            "indentation_expected",
	"PY-TAB-ERROR":              "tab_error",
	"PY-SYNTAX-INVALID":         "syntax_invalid",
	"PY-SYNTAX-MISSING-COLON":   "syntax_missing_colon",
	"PY-SYNTAX-UNCLOSED":        "syntax_unclosed",
	"PY-SYNTAX-STRING":          "syntax_unterminated_string",
	"PY-SYNTAX-ASSIGN":          "syntax_assign",
	"PY-TYPE-OPERANDS":          "type_operands",
	"PY-TYPE-CONCAT":            "type_concat",
	"PY-TYPE-NOT-SUBSCRIPTABLE": "type_not_subscriptable",
	"PY-TYPE-NOT-CALLABLE":      "type_not_callable",
	"PY-TYPE-NOT-ITERABLE":      "type_not_iterable",
	"PY-TYPE-MISSING-ARGS":      "missing_args",
	"PY-TYPE-TOO-MANY-ARGS":     "type_too_many_args",
	"PY-TYPE-INT-STR":           "type_compare",
	"PY-VALUE-LITERAL":          "value_literal",
	"PY-ZERO-DIVISION":          "nested_traceback",
	"PY-INDEX-RANGE":            "index_range",
	"PY-KEY-ERROR":              "key_error",
	"PY-ATTRIBUTE":              "attribute_error",
	"PY-MODULE-NOT-FOUND":       "module_not_found",
	"PY-IMPORT-NAME":            "import_name",
	"PY-RECURSION":              "recursion",
	"PY-FILE-NOT-FOUND":         "file_not_found",
	"PY-EOF-INPUT":              "eof_input",
	"PY-ASSERTION":              "assertion",
}

var unfilled = regexp.MustCompile(`{w+}`)

var ruffIDs = map[string]string{"F841": "W-PY-UNUSED-VAR", "F401": "W-PY-UNUSED-IMPORT", "F821": "W-PY-UNDEFINED"}

func newExplainer(t *testing.T) *errorcatalog.Explainer {
	t.Helper()
	explainer, err := NewExplainer()
	if err != nil {
		t.Fatal(err)
	}
	return explainer
}

func catalogIDs(t *testing.T) []string {
	t.Helper()
	var entries []struct {
		ID string `json:"id"`
	}
	if err := json.Unmarshal(catalogJSON, &entries); err != nil {
		t.Fatal(err)
	}
	ids := make([]string, 0, len(entries))
	for _, item := range entries {
		ids = append(ids, item.ID)
	}
	return ids
}

// readFixture returns the working directory of the header and the output of the program.
func readFixture(t *testing.T, name string) (workingDir, output string) {
	t.Helper()
	raw, err := os.ReadFile(filepath.Join("testdata", "python_output", name+".txt"))
	if err != nil {
		t.Fatal(err)
	}
	header, output, found := strings.Cut(string(raw), "# output:\n")
	if !found {
		t.Fatalf("%s: no '# output:' line", name)
	}
	for _, line := range strings.Split(header, "\n") {
		if value, ok := strings.CutPrefix(line, "# workingDir: "); ok {
			workingDir = strings.TrimSpace(value)
		}
	}
	return workingDir, output
}

func TestEveryCatalogIDHasAFixture(t *testing.T) {
	ids := catalogIDs(t)
	if len(ids) != len(fixtureOf)+len(ruffIDs) {
		t.Errorf("catalog has %d entries, fixtures cover %d", len(ids), len(fixtureOf)+len(ruffIDs))
	}
	for _, id := range ids {
		_, fromPython := fixtureOf[id]
		isWarning := strings.HasPrefix(id, "W-PY-")
		if !fromPython && !isWarning || !strings.HasPrefix(id, "PY-") && !isWarning {
			t.Errorf("id %q has no fixture or a wrong prefix", id)
		}
	}
}

func TestFixturesProduceTheirID(t *testing.T) {
	explainer := newExplainer(t)
	for id, name := range fixtureOf {
		workingDir, output := readFixture(t, name)
		diagnostics := explainer.Parse(output, workingDir)
		if len(diagnostics) != 1 || diagnostics[0].Code != id {
			t.Errorf("%s (%s): diagnostics = %+v", id, name, diagnostics)
			continue
		}
		for _, language := range []string{"en", "es"} {
			explanation := explainer.Explain(diagnostics[0], language)
			if explanation == nil || explanation.ExplanationID != id || unfilled.MatchString(explanation.Title+explanation.Body+explanation.FixHint) {
				t.Errorf("%s/%s: explanation = %+v", id, language, explanation)
			}
		}
	}
}

func TestPlaceholdersAreFilledInEnglishAndSpanish(t *testing.T) {
	explainer := newExplainer(t)
	diagnostic := domain.Diagnostic{Message: "NameError: name 'nombre' is not defined. Did you mean: 'nombres'?"}

	en := explainer.Explain(diagnostic, "en")
	if en.Title != `Python does not know the name "nombre"` || en.Placeholders["name"] != "nombre" {
		t.Errorf("en = %+v", en)
	}
	if es := explainer.Explain(diagnostic, "es"); !strings.Contains(es.Title, `"nombre"`) || strings.Contains(es.Body, "{") {
		t.Errorf("es = %+v", es)
	}
	if fr := explainer.Explain(diagnostic, "fr"); fr.Title != en.Title {
		t.Errorf("unknown language did not fall back to English: %+v", fr)
	}
}

func TestNamesWithAccentsAreCaptured(t *testing.T) {
	explainer := newExplainer(t)
	got := explainer.Explain(domain.Diagnostic{Message: "NameError: name 'año' is not defined"}, "es")
	if got == nil || got.Placeholders["name"] != "año" {
		t.Errorf("explanation = %+v", got)
	}
}

func TestOtherPythonVersionsAndShapesAreRecognised(t *testing.T) {
	explainer := newExplainer(t)
	messages := map[string]string{
		"UnboundLocalError: local variable 'x' referenced before assignment":                  "PY-UNBOUND-LOCAL",
		"IndentationError: unindent does not match any outer indentation level":               "PY-INDENTATION",
		"SyntaxError: invalid syntax. Perhaps you forgot a comma?":                            "PY-SYNTAX-INVALID",
		"SyntaxError: invalid syntax. Maybe you meant '==' or ':=' instead of '='?":           "PY-SYNTAX-ASSIGN",
		"SyntaxError: unterminated triple-quoted string literal (detected at line 9)":         "PY-SYNTAX-STRING",
		"TypeError: Perro.ladrar() takes 0 positional arguments but 1 was given":              "PY-TYPE-TOO-MANY-ARGS",
		"TypeError: f() takes from 1 to 2 positional arguments but 3 were given":              "PY-TYPE-TOO-MANY-ARGS",
		"TypeError: Perro.__init__() missing 2 required positional arguments: 'a' and 'b'":    "PY-TYPE-MISSING-ARGS",
		"ValueError: could not convert string to float: 'x'":                                  "PY-VALUE-LITERAL",
		"AttributeError: module 'random' has no attribute 'randin'. Did you mean: 'randint'?": "PY-ATTRIBUTE",
		"IndexError: list assignment index out of range":                                      "PY-INDEX-RANGE",
		"ZeroDivisionError: integer modulo by zero":                                           "PY-ZERO-DIVISION",
		// pyflakes through pylsp (live problems) words them differently from ruff.
		"undefined name 'nombre'":                              "W-PY-UNDEFINED",
		"'os' imported but unused":                             "W-PY-UNUSED-IMPORT",
		"local variable 'total' is assigned to but never used": "W-PY-UNUSED-VAR",
		"AssertionError: la edad no puede ser negativa":        "PY-ASSERTION",
		"TypeError: 'NoneType' object is not subscriptable":    "PY-TYPE-NOT-SUBSCRIPTABLE",
	}
	for message, want := range messages {
		got := explainer.Explain(domain.Diagnostic{Message: message}, "en")
		if got == nil || got.ExplanationID != want {
			t.Errorf("%q: explanation = %+v, want %s", message, got, want)
		}
	}
	if got := explainer.Explain(domain.Diagnostic{Message: "KeyboardInterrupt"}, "en"); got != nil {
		t.Errorf("explanation = %+v, want nil", got)
	}
}
