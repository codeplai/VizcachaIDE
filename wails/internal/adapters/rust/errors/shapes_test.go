package errors

import (
	"strings"
	"testing"

	"github.com/codeplai/VizcachaIDE/wails/internal/domain"
)

func TestRustAnalyzerDiagnosticsAreExplained(t *testing.T) {
	explainer := newExplainer(t)
	live := domain.Diagnostic{
		Severity: domain.SeverityError, Source: "rust-analyzer", Code: "E0382",
		Message: "borrow of moved value: `s`\nvalue borrowed here after move",
	}
	explanation := explainer.Explain(live, "en")
	if explanation == nil || explanation.ExplanationID != "RS-MOVED" || explanation.Placeholders["name"] != "s" {
		t.Fatalf("explanation = %+v", explanation)
	}
	// The code alone is enough, whatever the wording.
	for code, id := range map[string]string{"E0499": "RS-BORROW-MUT-TWICE", "E0308": "RS-MISMATCHED-TYPES", "E0599": "RS-NO-METHOD", "unused_variables": "W-RS-UNUSED-VAR"} {
		got := explainer.Explain(domain.Diagnostic{Source: "rust-analyzer", Code: code, Message: "algo que no se parece"}, "es")
		if got == nil || got.ExplanationID != id || unfilledRe.MatchString(got.Title+got.Body+got.FixHint) {
			t.Errorf("%s: explanation = %+v, want %s", code, got, id)
		}
	}
	if got := explainer.Explain(domain.Diagnostic{Source: "rust-analyzer", Code: "E9999", Message: "something new"}, "en"); got != nil {
		t.Errorf("explanation = %+v, want nil", got)
	}
}

func TestClippyLintNameIsInTheExplanation(t *testing.T) {
	explainer := newExplainer(t)
	for _, language := range []string{"en", "es"} {
		got := explainer.Explain(domain.Diagnostic{Code: "clippy::redundant_clone", Message: "redundant clone"}, language)
		if got == nil || got.ExplanationID != "W-RS-CLIPPY" || !strings.Contains(got.Title, "clippy::redundant_clone") {
			t.Errorf("%s: explanation = %+v", language, got)
		}
	}
}

func TestOtherShapesAreRecognised(t *testing.T) {
	explainer := newExplainer(t)
	messages := map[string]string{
		"use of moved value: `v`":                                                       "RS-MOVED",
		"cannot find type `Rc` in this scope":                                           "RS-FAILED-RESOLVE",
		"failed to resolve: use of unresolved module or unlinked crate `rand`":          "RS-FAILED-RESOLVE",
		"cannot find function `sumar` in this scope":                                    "RS-UNRESOLVED-NAME",
		"expected one of `!`, `.`, `::`, `;`, `?`, `{`, `}`, or an operator, found `x`": "RS-EXPECTED-TOKEN",
		"unexpected closing delimiter: `}`":                                             "RS-UNCLOSED-DELIMITER",
		"linker `link.exe` not found":                                                   "RS-LINKER-MISSING",
		"unused variable: `contador`":                                                   "W-RS-UNUSED-VAR",
		"unused imports: `a` and `b`":                                                   "W-RS-UNUSED-IMPORT",
		"field `edad` is never read":                                                    "W-RS-DEAD-CODE",
		"attempt to subtract with overflow":                                             "RS-PANIC-OVERFLOW",
		"attempt to calculate the remainder with a divisor of zero":                     "RS-PANIC-DIV-ZERO",
		"assertion `left == right` failed":                                              "RS-PANIC-EXPLICIT",
		"not yet implemented: later":                                                    "RS-PANIC-EXPLICIT",
		"Error: ParseIntError { kind: Empty }":                                          "RS-MAIN-ERR",
		"Segmentation fault":                                                            "RS-CRASH",
	}
	for message, want := range messages {
		got := explainer.Explain(domain.Diagnostic{Message: message}, "en")
		if (want == "") != (got == nil) || (got != nil && got.ExplanationID != want) {
			t.Errorf("%q: explanation = %+v, want %q", message, got, want)
		}
	}
	got := explainer.Explain(domain.Diagnostic{Message: "called `Result::unwrap()` on an `Err` value: ParseIntError { kind: InvalidDigit }"}, "en")
	if got.Placeholders["error"] != "ParseIntError { kind: InvalidDigit }" {
		t.Errorf("error = %q", got.Placeholders["error"])
	}
}
