package errors

import (
	"path/filepath"
	"strings"
	"testing"

	"github.com/codeplai/VizcachaIDE/wails/internal/domain"
)

func TestPanicsLocateTheStudentsLine(t *testing.T) {
	want := map[string][3]any{
		"panic_index":         {"RS-PANIC-INDEX", 4, 21},
		"panic_unwrap_none":   {"RS-PANIC-UNWRAP-NONE", 3, 23},
		"panic_unwrap_err":    {"RS-PANIC-UNWRAP-ERR", 2, 34},
		"panic_expect":        {"RS-PANIC-EXPECT", 2, 34},
		"panic_overflow":      {"RS-PANIC-OVERFLOW", 4, 9},
		"panic_div_zero":      {"RS-PANIC-DIV-ZERO", 4, 20},
		"panic_str_boundary":  {"RS-PANIC-STR-BOUNDARY", 3, 22},
		"panic_refcell":       {"RS-PANIC-REFCELL", 5, 16},
		"panic_explicit":      {"RS-PANIC-EXPLICIT", 2, 5},
		"panic_unreachable":   {"RS-PANIC-EXPLICIT", 4, 9},
		"panic_nested_unwrap": {"RS-PANIC-UNWRAP-ERR", 2, 15},
	}
	for name, expected := range want {
		diagnostics, workingDir := parseFixture(t, "runtime", name)
		if len(diagnostics) != 1 {
			t.Errorf("%s: diagnostics = %+v", name, diagnostics)
			continue
		}
		got := diagnostics[0]
		location := domain.SourceLocation{File: filepath.Join(workingDir, "main.rs"), Line: expected[1].(int), Column: expected[2].(int)}
		if got.Code != expected[0] || got.Source != "runtime" || got.Severity != domain.SeverityError || got.Location == nil || *got.Location != location {
			t.Errorf("%s: diagnostic = %+v, location = %+v, want %v %+v", name, got, got.Location, expected[0], location)
		}
		if strings.Contains(got.Message, "stack backtrace") || strings.Contains(got.Message, "note:") {
			t.Errorf("%s: noise in the message %q", name, got.Message)
		}
	}
}

func TestPanicInsideTheStandardLibraryUsesTheBacktrace(t *testing.T) {
	std := `/rustc/b940084d7eb6a299eb4bfeb8e34901bc051e7ac4/library\core\src/result.rs:1871:5`
	for _, folder := range [][2]string{{"runtime", "panic_nested_unwrap"}, {"cargo", "run_panic"}} {
		workingDir, output := readFixture(t, folder[0], folder[1])
		student := "main.rs:2:15"
		if folder[0] == "cargo" {
			student = `src\main.rs:4:21`
		}
		output = strings.Replace(output, "panicked at "+student+":", "panicked at "+std+":", 1)
		diagnostics := newExplainer(t).Parse(output, workingDir)
		if len(diagnostics) != 1 || diagnostics[0].Location == nil {
			t.Fatalf("%v: diagnostics = %+v", folder, diagnostics)
		}
		want := domain.SourceLocation{File: filepath.Join(workingDir, "main.rs"), Line: 2, Column: 15}
		if folder[0] == "cargo" {
			want = domain.SourceLocation{File: filepath.Join(workingDir, "src", "main.rs"), Line: 4, Column: 21}
		}
		if *diagnostics[0].Location != want {
			t.Errorf("%v: location = %+v, want %+v (the frame of the student, not std)", folder, *diagnostics[0].Location, want)
		}
		if lines := strings.Count(diagnostics[0].RawText, "\n") + 1; lines > maxRawLines || !strings.Contains(diagnostics[0].RawText, "stack backtrace:") {
			t.Errorf("%v: RawText has %d lines: %q", folder, lines, diagnostics[0].RawText)
		}
	}
}

func TestPanicInTheStandardLibraryWithoutBacktraceHasNoLocation(t *testing.T) {
	output := "thread 'main' (7) panicked at /rustc/abc/library/core/src/option.rs:1:1:\ncalled `Option::unwrap()` on a `None` value\nnote: run with `RUST_BACKTRACE=1` environment variable to display a backtrace\n"
	diagnostics := newExplainer(t).Parse(output, `C:\p`)
	if len(diagnostics) != 1 || diagnostics[0].Location != nil || diagnostics[0].Code != "RS-PANIC-UNWRAP-NONE" {
		t.Fatalf("diagnostics = %+v", diagnostics)
	}
	if !strings.Contains(diagnostics[0].RawText, "RUST_BACKTRACE") || strings.Contains(diagnostics[0].Message, "RUST_BACKTRACE") {
		t.Errorf("noise: message %q, raw %q", diagnostics[0].Message, diagnostics[0].RawText)
	}
}

