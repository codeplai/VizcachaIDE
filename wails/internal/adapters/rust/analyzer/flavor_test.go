package analyzer

import (
	"errors"
	"os"
	"path/filepath"
	"reflect"
	"slices"
	"strings"
	"testing"

	"github.com/codeplai/VizcachaIDE/wails/internal/adapters/rust"
	"github.com/codeplai/VizcachaIDE/wails/internal/adapters/rust/rusttest"
	"github.com/codeplai/VizcachaIDE/wails/internal/app"
)

func TestCommandIsRustAnalyzerWithoutArguments(t *testing.T) {
	flavor := NewFlavor(Config{Locator: locatorForTests(t)})
	executable, args, err := flavor.Command(nil)
	if err != nil || filepath.Base(executable) != rusttest.Exe("rust-analyzer") || len(args) != 0 {
		t.Fatalf("command = %q %v, %v", executable, args, err)
	}
}

func TestMissingRustAnalyzerIsMissingTool(t *testing.T) {
	empty := t.TempDir()
	locator := rust.NewLocator(rust.Options{AppDir: empty, BaseEnvironment: []string{"PATH=", "CARGO_HOME=" + empty}})
	_, _, err := NewFlavor(Config{Locator: locator}).Command(nil)
	if !errors.Is(err, app.ErrToolNotFound) || err.Error() != app.MissingTool("rust-analyzer").Error() {
		t.Errorf("err = %v", err)
	}
}

func writeManifest(t *testing.T, folder, content string) {
	t.Helper()
	if err := os.MkdirAll(folder, 0o750); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(folder, "Cargo.toml"), []byte(content), 0o600); err != nil {
		t.Fatal(err)
	}
}

func TestRootIsTheWorkspaceOfTheCrate(t *testing.T) {
	flavor := NewFlavor(Config{})
	workspace := t.TempDir()
	writeManifest(t, workspace, "[workspace]\nmembers = [\"app\"]\n")
	writeManifest(t, filepath.Join(workspace, "app"), "[package]\nname = \"app\"\nversion = \"0.1.0\"\n")
	if got := flavor.RootOf(filepath.Join(workspace, "app", "src", "main.rs")); got != workspace {
		t.Errorf("member: RootOf = %q, want %q", got, workspace)
	}
	alone := t.TempDir()
	writeManifest(t, alone, "[package]\nname = \"alone\"\nversion = \"0.1.0\"\n")
	if got := flavor.RootOf(filepath.Join(alone, "src", "main.rs")); got != alone {
		t.Errorf("crate: RootOf = %q, want %q", got, alone)
	}
}

func TestLooseFileRootIsItsFolderAndIsADetachedFile(t *testing.T) {
	flavor := NewFlavor(Config{})
	folder := t.TempDir()
	file := filepath.Join(folder, "main.rs")
	if got := flavor.RootOf(file); got != folder {
		t.Errorf("RootOf = %q, want %q", got, folder)
	}
	other := filepath.Join(t.TempDir(), "other.rs")
	flavor.RootOf(other)
	flavor.RootOf(file) // the same file again is listed once
	crate := t.TempDir()
	writeManifest(t, crate, "[package]\nname = \"c\"\nversion = \"0.1.0\"\n")
	flavor.RootOf(filepath.Join(crate, "src", "main.rs")) // a crate file is never detached
	options, _ := flavor.Configuration().(map[string]any)
	if detached, _ := options["detachedFiles"].([]string); !slices.Equal(detached, []string{file, other}) {
		t.Errorf("detachedFiles = %v, want [%s %s]", options["detachedFiles"], file, other)
	}
	if _, has := NewFlavor(Config{}).settings()["detachedFiles"]; has {
		t.Error("no loose file, no detached files")
	}
}

func TestInitializationOptionsAreThePlanSettings(t *testing.T) {
	options, _ := NewFlavor(Config{}).InitializationOptions().(map[string]any)
	hints, _ := options["inlayHints"].(map[string]any)
	enabled := func(name string) any { return hints[name].(map[string]any)["enable"] }
	if options["checkOnSave"] != true || options["check"].(map[string]any)["command"] != "clippy" ||
		options["cargo"].(map[string]any)["buildScripts"].(map[string]any)["enable"] != true ||
		options["procMacro"].(map[string]any)["enable"] != true {
		t.Errorf("options = %v", options)
	}
	if enabled("typeHints") != true || enabled("parameterHints") != true || enabled("chainingHints") != false || enabled("closingBraceHints") != false {
		t.Errorf("inlayHints = %v", hints)
	}
	if !reflect.DeepEqual(NewFlavor(Config{}).Configuration(), NewFlavor(Config{}).InitializationOptions()) {
		t.Error("Configuration must be the same settings as the handshake")
	}
}

func TestEnvironmentCarriesTheToolchainAndPlainOutput(t *testing.T) {
	env := NewFlavor(Config{Locator: locatorForTests(t)}).Environment()
	if env["CARGO_HOME"] == "" || env["RUSTUP_HOME"] == "" || env["CARGO_TERM_COLOR"] != "never" {
		t.Errorf("env = %v", env)
	}
	cargoBin := filepath.Join(env["CARGO_HOME"], "bin")
	for name, value := range env {
		if strings.EqualFold(name, "PATH") && !strings.HasPrefix(strings.ToLower(value), strings.ToLower(cargoBin)) {
			t.Errorf("PATH = %q does not start with %q", value, cargoBin)
		}
	}
}
