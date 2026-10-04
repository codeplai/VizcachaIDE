package rustfmt

import (
	"errors"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"testing"

	"github.com/codeplai/VizcachaIDE/wails/internal/adapters/rust"
	"github.com/codeplai/VizcachaIDE/wails/internal/adapters/rust/rusttest"
	"github.com/codeplai/VizcachaIDE/wails/internal/app"
)

const (
	crooked  = "fn main(){\nlet x=1;\n        if x>0 {\n  println!(\"{}\",x);\n}\n}\n"
	standard = "fn main() {\n    let x = 1;\n    if x > 0 {\n        println!(\"{}\", x);\n    }\n}\n"
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

func TestFormatsBadlyIndentedCodeWithFourSpaces(t *testing.T) {
	got, err := realTool(t).Format(filepath.Join(t.TempDir(), "main.rs"), crooked)
	if err != nil || got != standard {
		t.Errorf("got %q, %v", got, err)
	}
}

func TestRustfmtTomlOfTheFolderWins(t *testing.T) {
	folder := t.TempDir()
	writeFile(t, filepath.Join(folder, "rustfmt.toml"), "tab_spaces = 2\n")
	got, err := realTool(t).Format(filepath.Join(folder, "main.rs"), crooked)
	want := strings.ReplaceAll(standard, "    ", "  ")
	if err != nil || got != want {
		t.Errorf("got %q, %v, want %q", got, err, want)
	}
}

func TestDotRustfmtTomlAboveTheFileWins(t *testing.T) {
	folder := t.TempDir()
	writeFile(t, filepath.Join(folder, ".rustfmt.toml"), "tab_spaces = 2\n")
	file := filepath.Join(folder, "src", "main.rs")
	writeFile(t, file, crooked)
	got, err := realTool(t).Format(file, crooked)
	if err != nil || !strings.Contains(got, "\n  let x = 1;") {
		t.Errorf("got %q, %v", got, err)
	}
}

func TestSyntaxErrorIsErrFormatWithTheLine(t *testing.T) {
	_, err := realTool(t).Format(filepath.Join(t.TempDir(), "main.rs"), "fn main() {\n    let x = ;\n}\n")
	if !errors.Is(err, app.ErrFormat) {
		t.Fatalf("err = %v", err)
	}
	// The frontend reads "<file>:<line>:<column>: message".
	if !regexp.MustCompile(`main\.rs:2:\d+: expected expression`).MatchString(err.Error()) {
		t.Errorf("err = %q", err)
	}
}

// 2024 reads "gen" as a keyword; 2015 does not: the crate's edition decides.
func TestTheEditionOfTheCrateIsUsed(t *testing.T) {
	folder := t.TempDir()
	writeFile(t, filepath.Join(folder, "Cargo.toml"), "[package]\nname = \"c\"\nversion = \"0.1.0\"\nedition = \"2015\"\n")
	code := "fn main(){let r#try=1;let async=2;}\n"
	got, err := realTool(t).Format(filepath.Join(folder, "src", "main.rs"), code)
	if err != nil || !strings.Contains(got, "let async = 2;") {
		t.Errorf("2015 crate: got %q, %v", got, err)
	}
	if _, err := realTool(t).Format(filepath.Join(t.TempDir(), "main.rs"), code); !errors.Is(err, app.ErrFormat) {
		t.Errorf("a loose file is 2024 and async is a keyword: err = %v", err)
	}
}

func TestAnUntitledFileWithoutFolderIsFormatted(t *testing.T) {
	got, err := realTool(t).Format(filepath.Join(t.TempDir(), "gone", "untitled.rs"), crooked)
	if err != nil || got != standard {
		t.Errorf("got %q, %v", got, err)
	}
}

func TestMissingRustfmtIsMissingTool(t *testing.T) {
	empty := t.TempDir()
	tool := New(rust.NewLocator(rust.Options{AppDir: empty, BaseEnvironment: []string{"PATH=", "CARGO_HOME=" + empty}}))
	_, err := tool.Format(filepath.Join(empty, "main.rs"), crooked)
	if !errors.Is(err, app.ErrToolNotFound) || err.Error() != app.MissingTool("rustfmt").Error() {
		t.Errorf("err = %v", err)
	}
}
