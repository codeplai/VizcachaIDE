package pylsp

import (
	"context"
	"encoding/json"
	"errors"
	"os/exec"
	"path/filepath"
	"testing"

	"github.com/codeplai/VizcachaIDE/wails/internal/adapters/python"
	"github.com/codeplai/VizcachaIDE/wails/internal/adapters/python/pythontest"
	"github.com/codeplai/VizcachaIDE/wails/internal/app"
	"github.com/codeplai/VizcachaIDE/wails/internal/domain"
)

type pythonAt struct{ path string }

func (p pythonAt) Load() (domain.Settings, error) {
	settings := domain.DefaultSettings()
	settings.ToolPaths = map[string]string{python.ToolPython: p.path}
	return settings, nil
}
func (pythonAt) Save(domain.Settings) error { return nil }

func configAt(path string) Config {
	return Config{Locator: python.NewLocator(python.Options{Settings: pythonAt{path: path}})}
}

func TestConfigurationIsExactlyThePlan(t *testing.T) {
	const want = `{"pylsp":{"plugins":{"pyflakes":{"enabled":true},"pycodestyle":{"enabled":false},
	"mccabe":{"enabled":false},"pylint":{"enabled":false},"jedi_completion":{"include_params":false},
	"jedi_signature_help":{"enabled":true},"jedi_symbols":{"enabled":true}}}}`
	got, err := json.Marshal(Flavor{}.Configuration())
	if err != nil {
		t.Fatal(err)
	}
	var wantValue, gotValue any
	if json.Unmarshal([]byte(want), &wantValue) != nil || json.Unmarshal(got, &gotValue) != nil {
		t.Fatal("not JSON")
	}
	wantJSON, _ := json.Marshal(wantValue)
	if string(got) != string(wantJSON) {
		t.Errorf("configuration = %s, want %s", got, wantJSON)
	}
	if (Flavor{}).InitializationOptions() != nil {
		t.Error("pylsp takes no initialization options")
	}
}

func TestRootIsTheFolderOfTheFile(t *testing.T) {
	folder := t.TempDir()
	if got := (Flavor{}).RootOf(filepath.Join(folder, "sub", "main.py")); got != filepath.Join(folder, "sub") {
		t.Errorf("RootOf = %q", got)
	}
}

func TestMissingPythonIsMissingTool(t *testing.T) {
	cfg := Config{Locator: python.NewLocator(python.Options{
		Settings: pythonAt{path: filepath.Join(t.TempDir(), "no-python")}, AppDir: t.TempDir(),
		BaseEnvironment: []string{"PATH="}, Launcher: func(context.Context) string { return "" },
	})}
	_, _, err := NewFlavor(cfg).Command(nil)
	if !errors.Is(err, app.ErrToolNotFound) || err.Error() != app.MissingTool("python").Error() {
		t.Errorf("err = %v", err)
	}
}

func TestMissingModuleIsMissingToolPylsp(t *testing.T) {
	base := pythontest.Interpreter(t)
	venv := filepath.Join(t.TempDir(), "bare")
	if out, err := exec.Command(base, "-m", "venv", "--without-pip", venv).CombinedOutput(); err != nil {
		t.Skipf("cannot create a venv: %v: %s", err, out)
	}
	bare := filepath.Join(venv, "Scripts", "python.exe")
	if _, err := exec.LookPath(bare); err != nil {
		bare = filepath.Join(venv, "bin", "python")
	}
	flavor := NewFlavor(configAt(bare))
	_, _, err := flavor.Command(flavor.Environment())
	if !errors.Is(err, app.ErrToolNotFound) || err.Error() != app.MissingTool("pylsp").Error() {
		t.Errorf("err = %v", err)
	}
}

func TestCommandIsPythonDashMPylsp(t *testing.T) {
	path := pythontest.Interpreter(t)
	flavor := NewFlavor(configAt(path))
	executable, args, err := flavor.Command(flavor.Environment())
	if err != nil || executable != path || len(args) != 2 || args[0] != "-m" || args[1] != "pylsp" {
		t.Errorf("command = %q %v, %v", executable, args, err)
	}
	if env := flavor.Environment(); env["PYTHONUTF8"] != "1" {
		t.Errorf("environment = %v", env)
	}
}