func TestOtherPanicFormats(t *testing.T) {
	cases := map[string]string{
		"without thread id": "thread 'main' panicked at src/main.rs:7:5:\nexplicit panic\n",
		"before 1.73":       "thread 'main' panicked at 'explicit panic', src/main.rs:7:5\nnote: run with `RUST_BACKTRACE=1` environment variable to display a backtrace\n",
		"named thread":      "thread '<unnamed>' (3) panicked at src/main.rs:7:5:\nexplicit panic\n",
	}
	for name, output := range cases {
		diagnostics := newExplainer(t).Parse(output, "/work/mi proyecto")
		want := domain.SourceLocation{File: filepath.Join("/work/mi proyecto", "src", "main.rs"), Line: 7, Column: 5}
		if len(diagnostics) != 1 || diagnostics[0].Message != "explicit panic" || diagnostics[0].Location == nil || *diagnostics[0].Location != want {
			t.Errorf("%s: diagnostics = %+v", name, diagnostics)
		}
	}
}

func TestAssertPanicKeepsAllItsLinesInRawText(t *testing.T) {
	output := "thread 'main' (1) panicked at main.rs:3:5:\nassertion `left == right` failed\n  left: 1\n right: 2\nnote: run with `RUST_BACKTRACE=1` environment variable to display a backtrace\n"
	diagnostics := newExplainer(t).Parse(output, "/w")
	if len(diagnostics) != 1 || diagnostics[0].Message != "assertion `left == right` failed" || !strings.Contains(diagnostics[0].RawText, " right: 2") {
		t.Fatalf("diagnostics = %+v", diagnostics)
	}
}

func TestPanicMessageThatLooksLikeACompilerErrorIsStillAPanic(t *testing.T) {
	output := "thread 'main' (1) panicked at main.rs:3:5:\nexpected `a`, found `b`\n"
	diagnostics := newExplainer(t).Parse(output, "/w")
	if len(diagnostics) != 1 || diagnostics[0].Code != "RS-PANIC-EXPLICIT" {
		t.Fatalf("diagnostics = %+v", diagnostics)
	}
}

func TestStackOverflowIsOneDiagnostic(t *testing.T) {
	diagnostics, _ := parseFixture(t, "runtime", "stack_overflow") // plus the runner's "Stack overflow"
	if len(diagnostics) != 1 || diagnostics[0].Code != "RS-STACK-OVERFLOW" || diagnostics[0].Location != nil {
		t.Fatalf("diagnostics = %+v", diagnostics)
	}
	output := "thread 'main' (1) has overflowed its stack\nfatal runtime error: stack overflow, aborting\n"
	if got := newExplainer(t).Parse(output, "/w"); len(got) != 1 || got[0].Code != "RS-STACK-OVERFLOW" {
		t.Errorf("diagnostics = %+v", got)
	}
}

func TestRunnerCrashLines(t *testing.T) {
	explainer := newExplainer(t)
	for _, line := range []string{"Segmentation fault", "Aborted", "Floating point exception"} {
		diagnostics := explainer.Parse("antes\n"+line+"\n", "/w")
		if len(diagnostics) != 1 || diagnostics[0].Code != "RS-CRASH" || diagnostics[0].Message != line || diagnostics[0].Source != "runtime" {
			t.Errorf("%s: diagnostics = %+v", line, diagnostics)
		}
		if explainer.Explain(diagnostics[0], "es") == nil {
			t.Errorf("%s: not explained", line)
		}
	}
	if got := explainer.Parse("Segmentation fault happens when...\n", "/w"); len(got) != 0 {
		t.Errorf("only the exact lines count: %+v", got)
	}
}

func TestMainErrorOnlyWhenNothingElseWasFound(t *testing.T) {
	diagnostics, _ := parseFixture(t, "runtime", "main_err")
	if len(diagnostics) != 1 || diagnostics[0].Code != "RS-MAIN-ERR" || diagnostics[0].Message != `Error: "archivo no encontrado"` || diagnostics[0].Location != nil {
		t.Fatalf("diagnostics = %+v", diagnostics)
	}
	explainer := newExplainer(t)
	if got := explainer.Parse("Error: algo\nuna línea más\n", "/w"); len(got) != 0 {
		t.Errorf("Error: is only the last line: %+v", got)
	}
	panicThenError := "thread 'main' (1) panicked at main.rs:1:1:\nexplicit panic\n"
	if got := explainer.Parse(panicThenError, "/w"); len(got) != 1 || got[0].Code == "RS-MAIN-ERR" {
		t.Errorf("diagnostics = %+v", got)
	}
}
