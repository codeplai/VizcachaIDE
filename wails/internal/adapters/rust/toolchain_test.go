package rust_test

import (
	"context"
	"errors"
	"os"
	"path/filepath"
	"slices"
	"strings"
	"testing"

	"github.com/codeplai/VizcachaIDE/wails/internal/adapters/rust"
	"github.com/codeplai/VizcachaIDE/wails/internal/adapters/rust/rusttest"
	"github.com/codeplai/VizcachaIDE/wails/internal/app"
	"github.com/codeplai/VizcachaIDE/wails/internal/domain"
)

const verboseStable = "rustc 1.99.0 (b940084d7 2026-09-28)\nbinary: rustc\nhost: x86_64-pc-windows-gnu\nrelease: 1.99.0\nLLVM version: 21.1.2\n"

// fakeRustc is a rustc found in a fake CARGO_HOME/bin that answers with the given output.
func fakeRustc(t *testing.T, verbose string, fail bool) *rust.Locator {
	t.Helper()
	cargoHome := t.TempDir()
	bin := filepath.Join(cargoHome, "bin")
	if err := os.MkdirAll(bin, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(bin, rusttest.Exe("rustc")), nil, 0o755); err != nil {
		t.Fatal(err)
	}
	return rust.NewLocator(rust.Options{
		AppDir: t.TempDir(), BaseEnvironment: []string{"CARGO_HOME=" + cargoHome, "PATH="},
		Probe: func(_ context.Context, _ string, args ...string) (string, error) {
			if args[0] == "--print" {
				return "C:/rust/sysroot\n", nil
			}
			if fail {
				return verbose, errors.New("exit status 1")
			}
			return verbose, nil
		},
	})
}

func TestToolchainIsReadFromRustcVerbose(t *testing.T) {
	found, err := fakeRustc(t, verboseStable, false).Toolchain(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	if found.Version != "1.99.0" || found.Channel != rust.ChannelStable || found.Host != "x86_64-pc-windows-gnu" ||
		found.Sysroot != "C:/rust/sysroot" || found.Edition() != "2024" {
		t.Errorf("toolchain = %+v", found)
	}
	if advice := found.Advice("windows"); len(advice) != 0 {
		t.Errorf("advice = %v", advice)
	}
}

func TestAdviceForMsvcBetaAndOldToolchains(t *testing.T) {
	msvc := rust.Toolchain{Version: "1.99.0", Channel: rust.ChannelStable, Host: "x86_64-pc-windows-msvc"}
	if !slices.Contains(msvc.Advice("windows"), "errors.rustMsvcHost") || len(msvc.Advice("linux")) != 0 {
		t.Errorf("msvc advice = %v", msvc.Advice("windows"))
	}
	beta := rust.Toolchain{Version: "1.100.0", Channel: rust.ChannelBeta, Host: "x86_64-unknown-linux-gnu"}
	if !slices.Contains(beta.Advice("linux"), "errors.rustNotStable") {
		t.Errorf("beta advice = %v", beta.Advice("linux"))
	}
	old := rust.Toolchain{Version: "1.80.0", Channel: rust.ChannelStable, Host: "aarch64-apple-darwin"}
	if !slices.Contains(old.Advice("darwin"), "errors.rustTooOld") || old.Edition() != "2021" {
		t.Errorf("old = %v, edition %s", old.Advice("darwin"), old.Edition())
	}
}

func TestRustupWithoutToolchainIsReported(t *testing.T) {
	answer := "error: rustup could not choose a version of rustc to run, because one wasn't specified explicitly, and no default is configured.\nhelp: run 'rustup default stable' to download the latest stable release of Rust and set it as your default toolchain.\n"
	_, err := fakeRustc(t, answer, true).Toolchain(context.Background())
	if !errors.Is(err, rust.ErrNoToolchain) || !errors.Is(err, app.ErrToolNotFound) {
		t.Errorf("err = %v", err)
	}
}

func TestMissingRustcIsTheRustcTool(t *testing.T) {
	locator := rust.NewLocator(rust.Options{AppDir: t.TempDir(), BaseEnvironment: []string{"CARGO_HOME=" + t.TempDir(), "PATH="}})
	if _, err := locator.Toolchain(context.Background()); !errors.Is(err, app.ErrToolNotFound) {
		t.Errorf("err = %v", err)
	}
	if status := locator.Tool(rust.ToolCargo); status.Source != domain.ToolMissing || status.CodeLanguage != domain.CodeLanguageRust {
		t.Errorf("cargo = %+v", status)
	}
}

func TestEnvironmentAsksForPlainOutputAndTheBacktrace(t *testing.T) {
	env := rust.Environment([]string{"RUST_BACKTRACE=0", "Path=C:/x"})
	joined := strings.Join(env, "\n")
	for _, want := range []string{"RUST_BACKTRACE=1", "CARGO_TERM_COLOR=never", "NO_COLOR=1", "CARGO_INCREMENTAL=0", "Path=C:/x"} {
		if !strings.Contains(joined, want) {
			t.Errorf("%s missing from %v", want, env)
		}
	}
	if strings.Contains(joined, "RUST_BACKTRACE=0") {
		t.Errorf("RUST_BACKTRACE=0 kept: %v", env)
	}
}

func TestRealToolchainLinksWithItsOwnMinGW(t *testing.T) {
	env := rusttest.Environment(t)
	locator := rust.NewLocator(rust.Options{AppDir: t.TempDir(), BaseEnvironment: env})
	found, err := locator.Toolchain(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	if found.Channel != rust.ChannelStable || found.Sysroot == "" {
		t.Errorf("toolchain = %+v", found)
	}
	exists := func(path string) bool { _, err := os.Stat(path); return err == nil }
	if strings.HasSuffix(found.Host, "-windows-gnu") && !found.SelfContainedLinker(exists) {
		t.Errorf("the GNU toolchain has no self-contained linker in %s", found.Sysroot)
	}
	if _, err := os.Stat(filepath.Join(found.FormattersDir(), "lldb_lookup.py")); err != nil {
		t.Errorf("no LLDB formatters: %v", err)
	}
	for _, id := range []string{rust.ToolCargo, rust.ToolRustAnalyzer, rust.ToolClippy, rust.ToolRustfmt} {
		if status := locator.Tool(id); status.Source == domain.ToolMissing {
			t.Errorf("%s not found", id)
		}
	}
}

func TestWindowsGnuGetsLLVMDlltoolAndItsOwnLinker(t *testing.T) {
	llvm := filepath.Join(t.TempDir(), "llvm mingw", "bin")
	if err := os.MkdirAll(llvm, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(llvm, "llvm-dlltool.exe"), nil, 0o755); err != nil {
		t.Fatal(err)
	}
	locator := rust.NewLocator(rust.Options{AppDir: t.TempDir(), BaseEnvironment: []string{"PATH=" + llvm}})
	gnu := rust.Toolchain{Host: "x86_64-pc-windows-gnu", Sysroot: t.TempDir()}
	if got := locator.Dlltool(gnu); got != filepath.Join(llvm, "llvm-dlltool.exe") {
		t.Errorf("dlltool = %q", got)
	}
	if got := locator.Dlltool(rust.Toolchain{Host: "x86_64-unknown-linux-gnu"}); got != "" {
		t.Errorf("linux dlltool = %q, want none", got)
	}
}
