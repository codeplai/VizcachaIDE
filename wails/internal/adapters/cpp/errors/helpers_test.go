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

var families = []string{"gcc", "clang"}

var unfilled = regexp.MustCompile(`\{\w+\}`)

// crashLines are what the runner writes on stderr when the program ends by a Windows exception
// code (PLAN_CPP.md 4.3); the fixtures hold the output of the program before it.
var crashLines = map[string]string{
	"0xC0000005": "Segmentation fault",
	"0xC00000FD": "Stack overflow",
	"0xC0000094": "Floating point exception",
	"0xC0000409": "Aborted",
}

var runExit = regexp.MustCompile(`# runExit: \d+ \((0x[0-9A-F]+)\)`)

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

// readFixture returns the working directory of the header and what the compiler and the program
// printed, with the crash line of the runner when the recorded run ended by an exception code.
func readFixture(t *testing.T, family, name string) (workingDir, output string) {
	t.Helper()
	raw, err := os.ReadFile(filepath.Join("testdata", "cpp_output", family, name+".txt"))
	if err != nil {
		t.Fatal(err)
	}
	header, output, found := strings.Cut(strings.ReplaceAll(string(raw), "\r\n", "\n"), "# output:\n")
	if !found {
		t.Fatalf("%s/%s: no '# output:' line", family, name)
	}
	for _, line := range strings.Split(header, "\n") {
		if value, ok := strings.CutPrefix(line, "# workingDir: "); ok {
			workingDir = strings.TrimSpace(value)
		}
	}
	if code := runExit.FindStringSubmatch(header); code != nil {
		output += crashLines[code[1]] + "\n"
	}
	return workingDir, output
}

// parseFixture parses the fixture with the catalog of C++.
func parseFixture(t *testing.T, family, name string) (diagnostics []domain.Diagnostic, workingDir string) {
	t.Helper()
	workingDir, output := readFixture(t, family, name)
	return newExplainer(t).Parse(output, workingDir), workingDir
}

func codes(diagnostics []domain.Diagnostic) []string {
	result := make([]string, 0, len(diagnostics))
	for _, diagnostic := range diagnostics {
		result = append(result, diagnostic.Code)
	}
	return result
}

func withCode(diagnostics []domain.Diagnostic, id string) *domain.Diagnostic {
	for index := range diagnostics {
		if diagnostics[index].Code == id {
			return &diagnostics[index]
		}
	}
	return nil
}
