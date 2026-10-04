package toollocator_test

import (
	"os"
	"path/filepath"
	"runtime"
	"testing"

	"github.com/codeplai/VizcachaIDE/wails/internal/domain"
	"github.com/codeplai/VizcachaIDE/wails/internal/protocol/toollocator"
)

func tool(id string, bundled ...string) toollocator.Tool {
	return toollocator.Tool{
		Spec:               domain.ToolSpec{ID: id, Role: domain.RoleRuntime},
		Language:           domain.CodeLanguageGo,
		BundledDirectories: bundled,
	}
}

// fakeTool creates an empty executable-looking file named like the tool.
func fakeTool(t *testing.T, dir, name string) string {
	t.Helper()
	if err := os.MkdirAll(dir, 0o755); err != nil {
		t.Fatal(err)
	}
	if runtime.GOOS == "windows" {
		name += ".exe"
	}
	path := filepath.Join(dir, name)
	if err := os.WriteFile(path, []byte("x"), 0o755); err != nil {
		t.Fatal(err)
	}
	return path
}

func TestLocatorResolutionOrder(t *testing.T) {
	app, pathDir, other := t.TempDir(), t.TempDir(), t.TempDir()
	bundledGo := fakeTool(t, filepath.Join(app, "toolchain", "go", "bin"), "go")
	bundledDelve := fakeTool(t, filepath.Join(app, "toolchain", "bin"), "dlv")
	fakeTool(t, pathDir, "go")
	pathGopls := fakeTool(t, pathDir, "gopls")
	configuredGo := fakeTool(t, other, "go")
	goTool := tool("go", "toolchain/go/bin")

	cases := []struct {
		name       string
		tool       toollocator.Tool
		configured string
		wantSource domain.ToolSource
		wantPath   string
	}{
		{"configured wins over bundled and PATH", goTool, configuredGo, domain.ToolConfigured, configuredGo},
		{"configured path that does not exist is ignored", goTool, filepath.Join(other, "nope"), domain.ToolBundled, bundledGo},
		{"bundled wins over PATH", goTool, "", domain.ToolBundled, bundledGo},
		{"second bundled folder", tool("dlv", "toolchain/go/bin", "toolchain/bin"), "", domain.ToolBundled, bundledDelve},
		{"PATH when nothing is bundled", tool("gopls", "toolchain/bin"), "", domain.ToolOnPath, pathGopls},
		{"missing everywhere", tool("nothing"), "", domain.ToolMissing, ""},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			locator := toollocator.New(app, pathDir, func(string) string { return tc.configured })
			got := locator.Locate(tc.tool)
			if got.Source != tc.wantSource || got.Path != tc.wantPath {
				t.Errorf("Locate = %+v, want %s %s", got, tc.wantSource, tc.wantPath)
			}
			if got.ID != tc.tool.Spec.ID || got.CodeLanguage != domain.CodeLanguageGo || got.Role != domain.RoleRuntime {
				t.Errorf("status identity = %+v", got)
			}
		})
	}
}

func TestLocatorUsesTheExecutableNameWhenItDiffersFromTheID(t *testing.T) {
	pathDir := t.TempDir()
	want := fakeTool(t, pathDir, "python3")
	python := tool("python")
	python.Name = "python3"

	got := toollocator.New(t.TempDir(), pathDir, nil).Locate(python)

	if got.Source != domain.ToolOnPath || got.Path != want {
		t.Errorf("Locate = %+v, want path %s", got, want)
	}
}

func TestExistingDirectoriesSkipsMissingOnes(t *testing.T) {
	app := t.TempDir()
	if err := os.MkdirAll(filepath.Join(app, "toolchain", "bin"), 0o755); err != nil {
		t.Fatal(err)
	}
	locator := toollocator.New(app, "", nil)

	got := locator.ExistingDirectories("toolchain/go/bin", "toolchain/bin")

	if len(got) != 1 || got[0] != filepath.Join(app, "toolchain", "bin") {
		t.Errorf("directories = %v", got)
	}
}
