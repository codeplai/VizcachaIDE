package errors

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/codeplai/VizcachaIDE/wails/internal/domain"
)

// fileOf returns the last element of a Windows or POSIX path.
func fileOf(path string) string {
	return path[strings.LastIndexAny(path, `\/`)+1:]
}

func TestTracebackLocationMessageAndSource(t *testing.T) {
	workingDir, output := readFixture(t, "name_error")
	diagnostics := newExplainer(t).Parse(output, workingDir)
	if len(diagnostics) != 1 {
		t.Fatalf("diagnostics = %+v", diagnostics)
	}
	got := diagnostics[0]
	if got.Source != "runtime" || got.Severity != domain.SeverityError || got.Message != "NameError: name 'nombre' is not defined" {
		t.Errorf("diagnostic = %+v", got)
	}
	if got.Location == nil || fileOf(got.Location.File) != "name_error.py" || got.Location.Line != 2 {
		t.Errorf("location = %+v", got.Location)
	}
	if !strings.HasPrefix(got.RawText, "Traceback (most recent call last):") || !strings.HasSuffix(got.RawText, got.Message) || strings.Contains(got.RawText, "hola\n") {
		t.Errorf("rawText = %q", got.RawText)
	}
}

func TestTheLocationIsTheLastFrameOfTheStudentsCode(t *testing.T) {
	workingDir, output := readFixture(t, "nested_traceback")
	got := newExplainer(t).Parse(output, workingDir)[0]
	if got.Location.Line != 2 {
		t.Errorf("location = %+v, want the innermost frame (line 2)", got.Location)
	}
}

func TestLibraryFramesAreNeverTheLocation(t *testing.T) {
	output := "Traceback (most recent call last):\n" +
		"  File \"/home/ana/prog/main.py\", line 3, in <module>\n    json.loads(x)\n" +
		"  File \"/usr/lib/python3.12/json/__init__.py\", line 346, in loads\n    return _default_decoder.decode(s)\n" +
		"  File \"C:\\Python312\\Lib\\json\\decoder.py\", line 337, in decode\n    obj = 1\n" +
		"  File \"<frozen runpy>\", line 198, in _run_module_as_main\n" +
		"json.decoder.JSONDecodeError: Expecting value: line 1 column 1 (char 0)\n"
	for _, dir := range []string{"/home/ana/prog", ""} {
		got := newExplainer(t).Parse(output, dir)
		if len(got) != 1 || got[0].Location == nil || got[0].Location.Line != 3 {
			t.Errorf("workingDir %q: diagnostics = %+v %+v", dir, got, got[0].Location)
		}
	}
}

func TestChainedExceptionsReportTheLastOne(t *testing.T) {
	output := "Traceback (most recent call last):\n  File \"/w/a.py\", line 1, in <module>\n    int('x')\nValueError: bad\n\n" +
		"During handling of the above exception, another exception occurred:\n\n" +
		"Traceback (most recent call last):\n  File \"/w/a.py\", line 4, in <module>\n    raise KeyError('k')\nKeyError: 'k'\n"
	got := newExplainer(t).Parse(output, "/w")
	if len(got) != 1 || got[0].Message != "KeyError: 'k'" || got[0].Location.Line != 4 || got[0].Code != "PY-KEY-ERROR" {
		t.Errorf("diagnostics = %+v", got)
	}
}

func TestSyntaxErrorsWithoutTraceback(t *testing.T) {
	cases := []struct {
		fixture, message string
		line, column     int
	}{
		{"syntax_missing_colon", "SyntaxError: expected ':'", 1, 18},
		{"syntax_unclosed", "SyntaxError: '(' was never closed", 1, 6},
		{"syntax_unterminated_string", "SyntaxError: unterminated string literal (detected at line 1)", 1, 7},
		{"indentation_expected", "IndentationError: expected an indented block after function definition on line 1", 2, 1},
		{"indentation_unexpected", "IndentationError: unexpected indent", 2, 1},
		{"tab_error", "TabError: inconsistent use of tabs and spaces in indentation", 3, 1},
	}
	for _, c := range cases {
		workingDir, output := readFixture(t, c.fixture)
		got := newExplainer(t).Parse(output, workingDir)
		if len(got) != 1 || got[0].Source != "syntax" || got[0].Message != c.message {
			t.Errorf("%s: diagnostics = %+v", c.fixture, got)
			continue
		}
		location := got[0].Location
		if location == nil || fileOf(location.File) != c.fixture+".py" || location.Line != c.line || location.Column != c.column {
			t.Errorf("%s: location = %+v, want line %d column %d", c.fixture, location, c.line, c.column)
		}
	}
}

