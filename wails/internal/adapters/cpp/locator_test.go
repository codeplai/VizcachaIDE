package cpp_test

import (
	"context"
	"errors"
	"os"
	"path/filepath"
	"testing"

	"github.com/codeplai/VizcachaIDE/wails/internal/adapters/cpp"
	"github.com/codeplai/VizcachaIDE/wails/internal/adapters/cpp/cpptest"
	"github.com/codeplai/VizcachaIDE/wails/internal/app"
	"github.com/codeplai/VizcachaIDE/wails/internal/domain"
)

type toolPaths map[string]string

func (p toolPaths) Load() (domain.Settings, error) {
	settings := domain.DefaultSettings()
	settings.ToolPaths = p
	return settings, nil
}
func (toolPaths) Save(domain.Settings) error { return nil }

func TestFamilyOfReadsTheVersionOutput(t *testing.T) {
	cases := []struct {
		path, version string
		want          cpp.Family
	}{
		{"C:/mingw64/bin/g++.exe", "g++.exe (MinGW-W64 x86_64-ucrt-posix-seh) 16.2.0", cpp.GCC},
		{"C:/llvm-mingw/bin/g++.exe", "clang version 23.1.2 (https://github.com/llvm/llvm-project.git)", cpp.Clang},
		{"/usr/bin/c++", "Apple clang version 17.0.0", cpp.Clang},
		{"/usr/bin/x86_64-linux-gnu-g++", "x86_64-linux-gnu-g++ (Ubuntu 14.2.0) 14.2.0", cpp.GCC},
	}
	for _, tc := range cases {
		if got := cpp.FamilyOf(tc.path, tc.version); got != tc.want {
			t.Errorf("FamilyOf(%s) = %s, want %s", tc.path, got, tc.want)
		}
	}
}

func TestWindowsLinksStatically(t *testing.T) {
	flags := cpp.CompileFlags(cpp.Clang, "windows")
	if flags[len(flags)-1] != "-static" || cpp.CompileFlags(cpp.GCC, "linux")[0] != "-std=c++17" {
		t.Errorf("flags = %v", flags)
	}
}

func TestMissingCompilerIsReportedAsTheCxxTool(t *testing.T) {
	locator := cpp.NewLocator(cpp.Options{
		BaseEnvironment: []string{"PATH=" + t.TempDir()}, AppDir: t.TempDir(), XcodeReady: func() bool { return true },
	})
	if _, err := locator.Compiler(context.Background()); !errors.Is(err, app.ErrToolNotFound) {
		t.Fatalf("err = %v", err)
	}
}

func TestRealCompilersAndToolsNextToThem(t *testing.T) {
	llvm := cpptest.LLVMBin(t)
	locator := cpp.NewLocator(cpp.Options{
		Settings: toolPaths{cpp.ToolCompiler: filepath.Join(llvm, cpptest.Exe("clang++"))}, AppDir: t.TempDir(),
	})
	compiler, err := locator.Compiler(context.Background())
	if err != nil || compiler.Family != cpp.Clang || compiler.Source != domain.ToolConfigured {
		t.Fatalf("compiler = %+v, %v", compiler, err)
	}
	for _, id := range []string{cpp.ToolLldbDap, cpp.ToolClangd, cpp.ToolClangFormat} {
		status := locator.Tool(context.Background(), id)
		if _, err := os.Stat(status.Path); err != nil || filepath.Dir(status.Path) != llvm {
			t.Errorf("%s = %+v", id, status)
		}
	}
	gcc := cpptest.GCCBin(t)
	onPath := cpp.NewLocator(cpp.Options{BaseEnvironment: []string{"PATH=" + gcc}, AppDir: t.TempDir()})
	if compiler, err := onPath.Compiler(context.Background()); err != nil || compiler.Family != cpp.GCC || compiler.Source != domain.ToolOnPath {
		t.Fatalf("gcc on PATH = %+v, %v", compiler, err)
	}
}
