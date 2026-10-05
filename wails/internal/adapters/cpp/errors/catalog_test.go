package errors

import (
	"strings"
	"testing"

	"github.com/codeplai/VizcachaIDE/wails/internal/domain"
)

// fixtureOf names the real output (testdata/cpp_output/{gcc,clang}: GCC 16.2 and Clang 23.1) that
// each catalog id recognises, as {gcc fixture, clang fixture}.
var fixtureOf = map[string][2]string{
	"CPP-COUT-UNDECLARED":      {"cout_no_include", "cout_no_include"},
	"CPP-UNDECLARED":           {"undeclared", "undeclared"},
	"CPP-EXPECTED-SEMICOLON":   {"missing_semicolon", "missing_semicolon"},
	"CPP-EXPECTED-BRACE":       {"missing_brace", "missing_brace"},
	"CPP-EXPECTED-PRIMARY":     {"expected_primary", "expected_primary"},
	"CPP-NO-MATCHING-FUNCTION": {"no_matching_overload", "no_matching_overload"},
	"CPP-TOO-FEW-ARGS":         {"too_few_args_ptr", "too_few_args_ptr"},
	"CPP-TOO-MANY-ARGS":        {"too_many_args_ptr", "too_many_args_ptr"},
	"CPP-CANNOT-CONVERT":       {"cannot_convert", "cannot_convert"},
	"CPP-INVALID-OPERANDS":     {"invalid_operands", "invalid_operands"},
	"CPP-NO-MEMBER":            {"no_member", "no_member"},
	"CPP-REDECLARED":           {"redeclared", "redeclared"},
	"CPP-MISSING-RETURN":       {"missing_return", "missing_return"},
	"CPP-NO-SUCH-FILE":         {"no_such_file", "no_such_file"},
	"CPP-UNUSED-VARIABLE":      {"unused_variable", "unused_variable"},
	"CPP-UNINITIALIZED":        {"uninitialized", "uninitialized"},
	"CPP-SIGN-COMPARE":         {"sign_compare", "sign_compare"},
	"CPP-ASSIGN-IN-CONDITION":  {"assign_in_condition", "assign_in_condition"},
	"CPP-ARRAY-BOUNDS":         {"array_bounds", "array_bounds"},
	"CPP-STRING-COMPARE":       {"string_compare", "string_compare"},
	"CPP-UNDEFINED-REFERENCE":  {"undefined_reference", "undefined_reference"},
	"CPP-UNDEFINED-MAIN":       {"undefined_main", "undefined_main"},
	"CPP-SEGFAULT":             {"segfault", "segfault"},
	"CPP-STACK-OVERFLOW":       {"stack_overflow", "stack_overflow"},
	"CPP-DIVIDE-ZERO":          {"divide_zero", "divide_zero"},
	"CPP-OUT-OF-RANGE":         {"out_of_range", "out_of_range"},
	"CPP-TERMINATE":            {"terminate", "terminate"},
}

func TestEveryCatalogIDHasAFixtureInBothFamilies(t *testing.T) {
	var ids []string
	for _, id := range catalogIDs(t) {
		if !strings.HasPrefix(id, vcpkgIDPrefix) { // the libraries have their own tests (vcpkg_test.go)
			ids = append(ids, id)
		}
	}
	if len(ids) != len(fixtureOf) {
		t.Errorf("catalog has %d entries, fixtures cover %d", len(ids), len(fixtureOf))
	}
	for _, id := range ids {
		if _, ok := fixtureOf[id]; !ok || !strings.HasPrefix(id, "CPP-") {
			t.Errorf("id %q has no fixture or a wrong prefix", id)
		}
	}
}

