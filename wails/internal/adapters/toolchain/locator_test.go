package toolchain

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// fakeTool creates an empty executable-looking file named like the tool.
func fakeTool(t *testing.T, dir, tool string) string {
	t.Helper()
	if err := os.MkdirAll(dir, 0o755); err != nil {
		t.Fatal(err)
	}
	path := filepath.Join(dir, executableName(tool))
	if err := os.WriteFile(path, []byte("x"), 0o755); err != nil {
		t.Fatal(err)
	}
	return path
}

func TestLocatorResolutionOrder(t *testing.T) {
	app, pathDir, other := t.TempDir(), t.TempDir(), t.TempDir()
	bundledGo := fakeTool(t, filepath.Join(app, "toolchain", "go", "bin"), ToolGo)
	bundledDelve := fakeTool(t, filepath.Join(app, "toolchain", "bin"), ToolDelve)
	pathGo := fakeTool(t, pathDir, ToolGo)
	pathGopls := fakeTool(t, pathDir, ToolGopls)
	configuredGo := fakeTool(t, other, ToolGo)

	cases := []struct {
		name       string
		tool       string
		configured string
		want       Location
	}{
		{"configured wins over bundled and PATH", ToolGo, configuredGo, Location{ToolGo, OriginConfigured, configuredGo}},
		{"configured path that does not exist is ignored", ToolGo, filepath.Join(other, "nope"), Location{ToolGo, OriginBundled, bundledGo}},
		{"bundled wins over PATH", ToolGo, "", Location{ToolGo, OriginBundled, bundledGo}},
		{"bundled dlv lives in toolchain/bin", ToolDelve, "", Location{ToolDelve, OriginBundled, bundledDelve}},
		{"PATH when nothing is bundled", ToolGopls, "", Location{ToolGopls, OriginPath, pathGopls}},
		{"missing everywhere", "nothing", "", Location{"nothing", OriginMissing, ""}},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			configured := func(string) string { return tc.configured }
			locator := NewLocator(app, pathDir, configured)
			if got := locator.Locate(tc.tool); got != tc.want {
				t.Errorf("Locate(%s) = %+v, want %+v", tc.tool, got, tc.want)
			}
		})
	}
	if pathGo == bundledGo {
		t.Fatal("test setup: PATH go and bundled go must differ")
	}
}

func TestLocatorFindsGoOnPathWhenNotBundled(t *testing.T) {
	pathDir := t.TempDir()
	want := fakeTool(t, pathDir, ToolGo)
	got := NewLocator(t.TempDir(), pathDir, nil).Locate(ToolGo)
	if got.Origin != OriginPath || got.Path != want {
		t.Errorf("Locate = %+v, want path %s", got, want)
	}
}

func TestBundledEnvironment(t *testing.T) {
	app := t.TempDir()
	fakeTool(t, filepath.Join(app, "toolchain", "go", "bin"), ToolGo)
	fakeTool(t, filepath.Join(app, "toolchain", "bin"), ToolDelve)
	locator := NewLocator(app, "", nil)

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
	locator := NewLocator(t.TempDir(), "", nil)
	env := buildEnvironment([]string{"PATH=/bin", "=C:=C:\\x"}, locator)
	if env["PATH"] != "/bin" || env["=C:"] != "C:\\x" || env["GOROOT"] != "" {
		t.Errorf("env = %v", env)
	}
}
