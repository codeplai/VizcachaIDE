package errors

import (
	"path/filepath"
	"strings"
	"testing"

	"github.com/codeplai/VizcachaIDE/wails/internal/domain"
)

func TestWindowsPathsWithSpacesAndAccentsResolveToTheFile(t *testing.T) {
	for _, family := range families {
		diagnostics, workingDir := parseFixture(t, family, "undeclared")
		first := diagnostics[0]
		if first.Location == nil || first.Location.File != workingDir+`\main.cpp` || first.Location.Line != 3 || first.Location.Column != 12 {
			t.Errorf("%s: location = %+v, want %s:3:12", family, first.Location, workingDir)
		}
		if first.Severity != domain.SeverityError || first.Source != "compiler" || first.Code != "CPP-UNDECLARED" {
			t.Errorf("%s: diagnostic = %+v", family, first)
		}
	}
}

func TestRelativePathsAreJoinedToTheWorkingDirectory(t *testing.T) {
	dir := filepath.Join("mis programas", "ñandú")
	output := "main.cpp:3:12: error: use of undeclared identifier 'totl'\n"
	got := newExplainer(t).Parse(output, dir)
	if len(got) != 1 || got[0].Location.File != filepath.Join(dir, "main.cpp") {
		t.Errorf("diagnostics = %+v", got)
	}
}

func TestNotesAndContextAreAppendedToTheDiagnosticBefore(t *testing.T) {
	for _, family := range families {
		diagnostics, _ := parseFixture(t, family, "missing_brace")
		if len(diagnostics) != 1 {
			t.Fatalf("%s: diagnostics = %+v", family, diagnostics)
		}
		raw := diagnostics[0].RawText
		if !strings.Contains(raw, "to match this '{'") || !strings.Contains(raw, "return 0;") || strings.Contains(raw, "generated.") {
			t.Errorf("%s: raw = %q", family, raw)
		}
		if strings.Contains(diagnostics[0].Message, "\n") || diagnostics[0].Severity != domain.SeverityError {
			t.Errorf("%s: diagnostic = %+v", family, diagnostics[0])
		}
	}
}

func TestInFunctionIncludeChainsAndTotalsAreFolded(t *testing.T) {
	for _, family := range families {
		diagnostics, _ := parseFixture(t, family, "cout_no_std")
		for _, diagnostic := range diagnostics {
			for _, noise := range []string{"In function", "In file included", "errors generated"} {
				if strings.Contains(diagnostic.RawText, noise) {
					t.Errorf("%s: raw text has %q: %q", family, noise, diagnostic.RawText)
				}
			}
		}
		if !strings.Contains(diagnostics[0].RawText, "declared here") {
			t.Errorf("%s: note missing: %q", family, diagnostics[0].RawText)
		}
	}
}

func TestLongCandidateListsAreCapped(t *testing.T) {
	diagnostics, _ := parseFixture(t, "gcc", "invalid_operands")
	if len(diagnostics) != 1 || strings.Count(diagnostics[0].RawText, "\n") >= maxRawLines {
		t.Errorf("diagnostics = %d, raw lines = %d", len(diagnostics), strings.Count(diagnostics[0].RawText, "\n"))
	}
}

func TestWarningsAndErrorsOfTheSameCompilationAreSeparated(t *testing.T) {
	for _, family := range families {
		diagnostics, _ := parseFixture(t, family, "no_matching_function")
		var errors, warnings int
		for _, diagnostic := range diagnostics {
			switch diagnostic.Severity {
			case domain.SeverityError:
				errors++
			case domain.SeverityWarning:
				warnings++
			}
		}
		if errors != 1 || warnings != 2 {
			t.Errorf("%s: %d errors and %d warnings in %+v", family, errors, warnings, codes(diagnostics))
		}
	}
}

func TestFatalErrorIsAnError(t *testing.T) {
	for _, family := range families {
		diagnostics, _ := parseFixture(t, family, "no_such_file")
		if len(diagnostics) != 1 || diagnostics[0].Severity != domain.SeverityError || diagnostics[0].Location.Line != 1 {
			t.Errorf("%s: diagnostics = %+v", family, diagnostics)
		}
	}
}

