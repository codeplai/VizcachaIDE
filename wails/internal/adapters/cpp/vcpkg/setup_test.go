package vcpkg

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestRootPrefersTheConfiguredThenBundledThenEnvironment(t *testing.T) {
	configured, bundledApp, fromEnv := fakeRoot(t, "linux"), t.TempDir(), fakeRoot(t, "linux")
	bundled := filepath.Join(bundledApp, "toolchain", "cpp", "vcpkg")
	if err := os.MkdirAll(bundled, 0o755); err != nil {
		t.Fatal(err)
	}
	for _, name := range []string{".vcpkg-root", "vcpkg"} {
		_ = os.WriteFile(filepath.Join(bundled, name), nil, 0o644)
	}
	environment := []string{"VCPKG_ROOT=" + fromEnv}
	options := LocatorOptions{AppDir: bundledApp, BaseEnvironment: environment, GOOS: "linux"}

	if got := NewLocator(options).Root(); got != bundled {
		t.Errorf("bundled: %q", got)
	}
	options.Settings = settingsWith(configured)
	if got := NewLocator(options).Root(); got != configured {
		t.Errorf("configured: %q", got)
	}
	options.Settings = settingsWith(filepath.Join(configured, "nope"))
	options.AppDir = t.TempDir()
	if got := NewLocator(options).Root(); got != fromEnv {
		t.Errorf("environment: %q", got)
	}
	options.BaseEnvironment = nil
	options.Settings = nil
	if got := NewLocator(LocatorOptions{AppDir: t.TempDir(), BaseEnvironment: []string{}, GOOS: "linux"}).Root(); got != "" {
		t.Errorf("none: %q", got)
	}
}

func TestRootNeedsTheMarkerAndTheExecutable(t *testing.T) {
	root := fakeRoot(t, "linux")
	_ = os.Remove(filepath.Join(root, ".vcpkg-root"))
	locator := NewLocator(LocatorOptions{AppDir: t.TempDir(), BaseEnvironment: []string{"VCPKG_ROOT=" + root}, GOOS: "linux"})
	if got := locator.Root(); got != "" {
		t.Errorf("root without .vcpkg-root accepted: %q", got)
	}
}

func TestTripletByPlatform(t *testing.T) {
	cases := []struct{ goos, goarch, want string }{
		{"windows", "amd64", "x64-mingw-static"},
		{"linux", "amd64", "x64-linux"},
		{"linux", "arm64", "arm64-linux"},
		{"darwin", "arm64", "arm64-osx"},
		{"darwin", "amd64", "x64-osx"},
	}
	for _, c := range cases {
		locator := NewLocator(LocatorOptions{GOOS: c.goos, GOARCH: c.goarch})
		if got := locator.Triplet(); got != c.want {
			t.Errorf("%s/%s: %q, want %q", c.goos, c.goarch, got, c.want)
		}
	}
}

func newSetup(t *testing.T, goos string) (*Setup, string, string) {
	t.Helper()
	root, cache := fakeRoot(t, goos), t.TempDir()
	bin := filepath.Join(t.TempDir(), "llvm", "bin")
	locator := NewLocator(LocatorOptions{
		AppDir: t.TempDir(), BaseEnvironment: []string{"VCPKG_ROOT=" + root}, CacheDir: cache,
		CompilerBin: func() string { return bin }, GOOS: goos, GOARCH: "amd64",
	})
	return NewSetup(locator), root, cache
}

