package vcpkg

import (
	"path/filepath"
	"testing"
)

// A seed outside the application folder is found through VIZCACHA_TEST_VCPKG_SEED, for development
// and the E2E tests; the bundled folder and the explicit option come first.
func TestSeedComesFromTheEnvironmentWhenNothingIsBundled(t *testing.T) {
	locator := NewLocator(LocatorOptions{
		AppDir: t.TempDir(), BaseEnvironment: []string{SeedEnvironment + "=D:/seed"},
	})
	if got := locator.Seed(); got != "D:/seed" {
		t.Errorf("Seed = %q", got)
	}
	explicit := NewLocator(LocatorOptions{
		AppDir: t.TempDir(), SeedDir: "E:/own", BaseEnvironment: []string{SeedEnvironment + "=D:/seed"},
	})
	if got := explicit.Seed(); got != "E:/own" {
		t.Errorf("Seed = %q, want the explicit one", got)
	}
}

// Settings lets the student choose the vcpkg executable; the root is its folder.
func TestRootAcceptsTheChosenExecutable(t *testing.T) {
	root := t.TempDir()
	mustWrite(t, filepath.Join(root, ".vcpkg-root"), "")
	mustWrite(t, filepath.Join(root, "vcpkg"), "exe")
	locator := NewLocator(LocatorOptions{
		Settings: settingsWith(filepath.Join(root, "vcpkg")), AppDir: t.TempDir(), BaseEnvironment: []string{}, GOOS: "linux",
	})
	if got := locator.Root(); got != root {
		t.Errorf("Root = %q, want %q", got, root)
	}
}
