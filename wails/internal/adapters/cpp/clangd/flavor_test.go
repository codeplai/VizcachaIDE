package clangd

import (
	"errors"
	"path/filepath"
	"slices"
	"strings"
	"testing"

	"github.com/codeplai/VizcachaIDE/wails/internal/adapters/cpp"
	"github.com/codeplai/VizcachaIDE/wails/internal/adapters/cpp/cpptest"
	"github.com/codeplai/VizcachaIDE/wails/internal/app"
)

func TestFixedArgumentsAndNoQueryDriverForClang(t *testing.T) {
	llvm := cpptest.LLVMBin(t)
	executable, args, err := NewFlavor(Config{Locator: locatorFor(llvm, llvm)}).Command(nil)
	if err != nil || filepath.Base(executable) != cpptest.Exe("clangd") {
		t.Fatalf("command = %q %v, %v", executable, args, err)
	}
	want := []string{"--background-index=false", "--header-insertion=never", "--completion-style=detailed", "--log=error"}
	if !slices.Equal(args, want) {
		t.Errorf("args = %v, want %v", args, want)
	}
}

func TestQueryDriverPointsAtGCC(t *testing.T) {
	gcc := cpptest.GCCBin(t)
	_, args, err := NewFlavor(Config{Locator: locatorFor(gcc, cpptest.LLVMBin(t))}).Command(nil)
	want := "--query-driver=" + filepath.Join(gcc, cpptest.Exe("g++"))
	if err != nil || !slices.ContainsFunc(args, func(arg string) bool { return strings.EqualFold(arg, want) }) {
		t.Errorf("args = %v, %v, want %s", args, err, want)
	}
}

func TestInitializationOptionsAreTheFallbackFlags(t *testing.T) {
	got, _ := Flavor{}.InitializationOptions().(map[string]any)
	if flags, _ := got["fallbackFlags"].([]string); !slices.Equal(flags, []string{"-std=c++17", "-Wall", "-Wextra"}) {
		t.Errorf("options = %v", got)
	}
	if (Flavor{}).Configuration() != nil {
		t.Error("clangd takes no configuration")
	}
}

func TestRootIsTheFolderOfTheFile(t *testing.T) {
	folder := t.TempDir()
	if got := (Flavor{}).RootOf(filepath.Join(folder, "sub", "main.cpp")); got != filepath.Join(folder, "sub") {
		t.Errorf("RootOf = %q", got)
	}
}

func TestMissingClangdIsMissingTool(t *testing.T) {
	locator := cpp.NewLocator(cpp.Options{AppDir: t.TempDir(), BaseEnvironment: []string{"PATH="}})
	_, _, err := NewFlavor(Config{Locator: locator}).Command(nil)
	if !errors.Is(err, app.ErrToolNotFound) || err.Error() != app.MissingTool("clangd").Error() {
		t.Errorf("err = %v", err)
	}
}
