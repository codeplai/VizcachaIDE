package errors

import (
	"encoding/json"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"testing"

	"github.com/codeplai/VizcachaIDE/wails/internal/domain"
)

// fixtureFolders are the sources recorded with rustc, cargo and clippy 1.99 (testdata/rust_output).
var fixtureFolders = []string{"json", "text", "runtime", "clippy", "cargo"}

// crashLines are what the runner writes on stderr when the program ends by a Windows exception
// code; the fixtures hold the output of the program before it.
var crashLines = map[string]string{
	"0xC0000005": "Segmentation fault",
	"0xC00000FD": "Stack overflow",
	"0xC0000094": "Floating point exception",
	"0xC0000409": "Aborted",
}

var (
	runExit    = regexp.MustCompile(`# runExit: \d+ \((0x[0-9A-F]+)\)`)
	unfilledRe = regexp.MustCompile(`\{\w+\}`)
)

func newExplainer(t *testing.T) *Explainer {
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

// readFixture returns the working directory of the header and what the tool or the program
// printed, with the crash line of the runner when the recorded run ended by an exception code.
func readFixture(t *testing.T, folder, name string) (workingDir, output string) {
	t.Helper()
	raw, err := os.ReadFile(filepath.Join("testdata", "rust_output", folder, name+".txt"))
	if err != nil {
		t.Fatal(err)
	}
	header, output, found := strings.Cut(strings.ReplaceAll(string(raw), "\r\n", "\n"), "# output:\n")
	if !found {
		t.Fatalf("%s/%s: no '# output:' line", folder, name)
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

func fixtureNames(t *testing.T, folder string) []string {
	t.Helper()
	files, err := filepath.Glob(filepath.Join("testdata", "rust_output", folder, "*.txt"))
	if err != nil || len(files) == 0 {
		t.Fatalf("no fixtures in %s: %v", folder, err)
	}
	names := make([]string, 0, len(files))
	for _, file := range files {
		names = append(names, strings.TrimSuffix(filepath.Base(file), ".txt"))
	}
	return names
}

// parseFixture parses the fixture with the catalog of Rust.
func parseFixture(t *testing.T, folder, name string) (diagnostics []domain.Diagnostic, workingDir string) {
	t.Helper()
	workingDir, output := readFixture(t, folder, name)
	return newExplainer(t).Parse(output, workingDir), workingDir
}

func codes(diagnostics []domain.Diagnostic) []string {
	result := make([]string, 0, len(diagnostics))
	for _, diagnostic := range diagnostics {
		result = append(result, diagnostic.Code)
	}
	return result
}

// explainedAs returns the catalog id the explanation of any diagnostic has, or "".
func explainedAs(explainer *Explainer, diagnostics []domain.Diagnostic, id, language string) *domain.ErrorExplanation {
	for _, diagnostic := range diagnostics {
		if explanation := explainer.Explain(diagnostic, language); explanation != nil && explanation.ExplanationID == id {
			return explanation
		}
	}
	return nil
}