func TestFixturesProduceTheirIDAndTheirExplanation(t *testing.T) {
	explainer := newExplainer(t)
	for id, names := range fixtureOf {
		for index, family := range families {
			diagnostics, _ := parseFixture(t, family, names[index])
			found := withCode(diagnostics, id)
			if found == nil {
				t.Errorf("%s (%s/%s): codes = %v", id, family, names[index], codes(diagnostics))
				continue
			}
			for _, language := range []string{"en", "es"} {
				explanation := explainer.Explain(*found, language)
				if explanation == nil || explanation.ExplanationID != id || unfilled.MatchString(explanation.Title+explanation.Body+explanation.FixHint) {
					t.Errorf("%s/%s/%s: explanation = %+v", id, family, language, explanation)
				}
			}
		}
	}
}

func TestCoutUndeclaredSuggestsIncludeAndStd(t *testing.T) {
	explainer := newExplainer(t)
	for _, family := range families {
		diagnostics, _ := parseFixture(t, family, "cout_no_std")
		if len(diagnostics) != 2 || diagnostics[0].Code != "CPP-COUT-UNDECLARED" || diagnostics[1].Code != "CPP-COUT-UNDECLARED" {
			t.Fatalf("%s: diagnostics = %v", family, codes(diagnostics))
		}
		en := explainer.Explain(diagnostics[0], "en")
		if en.Placeholders["name"] != "cout" || !strings.Contains(en.FixHint, "#include <iostream>") || !strings.Contains(en.FixHint, "std::") {
			t.Errorf("%s: en = %+v", family, en)
		}
		if es := explainer.Explain(diagnostics[0], "es"); !strings.Contains(es.Title, `"cout"`) || !strings.Contains(es.FixHint, "#include <iostream>") {
			t.Errorf("%s: es = %+v", family, es)
		}
	}
}

func TestEachFamilyWordsTheSameMistakeItsOwnWay(t *testing.T) {
	gcc, _ := parseFixture(t, "gcc", "no_matching_function")
	clang, _ := parseFixture(t, "clang", "no_matching_function")
	if withCode(gcc, "CPP-CANNOT-CONVERT") == nil || withCode(clang, "CPP-NO-MATCHING-FUNCTION") == nil {
		t.Errorf("gcc = %v, clang = %v", codes(gcc), codes(clang))
	}
}

func TestPlaceholdersAreFilledInEnglishAndSpanish(t *testing.T) {
	explainer := newExplainer(t)
	diagnostic := domain.Diagnostic{Message: "'std::string' {aka 'class std::__cxx11::basic_string<char>'} has no member named 'lenght'; did you mean 'length'?"}

	en := explainer.Explain(diagnostic, "en")
	if en.ExplanationID != "CPP-NO-MEMBER" || en.Placeholders["type"] != "std::string" || en.Placeholders["name"] != "lenght" {
		t.Errorf("en = %+v", en)
	}
	if es := explainer.Explain(diagnostic, "es"); !strings.Contains(es.Title, `"lenght"`) || strings.Contains(es.Body, "{") {
		t.Errorf("es = %+v", es)
	}
	if fr := explainer.Explain(diagnostic, "fr"); fr.Title != en.Title {
		t.Errorf("unknown language did not fall back to English: %+v", fr)
	}
}

func TestNamesWithAccentsAndTemplatesAreCaptured(t *testing.T) {
	explainer := newExplainer(t)
	cases := map[string][2]string{
		"'año' was not declared in this scope":                        {"CPP-UNDECLARED", "año"},
		"use of undeclared identifier 'Ñandú::contador'":              {"CPP-UNDECLARED", "Ñandú::contador"},
		"no member named 'push' in 'std::vector<std::pair<int,int>>'": {"CPP-NO-MEMBER", "push"},
		"undefined symbol: Cuenta<int>::saldo()":                      {"CPP-UNDEFINED-REFERENCE", "Cuenta<int>::saldo()"},
	}
	for message, want := range cases {
		got := explainer.Explain(domain.Diagnostic{Message: message}, "es")
		if got == nil || got.ExplanationID != want[0] || !strings.Contains(got.Title+got.Body+got.FixHint, want[1]) {
			t.Errorf("%q: explanation = %+v", message, got)
		}
	}
}

