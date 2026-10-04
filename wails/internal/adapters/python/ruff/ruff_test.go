package ruff

import (
	"context"
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
	"testing"

	"github.com/codeplai/VizcachaIDE/wails/internal/adapters/python"
	"github.com/codeplai/VizcachaIDE/wails/internal/adapters/python/pythontest"
	"github.com/codeplai/VizcachaIDE/wails/internal/app"
	"github.com/codeplai/VizcachaIDE/wails/internal/domain"
)

type pythonAt struct{ path string }

func (p pythonAt) Load() (domain.Settings, error) {
	settings := domain.DefaultSettings()
	settings.ToolPaths = map[string]string{python.ToolPython: p.path}
	return settings, nil
}
func (pythonAt) Save(domain.Settings) error { return nil }

func realTool(t *testing.T) *Tool {
	t.Helper()
	locator := python.NewLocator(python.Options{Settings: pythonAt{path: pythontest.Interpreter(t)}})
	return New(Config{Locator: locator})
}

func noPythonTool(t *testing.T) *Tool {
	t.Helper()
	missing := filepath.Join(t.TempDir(), "no-python")
	locator := python.NewLocator(python.Options{
		Settings: pythonAt{path: missing}, AppDir: t.TempDir(), BaseEnvironment: []string{"PATH="},
		Launcher: func(context.Context) string { return "" },
	})
	return New(Config{Locator: locator, BaseEnvironment: []string{"PATH="}})
}

func TestFormatFixesTheIndentation(t *testing.T) {
	tool := realTool(t)
	got, err := tool.Format(filepath.Join(t.TempDir(), "main.py"), "x=1\nif x:\n        print( x )\n")
	if err != nil {
		t.Fatal(err)
	}
	if want := "x = 1\nif x:\n    print(x)\n"; got != want {
		t.Errorf("formatted = %q, want %q", got, want)
	}
}

func TestFormatSyntaxErrorWrapsErrFormatWithTheLineTheFrontendReads(t *testing.T) {
	tool := realTool(t)
	_, err := tool.Format("main.py", "x = 1\ndef f(:\n    pass\n")
	if !errors.Is(err, app.ErrFormat) {
		t.Fatalf("err = %v, want ErrFormat", err)
	}
	// The same expression as lineFromFormatError in frontend/src/lib/stores/saving.ts.
	match := regexp.MustCompile(`:(\d+):\d+`).FindStringSubmatch(err.Error())
	if match == nil || match[1] != "2" {
		t.Errorf("message %q does not carry line 2", err)
	}
}

func TestMissingPythonIsMissingTool(t *testing.T) {
	_, err := noPythonTool(t).Format("main.py", "x=1\n")
	if !errors.Is(err, app.ErrToolNotFound) || err.Error() != app.MissingTool("python").Error() {
		t.Errorf("err = %v", err)
	}
	folder := t.TempDir()
	output, err := noPythonTool(t).Check(context.Background(), domain.NewFileRunConfiguration(domain.CodeLanguagePython, filepath.Join(folder, "main.py"), nil))
	if output != "" || err != nil {
		t.Errorf("Check without python = %q, %v; want empty and no error", output, err)
	}
}

func checkFile(t *testing.T, tool *Tool, source string) string {
	t.Helper()
	file := filepath.Join(t.TempDir(), "main.py")
	if err := os.WriteFile(file, []byte(source), 0o600); err != nil {
		t.Fatal(err)
	}
	output, err := tool.Check(context.Background(), domain.NewFileRunConfiguration(domain.CodeLanguagePython, file, nil))
	if err != nil {
		t.Fatal(err)
	}
	return output
}

func codesOf(t *testing.T, output string) []string {
	t.Helper()
	var findings []struct {
		Code string `json:"code"`
	}
	if err := json.Unmarshal([]byte(output), &findings); err != nil {
		t.Fatalf("not JSON: %q: %v", output, err)
	}
	codes := make([]string, 0, len(findings))
	for _, finding := range findings {
		codes = append(codes, finding.Code)
	}
	sort.Strings(codes)
	return codes
}

func TestCheckFindsTheSameCodesAsTheRecordedFixture(t *testing.T) {
	data, err := os.ReadFile(filepath.Join("..", "errors", "testdata", "python_output", "ruff_check.json"))
	if err != nil {
		t.Fatal(err)
	}
	var fixture struct {
		Source string `json:"source"`
		Report []struct {
			Code string `json:"code"`
		} `json:"report"`
	}
	if err := json.Unmarshal(data, &fixture); err != nil {
		t.Fatal(err)
	}
	var want []string
	for _, finding := range fixture.Report {
		want = append(want, finding.Code)
	}
	sort.Strings(want)
	got := codesOf(t, checkFile(t, realTool(t), fixture.Source))
	if strings.Join(got, ",") != strings.Join(want, ",") || len(got) < 3 {
		t.Errorf("codes = %v, want %v (F401, F821, F841)", got, want)
	}
}

func TestCheckIsEmptyWhenThereAreNoFindings(t *testing.T) {
	if got := strings.TrimSpace(checkFile(t, realTool(t), "print(1)\n")); got != "" && got != "[]" {
		t.Errorf("output = %q, want empty", got)
	}
}
