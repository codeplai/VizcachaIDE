package python

import (
	"context"
	"errors"
	"os"
	"path/filepath"
	"runtime"
	"testing"

	"github.com/codeplai/VizcachaIDE/wails/internal/app"
	"github.com/codeplai/VizcachaIDE/wails/internal/domain"
)

type memorySettings struct{ paths map[string]string }

func (m memorySettings) Load() (domain.Settings, error) {
	settings := domain.DefaultSettings()
	settings.ToolPaths = m.paths
	return settings, nil
}
func (memorySettings) Save(domain.Settings) error { return nil }

// fakeProbe answers with the version listed for a path; unknown paths fail like the Store alias.
func fakeProbe(versions map[string]string) Probe {
	return func(_ context.Context, path string) (string, error) {
		if version, ok := versions[path]; ok {
			return version, nil
		}
		return "", errors.New("not a python")
	}
}

func touch(t *testing.T, path string) string {
	t.Helper()
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path, []byte("x"), 0o755); err != nil {
		t.Fatal(err)
	}
	return path
}

func exe(name string) string {
	if runtime.GOOS == "windows" {
		return name + ".exe"
	}
	return name
}

func noLauncher(context.Context) string { return "" }

func TestFindPrefersConfiguredThenBundledThenVenvThenPath(t *testing.T) {
	root := t.TempDir()
	configured := touch(t, filepath.Join(root, "custom", exe("python")))
	bundled := touch(t, filepath.Join(root, "app", "toolchain", "python", exe("python")))
	if runtime.GOOS != "windows" {
		bundled = touch(t, filepath.Join(root, "app", "toolchain", "python", "bin", "python3"))
	}
	project := filepath.Join(root, "project")
	venvFolder := "bin"
	if runtime.GOOS == "windows" {
		venvFolder = "Scripts"
	}
	venv := touch(t, filepath.Join(project, ".venv", venvFolder, exe("python")))
	onPath := touch(t, filepath.Join(root, "bin", exe("python")))
	versions := map[string]string{configured: "3.12.1", bundled: "3.12.14", venv: "3.11.9", onPath: "3.13.0"}
	locator := func(paths map[string]string) *Locator {
		return NewLocator(Options{
			Settings: memorySettings{paths: paths}, AppDir: filepath.Join(root, "app"),
			BaseEnvironment: []string{"PATH=" + filepath.Join(root, "bin")}, Probe: fakeProbe(versions), Launcher: noLauncher,
		})
	}

	got, err := locator(map[string]string{"python": configured}).Find(context.Background(), project)
	if err != nil || got.Path != configured || got.Source != domain.ToolConfigured {
		t.Fatalf("configured: %+v, %v", got, err)
	}
	got, _ = locator(nil).Find(context.Background(), project)
	if got.Path != bundled || got.Source != domain.ToolBundled || got.Version != "3.12.14" {
		t.Fatalf("bundled: %+v", got)
	}
	delete(versions, bundled) // a broken bundled interpreter is skipped
	got, _ = locator(nil).Find(context.Background(), project)
	if got.Path != venv || !got.Venv {
		t.Fatalf("venv: %+v (want %s)", got, venv)
	}
	got, _ = locator(nil).Find(context.Background(), "")
	if got.Path != onPath || got.Source != domain.ToolOnPath {
		t.Fatalf("PATH: %+v", got)
	}
}

func TestFindRejectsOldPythonAndTheStoreAlias(t *testing.T) {
	root := t.TempDir()
	old := touch(t, filepath.Join(root, "old", exe("python")))
	alias := touch(t, filepath.Join(root, "WindowsApps", exe("python")))
	for _, path := range []string{old, alias} {
		locator := NewLocator(Options{
			BaseEnvironment: []string{"PATH=" + filepath.Dir(path)}, AppDir: root,
			Probe: fakeProbe(map[string]string{old: "3.9.18"}), Launcher: noLauncher,
		})
		if got, err := locator.Find(context.Background(), ""); !errors.Is(err, app.ErrToolNotFound) {
			t.Errorf("%s: got %+v, %v; want ErrToolNotFound", path, got, err)
		}
	}
}

func TestSupportedVersion(t *testing.T) {
	cases := map[string]bool{"3.10.0": true, "3.12.14": true, "4.0": true, "3.9.18": false, "2.7.18": false, "x": false}
	for version, want := range cases {
		if got := supportedVersion(version); got != want {
			t.Errorf("supportedVersion(%q) = %v, want %v", version, got, want)
		}
	}
}

func TestEnvironmentAddsPythonVariables(t *testing.T) {
	env := Environment([]string{"PATH=/bin", "=C:=C:\\x"}, Interpreter{Source: domain.ToolBundled})
	for name, want := range map[string]string{"PYTHONUTF8": "1", "PYTHONUNBUFFERED": "1", "PYTHONDONTWRITEBYTECODE": "1", "PYTHONNOUSERSITE": "1", "PATH": "/bin", "=C:": "C:\\x"} {
		if env[name] != want {
			t.Errorf("%s = %q, want %q", name, env[name], want)
		}
	}
	if _, set := Environment(nil, Interpreter{Source: domain.ToolOnPath})["PYTHONNOUSERSITE"]; set {
		t.Error("PYTHONNOUSERSITE is only for the bundled interpreter")
	}
}
