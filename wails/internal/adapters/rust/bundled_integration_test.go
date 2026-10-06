package rust_test

import (
	"context"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	"github.com/codeplai/VizcachaIDE/wails/internal/adapters/rust"
	"github.com/codeplai/VizcachaIDE/wails/internal/domain"
)

// bundleVariable names a folder that contains toolchain/rust, as packaging/fetch_rust.py builds it.
const bundleVariable = "VIZCACHA_TEST_RUST_BUNDLE"

func bundleLocator(t *testing.T) (*rust.Locator, string) {
	t.Helper()
	dir := os.Getenv(bundleVariable)
	if dir == "" {
		t.Skipf("set %s to a folder with toolchain/rust (python packaging/fetch_rust.py --dest <dir>)", bundleVariable)
	}
	base := []string{}
	for _, entry := range os.Environ() { // the machine's own rustup must not leak in
		if name, _, _ := strings.Cut(entry, "="); !strings.EqualFold(name, "CARGO_HOME") && !strings.HasPrefix(strings.ToUpper(name), "RUSTUP_") {
			base = append(base, entry)
		}
	}
	cache := filepath.Join(t.TempDir(), "cargo home")
	return rust.NewLocator(rust.Options{AppDir: dir, CargoHome: cache, BaseEnvironment: base}), cache
}

func runIn(t *testing.T, env []string, dir, program string, args ...string) string {
	t.Helper()
	cmd := exec.Command(program, args...)
	cmd.Dir, cmd.Env = dir, env
	out, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("%s %v: %v\n%s", program, args, err, out)
	}
	return string(out)
}

func TestBundledToolchainCompilesAndRuns(t *testing.T) {
	locator, _ := bundleLocator(t)
	found, err := locator.Toolchain(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	if rustc := locator.Tool(rust.ToolRustc); rustc.Source != domain.ToolBundled {
		t.Fatalf("rustc = %+v, want bundled", rustc)
	}
	if want := filepath.Join(os.Getenv(bundleVariable), "toolchain", "rust"); !strings.EqualFold(filepath.Clean(found.Sysroot), filepath.Clean(want)) {
		t.Errorf("sysroot %q, want %q", found.Sysroot, want)
	}
	if _, err := os.Stat(filepath.Join(found.FormattersDir(), "lldb_lookup.py")); err != nil {
		t.Errorf("no LLDB formatters in the bundled sysroot: %v", err)
	}
	if !found.SelfContainedLinker(func(p string) bool { _, err := os.Stat(p); return err == nil }) {
		t.Errorf("no self-contained linker in %s", found.Sysroot)
	}
	folder := filepath.Join(t.TempDir(), "mis programas ñandú")
	if err := os.MkdirAll(folder, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(folder, "main.rs"), []byte("fn main() { println!(\"hola\"); }\n"), 0o600); err != nil {
		t.Fatal(err)
	}
	env := locator.Environment()
	runIn(t, env, folder, locator.Tool(rust.ToolRustc).Path, "main.rs", "-o", "main"+filepath.Ext(locator.Tool(rust.ToolRustc).Path))
	if out := runIn(t, env, folder, filepath.Join(folder, "main"+filepath.Ext(locator.Tool(rust.ToolRustc).Path))); !strings.Contains(out, "hola") {
		t.Errorf("program printed %q", out)
	}
}

func TestBundledCargoBuildsAProjectWithAWritableCargoHome(t *testing.T) {
	locator, cache := bundleLocator(t)
	root := filepath.Join(t.TempDir(), "proyecto")
	env := locator.Environment()
	cargo := locator.Tool(rust.ToolCargo).Path
	runIn(t, env, filepath.Dir(root), cargo, "new", "--vcs", "none", "proyecto")
	out := runIn(t, env, root, cargo, "run", "--quiet")
	if !strings.Contains(out, "Hello, world!") {
		t.Errorf("cargo run printed %q", out)
	}
	if _, err := os.Stat(filepath.Join(root, "target")); err != nil {
		t.Errorf("target/ must stay in the project: %v", err)
	}
	if clippy := runIn(t, env, root, cargo, "clippy", "--quiet"); strings.Contains(clippy, "error") {
		t.Errorf("clippy: %s", clippy)
	}
	if _, err := os.Stat(cache); err != nil {
		t.Errorf("CARGO_HOME %s was not used: %v", cache, err)
	}
}
