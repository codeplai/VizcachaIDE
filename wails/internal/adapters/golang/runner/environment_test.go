package runner

import (
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"

	"github.com/codeplai/VizcachaIDE/wails/internal/domain"
	"github.com/codeplai/VizcachaIDE/wails/internal/protocol/process"
	"github.com/codeplai/VizcachaIDE/wails/internal/protocol/toollocator"
)

// fakeTool creates an empty executable-looking file named like the tool.
func fakeTool(t *testing.T, dir, tool string) {
	t.Helper()
	if err := os.MkdirAll(dir, 0o755); err != nil {
		t.Fatal(err)
	}
	if runtime.GOOS == "windows" {
		tool += ".exe"
	}
	if err := os.WriteFile(filepath.Join(dir, tool), []byte("x"), 0o755); err != nil {
		t.Fatal(err)
	}
}

func TestBundledEnvironment(t *testing.T) {
	app := t.TempDir()
	fakeTool(t, filepath.Join(app, "toolchain", "go", "bin"), "go")
	fakeTool(t, filepath.Join(app, "toolchain", "bin"), "dlv")
	locator := toollocator.New(app, "", nil)

	env := buildEnvironment([]string{"Path=C:\\Windows", "HOME=/h"}, locator)

	if env["GOROOT"] != filepath.Join(app, "toolchain", "go") {
		t.Errorf("GOROOT = %q", env["GOROOT"])
	}
	if env["GOTOOLCHAIN"] != "local" {
		t.Errorf("GOTOOLCHAIN = %q, want local", env["GOTOOLCHAIN"])
	}
	wantPrefix := filepath.Join(app, "toolchain", "go", "bin") + string(os.PathListSeparator) +
		filepath.Join(app, "toolchain", "bin") + string(os.PathListSeparator)
	if !strings.HasPrefix(env["Path"], wantPrefix) || !strings.HasSuffix(env["Path"], "C:\\Windows") {
		t.Errorf("Path = %q, want bundled folders in front", env["Path"])
	}
	if _, duplicated := env["PATH"]; duplicated {
		t.Error("the existing spelling of the PATH variable must be reused")
	}
}

func TestEnvironmentWithoutBundleIsUntouched(t *testing.T) {
	locator := toollocator.New(t.TempDir(), "", nil)
	env := buildEnvironment([]string{"PATH=/bin", "=C:=C:\\x"}, locator)
	if env["PATH"] != "/bin" || env["=C:"] != "C:\\x" || env["GOROOT"] != "" {
		t.Errorf("env = %v", env)
	}
}

func TestToolsReportsTheThreeGoToolsWithTheirSource(t *testing.T) {
	app, pathDir := t.TempDir(), t.TempDir()
	fakeTool(t, filepath.Join(app, "toolchain", "go", "bin"), "go")
	fakeTool(t, pathDir, "gopls")
	runner := New(process.New(newTestSink()), Options{AppDir: app, BaseEnvironment: []string{"PATH=" + pathDir}})

	tools := runner.Tools(t.Context())

	if len(tools) != 3 {
		t.Fatalf("tools = %+v, want go, dlv and gopls", tools)
	}
	want := map[string]domain.ToolSource{"go": domain.ToolBundled, "dlv": domain.ToolMissing, "gopls": domain.ToolOnPath}
	for _, status := range tools {
		if status.Source != want[status.ID] || status.CodeLanguage != domain.CodeLanguageGo {
			t.Errorf("%s = %+v, want source %s", status.ID, status, want[status.ID])
		}
	}
}
