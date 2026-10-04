package lldbdap

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/codeplai/VizcachaIDE/wails/internal/domain"
)

func TestRealCrateWithoutPythonShowsTheNoticeAndRawValues(t *testing.T) {
	debugger, sink, _ := realDebugger(t)
	debugger.hasPython = func(string) bool { return false }
	crate := t.TempDir()
	if err := os.MkdirAll(filepath.Join(crate, "src"), 0o750); err != nil {
		t.Fatal(err)
	}
	manifest := "[package]\nname = \"demo\"\nversion = \"0.1.0\"\nedition = \"2021\"\n"
	if err := os.WriteFile(filepath.Join(crate, "Cargo.toml"), []byte(manifest), 0o600); err != nil {
		t.Fatal(err)
	}
	main := filepath.Join(crate, "src", "main.rs")
	program := "fn main() {\n    let name = String::from(\"Ana\");\n    println!(\"{}\", name);\n}\n"
	if err := os.WriteFile(main, []byte(program), 0o600); err != nil {
		t.Fatal(err)
	}
	start(t, debugger, main, 3)

	stop := receive(t, sink.stops)
	if stop.Reason != domain.StopBreakpoint || stop.Frames[0].Location.Line != 3 {
		t.Fatalf("stop = %+v", stop)
	}
	if got := variablesOf(stop)["name"].Value; strings.Contains(got, "Ana") {
		t.Errorf("without the formatters String should read raw, got %q", got)
	}
	if !strings.Contains(sink.text(), "errors.rustLldbNoPython") {
		t.Errorf("the no-Python notice is missing: %q", sink.debugEvents())
	}
}
