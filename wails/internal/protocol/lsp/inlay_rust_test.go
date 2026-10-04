package lsp

import (
	"context"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
	"time"

	"github.com/codeplai/VizcachaIDE/wails/internal/domain"
)

// analyzerFlavor is a minimal rust-analyzer flavor for this test (the real one is the Rust
// adapter's): the rustup proxy of the toolchain named by the same variables as rusttest.
type analyzerFlavor struct{ cargoHome, rustupHome string }

func (f analyzerFlavor) Command(map[string]string) (string, []string, error) {
	name := "rust-analyzer"
	if runtime.GOOS == "windows" {
		name += ".exe"
	}
	return filepath.Join(f.cargoHome, "bin", name), nil, nil
}
func (analyzerFlavor) RootOf(path string) string  { return filepath.Dir(filepath.Dir(path)) }
func (analyzerFlavor) InitializationOptions() any { return nil }
func (analyzerFlavor) Configuration() any         { return nil }
func (f analyzerFlavor) Environment() map[string]string {
	path := filepath.Join(f.cargoHome, "bin") + string(os.PathListSeparator) + os.Getenv("PATH")
	return map[string]string{"CARGO_HOME": f.cargoHome, "RUSTUP_HOME": f.rustupHome, "PATH": path}
}

func TestRealRustAnalyzerGivesTheTypeOfAVec(t *testing.T) {
	cargoHome, rustupHome := os.Getenv("VIZCACHA_TEST_CARGO_HOME"), os.Getenv("VIZCACHA_TEST_RUSTUP_HOME")
	if cargoHome == "" || rustupHome == "" {
		t.Skip("set VIZCACHA_TEST_CARGO_HOME and VIZCACHA_TEST_RUSTUP_HOME to test with rust-analyzer")
	}
	flavor := analyzerFlavor{cargoHome, rustupHome}
	if executable, _, _ := flavor.Command(nil); executable == "" || !fileExists(executable) {
		t.Skip("rust-analyzer is not installed")
	}
	root := t.TempDir()
	source := "fn main() {\n    let v = vec![1];\n    println!(\"{:?}\", v);\n}\n"
	file := filepath.Join(root, "src", "main.rs")
	cargo := "[package]\nname = \"demo\"\nversion = \"0.1.0\"\nedition = \"2021\"\n"
	if os.MkdirAll(filepath.Dir(file), 0o750) != nil ||
		os.WriteFile(filepath.Join(root, "Cargo.toml"), []byte(cargo), 0o600) != nil ||
		os.WriteFile(file, []byte(source), 0o600) != nil {
		t.Fatal("could not write the test crate")
	}
	server := New(newRecordingSink(), flavor, Options{Name: "rust-analyzer", LanguageID: "rust"})
	t.Cleanup(func() { _ = server.Shutdown(context.Background()) })
	if err := server.OpenDocument(context.Background(), file, source); err != nil {
		t.Fatal(err)
	}
	visible := domain.SourceRange{
		Start: domain.SourceLocation{File: file, Line: 1, Column: 1},
		End:   domain.SourceLocation{File: file, Line: 4, Column: 1},
	}
	deadline := time.Now().Add(3 * time.Minute)
	for time.Now().Before(deadline) {
		hints, _ := server.InlayHints(context.Background(), visible)
		for _, hint := range hints {
			if hint.Kind == domain.InlayHintType && hint.Line == 2 && strings.Contains(hint.Label, "Vec<i32>") {
				return
			}
		}
		time.Sleep(500 * time.Millisecond)
	}
	t.Fatal("no Vec<i32> type hint on line 2")
}

func fileExists(path string) bool {
	info, err := os.Stat(path)
	return err == nil && !info.IsDir()
}
