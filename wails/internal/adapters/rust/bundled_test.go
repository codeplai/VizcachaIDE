package rust_test

import (
	"context"
	"os"
	"path/filepath"
	"slices"
	"strings"
	"testing"

	"github.com/codeplai/VizcachaIDE/wails/internal/adapters/rust"
	"github.com/codeplai/VizcachaIDE/wails/internal/adapters/rust/rusttest"
	"github.com/codeplai/VizcachaIDE/wails/internal/domain"
)

type settingsStub struct{ paths map[string]string }

func (s settingsStub) Load() (domain.Settings, error) {
	return domain.Settings{ToolPaths: s.paths}, nil
}
func (s settingsStub) Save(domain.Settings) error { return nil }

// fakeBin creates empty executable-looking files in dir and returns the path of the first.
func fakeBin(t *testing.T, dir string, names ...string) string {
	t.Helper()
	if err := os.MkdirAll(dir, 0o755); err != nil {
		t.Fatal(err)
	}
	for _, name := range names {
		if err := os.WriteFile(filepath.Join(dir, rusttest.Exe(name)), []byte("x"), 0o755); err != nil {
			t.Fatal(err)
		}
	}
	return filepath.Join(dir, rusttest.Exe(names[0]))
}

func bundledApp(t *testing.T) (app string) {
	t.Helper()
	app = t.TempDir()
	fakeBin(t, filepath.Join(app, "toolchain", "rust", "bin"),
		"rustc", "cargo", "rustfmt", "rust-analyzer", "cargo-clippy", "clippy-driver")
	return app
}

func TestBundledToolchainWinsOverRustup(t *testing.T) {
	app := bundledApp(t)
	cargoHome := t.TempDir()
	fakeBin(t, filepath.Join(cargoHome, "bin"), "rustc", "cargo", "rustfmt", "rust-analyzer", "cargo-clippy")
	locator := rust.NewLocator(rust.Options{AppDir: app, BaseEnvironment: []string{"CARGO_HOME=" + cargoHome, "PATH="}})
	for _, id := range []string{rust.ToolRustc, rust.ToolCargo, rust.ToolRustfmt, rust.ToolRustAnalyzer, rust.ToolClippy} {
		status := locator.Tool(id)
		if status.Source != domain.ToolBundled || !strings.HasPrefix(status.Path, filepath.Join(app, "toolchain", "rust", "bin")) {
			t.Errorf("%s = %+v, want the bundled one", id, status)
		}
	}
	if !locator.Bundled() {
		t.Error("Bundled() = false")
	}
	if got := locator.Tool(rust.ToolClippy).Path; filepath.Base(got) != rusttest.Exe("cargo-clippy") {
		t.Errorf("clippy is driven by cargo-clippy, got %s", got)
	}
}

func TestConfiguredToolWinsOverBundled(t *testing.T) {
	app := bundledApp(t)
	configured := fakeBin(t, filepath.Join(t.TempDir(), "mine"), "rustc")
	locator := rust.NewLocator(rust.Options{
		AppDir: app, BaseEnvironment: []string{"PATH="},
		Settings: settingsStub{paths: map[string]string{rust.ToolRustc: configured}},
	})
	if got := locator.Tool(rust.ToolRustc); got.Source != domain.ToolConfigured || got.Path != configured {
		t.Errorf("rustc = %+v", got)
	}
	if got := locator.Tool(rust.ToolCargo); got.Source != domain.ToolBundled {
		t.Errorf("cargo = %+v", got)
	}
}

func TestRustupIsUsedWhenNothingIsBundled(t *testing.T) {
	cargoHome := t.TempDir()
	proxy := fakeBin(t, filepath.Join(cargoHome, "bin"), "rustc", "cargo")
	locator := rust.NewLocator(rust.Options{AppDir: t.TempDir(), BaseEnvironment: []string{"CARGO_HOME=" + cargoHome, "PATH="}})
	if got := locator.Tool(rust.ToolRustc); got.Source != domain.ToolOnPath || got.Path != proxy {
		t.Errorf("rustc = %+v", got)
	}
	if locator.Bundled() {
		t.Error("Bundled() = true without toolchain/rust")
	}
	if vars := locator.ShellVariables(context.Background()); vars != nil {
		t.Errorf("rustup needs no CARGO_HOME redirect, got %v", vars)
	}
}

func TestBundledToolchainGetsAWritableCargoHome(t *testing.T) {
	app, home := bundledApp(t), filepath.Join(t.TempDir(), "cache", "cargo")
	locator := rust.NewLocator(rust.Options{AppDir: app, CargoHome: home, BaseEnvironment: []string{"PATH=" + os.Getenv("PATH")}})
	want := "CARGO_HOME=" + home
	if vars := locator.ShellVariables(context.Background()); len(vars) != 1 || vars[0] != want {
		t.Errorf("ShellVariables = %v, want [%s]", vars, want)
	}
	env := locator.Environment()
	if !slices.Contains(env, want) {
		t.Errorf("%s missing from the environment", want)
	}
	first := strings.SplitN(pathEntry(env), string(os.PathListSeparator), 2)[0]
	if first != filepath.Join(app, "toolchain", "rust", "bin") {
		t.Errorf("PATH starts with %q, want the bundled bin", first)
	}
	for _, entry := range env {
		if strings.HasPrefix(entry, "CARGO_TARGET_DIR=") {
			t.Errorf("target stays per project, got %s", entry)
		}
	}
}

func TestAnExplicitCargoHomeIsKeptWithTheBundledToolchain(t *testing.T) {
	locator := rust.NewLocator(rust.Options{
		AppDir: bundledApp(t), CargoHome: t.TempDir(), BaseEnvironment: []string{"CARGO_HOME=/mine", "PATH="},
	})
	if vars := locator.ShellVariables(context.Background()); vars != nil {
		t.Errorf("ShellVariables = %v, want none", vars)
	}
	if env := locator.Environment(); !slices.Contains(env, "CARGO_HOME=/mine") {
		t.Errorf("CARGO_HOME of the user lost: %v", env)
	}
}

func TestBundledSysrootIsTheToolchainFolder(t *testing.T) {
	app := bundledApp(t)
	sysroot := filepath.Join(app, "toolchain", "rust")
	locator := rust.NewLocator(rust.Options{
		AppDir: app, BaseEnvironment: []string{"PATH="},
		Probe: func(_ context.Context, path string, args ...string) (string, error) {
			if path != filepath.Join(sysroot, "bin", rusttest.Exe("rustc")) {
				t.Errorf("probed %s, want the bundled rustc", path)
			}
			if args[0] == "--print" {
				return sysroot + "\n", nil
			}
			return verboseStable, nil
		},
	})
	found, err := locator.Toolchain(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	wantFormatters := filepath.Join(sysroot, "lib", "rustlib", "etc")
	if found.Sysroot != sysroot || found.FormattersDir() != wantFormatters {
		t.Errorf("sysroot %q, formatters %q", found.Sysroot, found.FormattersDir())
	}
	wantGCC := filepath.Join(sysroot, "lib", "rustlib", "x86_64-pc-windows-gnu", "bin", "self-contained", "x86_64-w64-mingw32-gcc.exe")
	if found.SelfContainedGCC() != wantGCC {
		t.Errorf("self-contained gcc = %q, want %q", found.SelfContainedGCC(), wantGCC)
	}
}

func pathEntry(env []string) string {
	for _, entry := range env {
		if key, value, ok := strings.Cut(entry, "="); ok && strings.EqualFold(key, "PATH") {
			return value
		}
	}
	return ""
}