func TestCannotOpenFileHasNoLocation(t *testing.T) {
	workingDir, output := readFixture(t, "cant_open_file")
	got := newExplainer(t).Parse(output, workingDir)
	if len(got) != 1 || got[0].Location != nil || !strings.Contains(got[0].Message, "can't open file") || got[0].Source != "runtime" {
		t.Errorf("diagnostics = %+v", got)
	}
}

func TestOutputWithoutErrorsGivesNothing(t *testing.T) {
	noise := "hola\n/tmp/x.py:3: DeprecationWarning: x is deprecated\n  import imp\n\nfin\n"
	if got := newExplainer(t).Parse(noise, "/tmp"); len(got) != 0 {
		t.Errorf("diagnostics = %+v", got)
	}
	if got := newExplainer(t).Parse("", ""); len(got) != 0 {
		t.Errorf("diagnostics = %+v", got)
	}
}

// ruffFixture is ruff_check.json: the report is what ruff check printed.
func ruffFixture(t *testing.T) (workingDir, report string) {
	t.Helper()
	raw, err := os.ReadFile(filepath.Join("testdata", "python_output", "ruff_check.json"))
	if err != nil {
		t.Fatal(err)
	}
	var recorded struct {
		WorkingDir string          `json:"workingDir"`
		Report     json.RawMessage `json:"report"`
	}
	if err := json.Unmarshal(raw, &recorded); err != nil {
		t.Fatal(err)
	}
	return recorded.WorkingDir, string(recorded.Report)
}

func TestRuffFixtureProducesTheWarningIDs(t *testing.T) {
	explainer := newExplainer(t)
	workingDir, report := ruffFixture(t)
	diagnostics := explainer.Parse(report, workingDir)
	if len(diagnostics) != 3 {
		t.Fatalf("diagnostics = %+v", diagnostics)
	}
	want := map[string]int{"F401": 1, "F841": 4, "F821": 5}
	for _, d := range diagnostics {
		if d.Source != "ruff" || d.Severity != domain.SeverityWarning || d.Location == nil || d.Location.Line != want[d.Code] || fileOf(d.Location.File) != "ruff_case.py" {
			t.Errorf("diagnostic = %+v", d)
		}
		explanation := explainer.Explain(d, "es")
		if explanation == nil || explanation.ExplanationID != ruffIDs[d.Code] {
			t.Errorf("%s: explanation = %+v", d.Code, explanation)
		}
	}
	if diagnostics[0].Location.Column != 8 || diagnostics[0].End == nil || diagnostics[0].End.Column != 10 {
		t.Errorf("first = %+v", diagnostics[0])
	}
}

func TestRuffSyntaxErrorHasNoCode(t *testing.T) {
	report := `[{"code": null, "message": "SyntaxError: Expected ':', found newline", "filename": "a.py",
		"location": {"row": 2, "column": 0}, "end_location": {"row": 2, "column": 3}}]`
	got := newExplainer(t).Parse(report, "/w")
	if len(got) != 1 || got[0].Code != "E999" || got[0].Severity != domain.SeverityError || got[0].Location.Column != 1 {
		t.Errorf("diagnostics = %+v", got)
	}
	if got[0].Location.File != filepath.Join("/w", "a.py") {
		t.Errorf("file = %q", got[0].Location.File)
	}
	if bad := newExplainer(t).Parse("[not json", "/w"); len(bad) != 0 {
		t.Errorf("diagnostics = %+v", bad)
	}
}
