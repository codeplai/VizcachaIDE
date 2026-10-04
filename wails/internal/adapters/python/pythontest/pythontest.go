// Package pythontest gives the tests of the Python adapter a real interpreter, or skips them.
// It is imported only from _test.go files.
package pythontest

import (
	"os"
	"path/filepath"
	"runtime"
	"testing"
)

// EnvVariable names the interpreter to test with. Without it, the development venv of
// docs/PLAN_PYTHON.md section 0 (wails/.venv-py312) is used when it exists.
const EnvVariable = "VIZCACHA_TEST_PYTHON"

// Interpreter returns the path of a Python 3.12 with debugpy, python-lsp-server and ruff, or
// skips the test when there is none.
func Interpreter(t *testing.T) string {
	t.Helper()
	if path := os.Getenv(EnvVariable); path != "" {
		return path
	}
	if path := developmentVenv(); path != "" {
		return path
	}
	t.Skipf("no Python to test with: set %s or create wails/.venv-py312 (docs/PLAN_PYTHON.md section 0)", EnvVariable)
	return ""
}

// developmentVenv finds wails/.venv-py312 walking up from the test's folder.
func developmentVenv() string {
	dir, err := os.Getwd()
	if err != nil {
		return ""
	}
	for {
		candidate := filepath.Join(dir, ".venv-py312", "bin", "python")
		if runtime.GOOS == "windows" {
			candidate = filepath.Join(dir, ".venv-py312", "Scripts", "python.exe")
		}
		if _, err := os.Stat(candidate); err == nil {
			return candidate
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			return ""
		}
		dir = parent
	}
}
