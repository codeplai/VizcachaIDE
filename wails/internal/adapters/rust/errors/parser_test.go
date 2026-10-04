package errors

import (
	"path/filepath"
	"reflect"
	"strings"
	"testing"

	"github.com/codeplai/VizcachaIDE/wails/internal/domain"
)

func TestEveryFixtureParses(t *testing.T) {
	for _, folder := range fixtureFolders {
		for _, name := range fixtureNames(t, folder) {
			diagnostics, workingDir := parseFixture(t, folder, name)
			if len(diagnostics) == 0 {
				t.Errorf("%s/%s: no diagnostics", folder, name)
			}
			for _, diagnostic := range diagnostics {
				if diagnostic.Message == "" || diagnostic.RawText == "" || diagnostic.Source == "" {
					t.Errorf("%s/%s: incomplete diagnostic %+v", folder, name, diagnostic)
				}
				if diagnostic.Location != nil && !strings.HasPrefix(diagnostic.Location.File, workingDir) {
					t.Errorf("%s/%s: %q is not under %q", folder, name, diagnostic.Location.File, workingDir)
				}
			}
		}
	}
}

type summary struct {
	Code     string
	Severity domain.Severity
	Source   string
	Message  string
	Location domain.SourceLocation
}

func summaries(diagnostics []domain.Diagnostic) []summary {
	result := make([]summary, 0, len(diagnostics))
	for _, diagnostic := range diagnostics {
		item := summary{Code: diagnostic.Code, Severity: diagnostic.Severity, Source: diagnostic.Source, Message: diagnostic.Message}
		if diagnostic.Location != nil {
			item.Location = *diagnostic.Location
		}
		result = append(result, item)
	}
	return result
}

func TestJSONAndTextGiveTheSameCodeAndLocation(t *testing.T) {
	for _, name := range fixtureNames(t, "json") {
		fromJSON, _ := parseFixture(t, "json", name)
		fromText, _ := parseFixture(t, "text", name)
		if !reflect.DeepEqual(summaries(fromJSON), summaries(fromText)) {
			t.Errorf("%s:\n json = %+v\n text = %+v", name, summaries(fromJSON), summaries(fromText))
		}
	}
}

func TestJSONDiagnosticHasTheExactRange(t *testing.T) {
	diagnostics, workingDir := parseFixture(t, "json", "moved")
	if len(diagnostics) != 1 {
		t.Fatalf("diagnostics = %+v (the summaries must be skipped)", diagnostics)
	}
	moved := diagnostics[0]
	file := filepath.Join(workingDir, "main.rs")
	if moved.Code != "E0382" || moved.Severity != domain.SeverityError || moved.Source != "compiler" || moved.Message != "borrow of moved value: `s`" {
		t.Errorf("diagnostic = %+v", moved)
	}
	if *moved.Location != (domain.SourceLocation{File: file, Line: 4, Column: 23}) || *moved.End != (domain.SourceLocation{File: file, Line: 4, Column: 24}) {
		t.Errorf("location = %+v, end = %+v", moved.Location, moved.End)
	}
	if !strings.HasPrefix(moved.RawText, "error[E0382]: borrow of moved value: `s`") || !strings.Contains(moved.RawText, "help: consider cloning") {
		t.Errorf("RawText = %q", moved.RawText)
	}
}

func TestWarningsKeepTheirLintAsCode(t *testing.T) {
	diagnostics, _ := parseFixture(t, "json", "warnings")
	want := []string{"unused_imports", "unused_mut", "unused_variables", "dead_code"}
	if !reflect.DeepEqual(codes(diagnostics), want) {
		t.Fatalf("codes = %v, want %v", codes(diagnostics), want)
	}
	for _, diagnostic := range diagnostics {
		if diagnostic.Severity != domain.SeverityWarning {
			t.Errorf("%s: severity = %s", diagnostic.Code, diagnostic.Severity)
		}
	}
}