func TestOtherShapesAreRecognised(t *testing.T) {
	explainer := newExplainer(t)
	messages := map[string]string{
		"'string' was not declared in this scope; did you mean 'std::string'?":                       "CPP-COUT-UNDECLARED",
		"no member named 'sort' in namespace 'std'":                                                  "CPP-COUT-UNDECLARED",
		"use of undeclared identifier 'std'":                                                         "CPP-COUT-UNDECLARED",
		"expected ';' before '}' token":                                                              "CPP-EXPECTED-SEMICOLON",
		"expected ';' after expression":                                                              "CPP-EXPECTED-SEMICOLON",
		"expected unqualified-id before '}' token":                                                   "CPP-EXPECTED-BRACE",
		"expected primary-expression before ')' token":                                               "CPP-EXPECTED-PRIMARY",
		"no matching function for call to 'saludar(int, const char [5])'":                            "CPP-NO-MATCHING-FUNCTION",
		"too few arguments to function 'int sumar(int, int)'":                                        "CPP-TOO-FEW-ARGS",
		"invalid conversion from 'const char*' to 'int'":                                             "CPP-CANNOT-CONVERT",
		"cannot initialize a variable of type 'int *' with an rvalue of type 'int'":                  "CPP-CANNOT-CONVERT",
		"invalid operands of types 'int' and 'const char [3]' to binary 'operator+'":                 "CPP-INVALID-OPERANDS",
		"no match for 'operator<<' (operand types are 'std::ostream' and 'Persona')":                 "CPP-INVALID-OPERANDS",
		"variable 'x' set but not used [-Wunused-but-set-variable]":                                  "CPP-UNUSED-VARIABLE",
		"'x' may be used uninitialized [-Wmaybe-uninitialized]":                                      "CPP-UNINITIALIZED",
		"control reaches end of non-void function [-Wreturn-type]":                                   "CPP-MISSING-RETURN",
		"undefined reference to `main'":                                                              "CPP-UNDEFINED-MAIN",
		"undefined reference to `Persona::saludar()'":                                                "CPP-UNDEFINED-REFERENCE",
		"Segmentation fault":                                                                         "CPP-SEGFAULT",
		"terminate called after throwing an instance of 'int'":                                       "CPP-TERMINATE",
		"libc++abi: terminating due to uncaught exception of type int":                               "CPP-TERMINATE",
		"libc++abi: terminating due to uncaught exception of type std::invalid_argument: stoi":       "CPP-TERMINATE",
		"terminate called after throwing an instance of 'std::out_of_range': vector::_M_range_check": "CPP-OUT-OF-RANGE",
	}
	for message, want := range messages {
		got := explainer.Explain(domain.Diagnostic{Message: message}, "en")
		if got == nil || got.ExplanationID != want {
			t.Errorf("%q: explanation = %+v, want %s", message, got, want)
		}
	}
	if got := explainer.Explain(domain.Diagnostic{Message: "unused parameter 'a' [-Wunused-parameter]"}, "en"); got != nil {
		t.Errorf("explanation = %+v, want nil", got)
	}
	terminate := explainer.Explain(domain.Diagnostic{Message: "libc++abi: terminating due to uncaught exception of type std::invalid_argument: stoi"}, "en")
	if terminate.Placeholders["type"] != "std::invalid_argument" {
		t.Errorf("type = %q", terminate.Placeholders["type"])
	}
}

func TestClangdMessagesAreExplainedLikeTheCompilers(t *testing.T) {
	explainer, err := NewExplainer()
	if err != nil {
		t.Fatal(err)
	}
	live := domain.Diagnostic{Severity: domain.SeverityError, Source: "clangd", Message: "Use of undeclared identifier 'totl'; did you mean 'total'? (fix available)"}
	explanation := explainer.Explain(live, "en")
	if explanation == nil || explanation.ExplanationID != "CPP-UNDECLARED" || explanation.Placeholders["name"] != "totl" {
		t.Fatalf("explanation = %+v", explanation)
	}
}
