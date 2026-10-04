package python_test

import (
	"context"
	"os"
	"strings"
	"testing"

	"github.com/codeplai/VizcachaIDE/wails/internal/adapters/python"
	"github.com/codeplai/VizcachaIDE/wails/internal/adapters/python/pythontest"
	"github.com/codeplai/VizcachaIDE/wails/internal/domain"
)

type configuredPython struct{ path string }

func (c configuredPython) Load() (domain.Settings, error) {
	settings := domain.DefaultSettings()
	settings.ToolPaths = map[string]string{python.ToolPython: c.path}
	return settings, nil
}
func (configuredPython) Save(domain.Settings) error { return nil }

func TestRealInterpreterAndItsModules(t *testing.T) {
	path := pythontest.Interpreter(t)
	locator := python.NewLocator(python.Options{Settings: configuredPython{path: path}})

	interpreter, err := locator.Find(context.Background(), "")
	if err != nil {
		t.Fatal(err)
	}
	if interpreter.Path != path || interpreter.Source != domain.ToolConfigured || !strings.HasPrefix(interpreter.Version, "3.") {
		t.Fatalf("interpreter = %+v", interpreter)
	}
	versions := python.ModuleVersions(context.Background(), interpreter, python.Environment(os.Environ(), interpreter))
	for _, tool := range []string{python.ToolDebugpy, python.ToolPylsp, python.ToolRuff} {
		if versions[tool] == "" {
			t.Errorf("%s not found in %s: %v", tool, path, versions)
		}
	}
}