func TestConfigureArgsPointVcpkgAtTheCache(t *testing.T) {
	setup, root, cache := newSetup(t, "windows")
	joined := strings.Join(setup.ConfigureArgs("C:/Ñandú"), "\n")
	slashed := strings.ReplaceAll(root, `\`, "/")
	for _, want := range []string{
		"-DCMAKE_TOOLCHAIN_FILE=" + slashed + "/scripts/buildsystems/vcpkg.cmake",
		"-DVCPKG_TARGET_TRIPLET=x64-mingw-static",
		"-DVCPKG_HOST_TRIPLET=x64-mingw-static",
		"--x-buildtrees-root=" + strings.ReplaceAll(filepath.Join(cache, "vcpkg", "buildtrees"), `\`, "/") +
			";--x-packages-root=" + strings.ReplaceAll(filepath.Join(cache, "vcpkg", "packages"), `\`, "/"),
	} {
		if !strings.Contains(joined, want) {
			t.Errorf("missing %q in\n%s", want, joined)
		}
	}
}

func TestWithoutVcpkgThereAreNoArgumentsAndTheEnvironmentIsUntouched(t *testing.T) {
	locator := NewLocator(LocatorOptions{AppDir: t.TempDir(), BaseEnvironment: []string{}, GOOS: "linux"})
	setup := NewSetup(locator)
	if args := setup.ConfigureArgs("x"); args != nil {
		t.Errorf("args = %v", args)
	}
	if env := setup.Environment([]string{"A=1"}); len(env) != 1 {
		t.Errorf("env = %v", env)
	}
}

func TestEnvironmentSetsVcpkgVariablesAndPutsTheCompilerFirstOnWindows(t *testing.T) {
	setup, root, cache := newSetup(t, "windows")
	environment := setup.Environment([]string{`Path=C:\Windows`, "VCPKG_DISABLE_METRICS=0", "KEEP=1"})
	got := map[string]string{}
	for _, entry := range environment {
		name, value, _ := strings.Cut(entry, "=")
		got[name] = value
	}
	if got["VCPKG_ROOT"] != root || got["VCPKG_DISABLE_METRICS"] != "1" || got["KEEP"] != "1" {
		t.Errorf("environment = %v", got)
	}
	if got["VCPKG_DOWNLOADS"] != filepath.Join(cache, "vcpkg", "downloads") {
		t.Errorf("downloads = %q", got["VCPKG_DOWNLOADS"])
	}
	if info, err := os.Stat(got["VCPKG_DEFAULT_BINARY_CACHE"]); err != nil || !info.IsDir() {
		t.Errorf("the binary cache folder must exist: %v", err)
	}
	if !strings.Contains(got["PATH"], filepath.Join("llvm", "bin")+string(os.PathListSeparator)+`C:\Windows`) {
		t.Errorf("PATH = %q", got["PATH"])
	}
	if _, duplicate := got["Path"]; duplicate {
		t.Error("Path was kept next to PATH")
	}
}

func TestPrepareCopiesTheSeedWithoutOverwriting(t *testing.T) {
	setup, _, cache := newSetup(t, "windows")
	seed := setup.locator.Seed()
	if seed != "" {
		t.Fatalf("no bundled seed expected, got %q", seed)
	}
	appDir := t.TempDir()
	seed = filepath.Join(appDir, "toolchain", "cpp", "vcpkg-seed")
	mustWrite(t, filepath.Join(seed, "PowerShell.zip"), "seed zip")
	mustWrite(t, filepath.Join(seed, "tools", "7zr-26", "7zr.exe"), "seed 7zr")
	mustWrite(t, filepath.Join(seed, "existing.7z"), "from the seed")
	downloads := filepath.Join(cache, "vcpkg", "downloads")
	mustWrite(t, filepath.Join(downloads, "existing.7z"), "already downloaded")

	setup.locator.options.AppDir = appDir
	for range 2 { // the second call must change nothing
		if err := setup.Prepare(context.Background()); err != nil {
			t.Fatal(err)
		}
	}
	checks := map[string]string{
		"PowerShell.zip":       "seed zip",
		"tools/7zr-26/7zr.exe": "seed 7zr",
		"existing.7z":          "already downloaded",
	}
	for name, want := range checks {
		data, err := os.ReadFile(filepath.Join(downloads, filepath.FromSlash(name)))
		if err != nil || string(data) != want {
			t.Errorf("%s = %q (%v), want %q", name, data, err, want)
		}
	}
	if _, err := os.Stat(filepath.Join(downloads, "PowerShell.zip.part")); err == nil {
		t.Error("a .part file was left behind")
	}
}

func TestPrepareWithoutSeedDoesNothing(t *testing.T) {
	setup, _, cache := newSetup(t, "linux")
	if err := setup.Prepare(context.Background()); err != nil {
		t.Fatal(err)
	}
	if _, err := os.Stat(filepath.Join(cache, "vcpkg", "downloads")); err == nil {
		t.Error("downloads created without a seed")
	}
}