func TestClippyLintsHaveTheirOwnSource(t *testing.T) {
	for _, fixture := range [][2]string{{"clippy", "needless_range_loop"}, {"cargo", "clippy_project"}} {
		diagnostics, _ := parseFixture(t, fixture[0], fixture[1])
		if diagnostics[0].Code != "clippy::needless_range_loop" || diagnostics[0].Source != "clippy" || diagnostics[0].Location.Line != 3 {
			t.Errorf("%v: diagnostic = %+v", fixture, diagnostics[0])
		}
	}
}

func TestCargoMessagesAreUnwrappedAndOthersIgnored(t *testing.T) {
	diagnostics, workingDir := parseFixture(t, "cargo", "build_error")
	if len(diagnostics) != 1 || diagnostics[0].Code != "E0382" {
		t.Fatalf("diagnostics = %+v", diagnostics)
	}
	if want := filepath.Join(workingDir, "src", "main.rs"); diagnostics[0].Location.File != want {
		t.Errorf("file = %q, want %q", diagnostics[0].Location.File, want)
	}
}

func TestNetworkFailureOfCargoHasNoLocation(t *testing.T) {
	diagnostics, _ := parseFixture(t, "cargo", "network")
	if len(diagnostics) != 1 || diagnostics[0].Code != "RS-NETWORK" || diagnostics[0].Location != nil {
		t.Fatalf("diagnostics = %+v", diagnostics)
	}
	if !strings.Contains(diagnostics[0].RawText, "Could not connect to server") {
		t.Errorf("RawText lost the Caused by paragraphs: %q", diagnostics[0].RawText)
	}
}

func TestProgramThatPrintsErrorIsNotACompilerError(t *testing.T) {
	output := "error: no such file\nwarning: careful\n"
	if diagnostics := newExplainer(t).Parse(output, "C:/p"); len(diagnostics) != 0 {
		t.Errorf("diagnostics = %+v", diagnostics)
	}
}

func TestWindowsPathsWithSpacesAndAccentsResolve(t *testing.T) {
	workingDir := `C:\Users\Ñandú Pérez\mis programas`
	line := `{"$message_type":"diagnostic","message":"mismatched types","code":{"code":"E0308"},"level":"error","spans":[{"file_name":"src\\ñ.rs","line_start":2,"line_end":2,"column_start":5,"column_end":9,"is_primary":true}],"children":[],"rendered":"error[E0308]: mismatched types\n"}`
	diagnostics := newExplainer(t).Parse(line, workingDir)
	if len(diagnostics) != 1 || diagnostics[0].Location.File != filepath.Join(workingDir, "src", "ñ.rs") && diagnostics[0].Location.File != workingDir+`\src\ñ.rs` {
		t.Fatalf("diagnostics = %+v", diagnostics)
	}
	text := "error[E0308]: mismatched types\n  --> src\\ñ.rs:2:5\n"
	diagnostics = newExplainer(t).Parse(text, workingDir)
	if len(diagnostics) != 1 || diagnostics[0].Location == nil || diagnostics[0].Location.Line != 2 || diagnostics[0].Location.Column != 5 {
		t.Fatalf("text: diagnostics = %+v", diagnostics)
	}
	absolute := `{"message":"x","code":null,"level":"error","spans":[{"file_name":"D:\\otra carpeta\\a.rs","line_start":1,"line_end":1,"column_start":1,"column_end":2,"is_primary":true}],"rendered":"x"}`
	if got := newExplainer(t).Parse(absolute, workingDir); got[0].Location.File != `D:\otra carpeta\a.rs` {
		t.Errorf("absolute path changed: %+v", got[0].Location)
	}
}

func TestNonDiagnosticJSONLinesAreIgnored(t *testing.T) {
	output := `{"reason":"compiler-artifact","package_id":"x"}
{"$message_type":"artifact","artifact":"a","emit":"link"}
{"reason":"build-finished","success":true}
not json {
{"broken
`
	if got := newExplainer(t).Parse(output, "/w"); len(got) != 0 {
		t.Errorf("diagnostics = %+v", got)
	}
}
