package clippy

import (
	"bufio"
	"context"
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/codeplai/VizcachaIDE/wails/internal/adapters/rust"
	"github.com/codeplai/VizcachaIDE/wails/internal/adapters/rust/rusttest"
	"github.com/codeplai/VizcachaIDE/wails/internal/app"
	"github.com/codeplai/VizcachaIDE/wails/internal/domain"
)

const (
	// loopWithIndex makes clippy say needless_range_loop.
	loopWithIndex = "fn main() {\n    let v = vec![1, 2, 3];\n    for i in 0..v.len() {\n        println!(\"{}\", v[i]);\n    }\n}\n"
	cleanProgram  = "fn main() {\n    let v = vec![1, 2, 3];\n    for x in &v {\n        println!(\"{x}\");\n    }\n}\n"
	// unusedAndLoop has a compiler warning (unused_variables) and a clippy lint.
	unusedAndLoop = "fn main() {\n    let unused = 1;\n    let v = vec![1, 2, 3];\n    for i in 0..v.len() {\n        println!(\"{}\", v[i]);\n    }\n}\n"
	manifest      = "[package]\nname = \"sample\"\nversion = \"0.1.0\"\nedition = \"2021\"\n"
)

func realTool(t *testing.T) *Tool {
	t.Helper()
	return New(rust.NewLocator(rust.Options{AppDir: os.TempDir(), BaseEnvironment: rusttest.Environment(t)}))
}

func writeFile(t *testing.T, path, content string) {
	t.Helper()
	if err := os.MkdirAll(filepath.Dir(path), 0o750); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path, []byte(content), 0o600); err != nil {
		t.Fatal(err)
	}
}

// looseConfig is the run configuration of a loose main.rs with the text.
func looseConfig(t *testing.T, text string) domain.RunConfiguration {
	t.Helper()
	file := filepath.Join(t.TempDir(), "main.rs")
	writeFile(t, file, text)
	return domain.NewFileRunConfiguration(domain.CodeLanguageRust, file, nil)
}

// crateConfig is the run configuration of src/main.rs of a crate (in the workspace when given).
func crateConfig(t *testing.T, folder, text string) domain.RunConfiguration {
	t.Helper()
	writeFile(t, filepath.Join(folder, "Cargo.toml"), manifest)
	file := filepath.Join(folder, "src", "main.rs")
	writeFile(t, file, text)
	return domain.NewFileRunConfiguration(domain.CodeLanguageRust, file, nil)
}

// lintNames returns the codes of the diagnostics of the JSON lines, and fails on a line that is
// not JSON. cargo wraps each in {"reason":"compiler-message","message":{...}}, rustc does not.
func lintNames(t *testing.T, output string) []string {
	t.Helper()
	names := []string{}
	scanner := bufio.NewScanner(strings.NewReader(output))
	scanner.Buffer(nil, 1<<20)
	for scanner.Scan() {
		var line struct {
			Message json.RawMessage        `json:"message"` // an object in cargo's lines, text in rustc's
			Code    *struct{ Code string } `json:"code"`
		}
		if err := json.Unmarshal(scanner.Bytes(), &line); err != nil {
			t.Fatalf("not a JSON line: %q", scanner.Text())
		}
		var inner struct {
			Code *struct{ Code string } `json:"code"`
		}
		if json.Unmarshal(line.Message, &inner) == nil && inner.Code != nil {
			names = append(names, inner.Code.Code)
		} else if line.Code != nil {
			names = append(names, line.Code.Code)
		}
	}
	return names
}

func contains(names []string, want string) bool {
	for _, name := range names {
		if name == want {
			return true
		}
	}
	return false
}

func TestLooseFileGetsClippyAdviceAsJSONLines(t *testing.T) {
	output, err := realTool(t).Check(context.Background(), looseConfig(t, loopWithIndex))
	if err != nil || !contains(lintNames(t, output), "clippy::needless_range_loop") {
		t.Errorf("output = %q, %v", output, err)
	}
}

func TestCrateGetsClippyAdviceAsJSONLines(t *testing.T) {
	output, err := realTool(t).Check(context.Background(), crateConfig(t, t.TempDir(), loopWithIndex))
	if err != nil || !contains(lintNames(t, output), "clippy::needless_range_loop") {
		t.Errorf("output = %q, %v", output, err)
	}
	if strings.Contains(output, `"reason":"compiler-artifact"`) || strings.Contains(output, "build-finished") {
		t.Errorf("only diagnostics belong in the output: %q", output)
	}
}

func TestMemberOfAWorkspaceIsCheckedWithItsName(t *testing.T) {
	workspace := t.TempDir()
	writeFile(t, filepath.Join(workspace, "Cargo.toml"), "[workspace]\nmembers = [\"sample\"]\nresolver = \"2\"\n")
	config := crateConfig(t, filepath.Join(workspace, "sample"), loopWithIndex)
	output, err := realTool(t).Check(context.Background(), config)
	if err != nil || !contains(lintNames(t, output), "clippy::needless_range_loop") {
		t.Errorf("output = %q, %v", output, err)
	}
}

func TestCleanCodeIsEmpty(t *testing.T) {
	tool := realTool(t)
	for name, config := range map[string]domain.RunConfiguration{
		"loose": looseConfig(t, cleanProgram),
		"crate": crateConfig(t, t.TempDir(), cleanProgram),
	} {
		if output, err := tool.Check(context.Background(), config); err != nil || output != "" {
			t.Errorf("%s: output = %q, %v", name, output, err)
		}
	}
}

func TestErrorsAreReportedToo(t *testing.T) {
	output, err := realTool(t).Check(context.Background(), looseConfig(t, "fn main() { let x: i32 = \"a\"; }\n"))
	if err != nil || !contains(lintNames(t, output), "E0308") {
		t.Errorf("output = %q, %v", output, err)
	}
}

func TestWithoutClippyOnlyTheCompilerWarns(t *testing.T) {
	tool := withoutClippy(t)
	if tool.hasClippy() {
		t.Fatal("the test toolchain still has clippy")
	}
	for name, config := range map[string]domain.RunConfiguration{
		"loose": looseConfig(t, unusedAndLoop),
		"crate": crateConfig(t, t.TempDir(), unusedAndLoop),
	} {
		output, err := tool.Check(context.Background(), config)
		names := lintNames(t, output)
		if err != nil || !strings.Contains(output, "unused_variables") && !contains(names, "unused_variables") {
			t.Errorf("%s: output = %q, %v", name, output, err)
		}
		if contains(names, "clippy::needless_range_loop") {
			t.Errorf("%s: clippy advice without clippy", name)
		}
	}
}

func TestMissingRustcIsMissingTool(t *testing.T) {
	empty := t.TempDir()
	tool := New(rust.NewLocator(rust.Options{AppDir: empty, BaseEnvironment: []string{"PATH=", "CARGO_HOME=" + empty}}))
	_, err := tool.Check(context.Background(), looseConfig(t, cleanProgram))
	if !errors.Is(err, app.ErrToolNotFound) {
		t.Errorf("err = %v", err)
	}
}

func TestAFolderThatIsGoneIsNotChecked(t *testing.T) {
	config := looseConfig(t, cleanProgram)
	if err := os.RemoveAll(config.WorkingDir); err != nil {
		t.Fatal(err)
	}
	if output, err := realTool(t).Check(context.Background(), config); err != nil || output != "" {
		t.Errorf("output = %q, %v", output, err)
	}
}
