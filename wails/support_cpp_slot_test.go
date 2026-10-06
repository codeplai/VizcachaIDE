package main

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/codeplai/VizcachaIDE/wails/internal/adapters/cpp"
	"github.com/codeplai/VizcachaIDE/wails/internal/adapters/cpp/vcpkg"
	"github.com/codeplai/VizcachaIDE/wails/internal/adapters/settings"
	"github.com/codeplai/VizcachaIDE/wails/internal/bridge"
	"github.com/codeplai/VizcachaIDE/wails/internal/domain"
	"github.com/codeplai/VizcachaIDE/wails/internal/protocol/process"
)

// The package manager and the runner of C++ share one slot, so neither starts while the other works.
func TestCppSupportOffersPackagesSearchAndTheCMakeScaffold(t *testing.T) {
	texts, err := newBackendTexts(func() string { return domain.LanguageEN })
	if err != nil {
		t.Fatal(err)
	}
	sink := bridge.NewWailsEventSink()
	store := settings.NewStore(filepath.Join(t.TempDir(), "settings.json"))
	support, _, err := newCppSupport(sink, store, texts, process.New(sink))
	if err != nil {
		t.Fatal(err)
	}
	if support.Packages == nil || support.Search == nil || len(support.Profile.Capabilities.PackageActions) == 0 {
		t.Fatalf("C++ must offer packages and search: %+v", support.Profile.Capabilities)
	}
	if support.Runner.IsRunning() {
		t.Error("nothing runs at the start")
	}
	if err := support.Runner.Stop(); err != nil {
		t.Errorf("Stop while idle: %v", err)
	}
	files, _ := support.Scaffold.Scaffold("Ñandú")
	if !strings.Contains(files["CMakeLists.txt"], "nandu") {
		t.Errorf("the scaffold is not a CMake project: %v", files)
	}
}

func TestCppShellVariablesGiveVcpkgRoot(t *testing.T) {
	root := t.TempDir()
	locator := vcpkg.NewLocator(vcpkg.LocatorOptions{
		AppDir: t.TempDir(), BaseEnvironment: []string{"VCPKG_ROOT=" + root}, GOOS: "linux",
	})
	shell := cppShell{Locator: cpp.NewLocator(cpp.Options{}), libraries: locator}
	if got := shell.ShellVariables(context.Background()); len(got) != 0 {
		t.Errorf("an invalid root must give nothing, got %v", got)
	}
	for _, name := range []string{".vcpkg-root", "vcpkg"} {
		if err := os.WriteFile(filepath.Join(root, name), nil, 0o600); err != nil {
			t.Fatal(err)
		}
	}
	if got := shell.ShellVariables(context.Background()); len(got) != 1 || got[0] != "VCPKG_ROOT="+root {
		t.Errorf("ShellVariables = %v", got)
	}
}
