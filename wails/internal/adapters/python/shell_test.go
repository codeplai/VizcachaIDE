package python

import (
	"context"
	"path/filepath"
	"runtime"
	"slices"
	"testing"
)

func TestShellPathsHoldTheConfiguredInterpreterFolderAndItsScripts(t *testing.T) {
	root := t.TempDir()
	configured := touch(t, filepath.Join(root, "custom", exe("python")))
	scripts := filepath.Join(root, "custom", "Scripts")
	touch(t, filepath.Join(scripts, exe("pip")))
	locator := NewLocator(Options{
		Settings:        memorySettings{paths: map[string]string{"python": configured}},
		AppDir:          filepath.Join(root, "app"),
		BaseEnvironment: []string{},
		Probe:           fakeProbe(map[string]string{configured: "3.12.1"}),
		Launcher:        noLauncher,
	})
	paths := locator.ShellPaths(context.Background())
	if len(paths) == 0 || paths[0] != filepath.Join(root, "custom") {
		t.Fatalf("paths = %v, want the folder of the configured python first", paths)
	}
	if runtime.GOOS == "windows" && !slices.Contains(paths, scripts) {
		t.Errorf("paths = %v, want the Scripts folder on Windows", paths)
	}
}

func TestShellPathsAreEmptyWithoutAnInterpreter(t *testing.T) {
	locator := NewLocator(Options{
		AppDir:          t.TempDir(),
		BaseEnvironment: []string{},
		Probe:           fakeProbe(nil),
		Launcher:        noLauncher,
	})
	if paths := locator.ShellPaths(context.Background()); len(paths) != 0 {
		t.Fatalf("paths = %v", paths)
	}
}