func TestLinkerErrors(t *testing.T) {
	want := map[string]string{
		"gcc":   "undefined reference to `calcular(int)'",
		"clang": "undefined symbol: calcular(int)",
	}
	for _, family := range families {
		diagnostics, workingDir := parseFixture(t, family, "undefined_reference")
		if len(diagnostics) != 1 {
			t.Fatalf("%s: diagnostics = %+v (collect2 and the linker summary are noise)", family, diagnostics)
		}
		got := diagnostics[0]
		if got.Source != "linker" || got.Message != want[family] || got.Severity != domain.SeverityError {
			t.Errorf("%s: diagnostic = %+v", family, got)
		}
		if got.Location == nil || got.Location.Line != 3 || !strings.HasPrefix(filepath.ToSlash(got.Location.File), filepath.ToSlash(workingDir)) {
			t.Errorf("%s: location = %+v", family, got.Location)
		}
	}
}

func TestMissingMainHasNoLocationInsideTheRuntimeOfTheCompiler(t *testing.T) {
	for _, family := range families {
		diagnostics, _ := parseFixture(t, family, "undefined_main")
		if len(diagnostics) != 1 || diagnostics[0].Code != "CPP-UNDEFINED-MAIN" || diagnostics[0].Location != nil {
			t.Errorf("%s: diagnostics = %+v", family, diagnostics)
		}
	}
}

func TestRuntimeCrashesComeFromTheRunnersLine(t *testing.T) {
	want := map[string]string{
		"segfault":       "Segmentation fault",
		"stack_overflow": "Stack overflow",
		"divide_zero":    "Floating point exception",
	}
	for name, message := range want {
		for _, family := range families {
			diagnostics, _ := parseFixture(t, family, name)
			crash := diagnostics[len(diagnostics)-1]
			if crash.Message != message || crash.Source != "runtime" || crash.Location != nil || crash.Severity != domain.SeverityError {
				t.Errorf("%s/%s: crash = %+v", family, name, crash)
			}
		}
	}
}

func TestUncaughtExceptionsFoldTheAbortedLine(t *testing.T) {
	for _, family := range families {
		diagnostics, _ := parseFixture(t, family, "terminate")
		if len(diagnostics) != 1 || diagnostics[0].Source != "runtime" || !strings.Contains(diagnostics[0].Message, "std::runtime_error") || !strings.Contains(diagnostics[0].Message, "algo salio mal") {
			t.Errorf("%s: diagnostics = %+v", family, diagnostics)
		}
	}
	gcc, _ := parseFixture(t, "gcc", "terminate")
	if !strings.Contains(gcc[0].RawText, "what():  algo salio mal") {
		t.Errorf("raw = %q", gcc[0].RawText)
	}
}

func TestAbortedAloneIsAnUnexplainedRuntimeError(t *testing.T) {
	got := newExplainer(t).Parse("Assertion failed\nAborted\n", "")
	if len(got) != 1 || got[0].Message != "Aborted" || got[0].Code != "" {
		t.Errorf("diagnostics = %+v", got)
	}
}

func TestProgramOutputIsNotContextOfTheDiagnosticBefore(t *testing.T) {
	output := "main.cpp:1:5: warning: unused variable 'a' [-Wunused-variable]\n    1 | int a;\n      |     ^\nhola mundo\n    texto con sangría\n"
	got := newExplainer(t).Parse(output, "")
	if len(got) != 1 || strings.Contains(got[0].RawText, "hola") || strings.Contains(got[0].RawText, "sangría") {
		t.Errorf("diagnostics = %+v", got)
	}
}

func TestCheckerOutputYieldsWarnings(t *testing.T) {
	ids := map[string]string{"unused_variable": "CPP-UNUSED-VARIABLE", "sign_compare": "CPP-SIGN-COMPARE", "uninitialized": "CPP-UNINITIALIZED", "string_compare": "CPP-STRING-COMPARE"}
	for name, id := range ids {
		for _, family := range families {
			diagnostics, _ := parseFixture(t, family, name)
			if len(diagnostics) == 0 || diagnostics[0].Severity != domain.SeverityWarning || diagnostics[0].Code != id || diagnostics[0].Source != "compiler" {
				t.Errorf("%s/%s: diagnostics = %+v", family, name, diagnostics)
			}
		}
	}
}

func TestEmptyOutputHasNoDiagnostics(t *testing.T) {
	if got := newExplainer(t).Parse("", `C:\a`); len(got) != 0 {
		t.Errorf("diagnostics = %+v", got)
	}
}
