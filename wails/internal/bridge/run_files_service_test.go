package bridge

import (
	"context"
	"errors"
	"os"
	"path/filepath"
	"testing"

	"github.com/codeplai/VizcachaIDE/wails/internal/app"
	"github.com/codeplai/VizcachaIDE/wails/internal/domain"
)

// fakeToolchain records what RunService asks for.
type fakeToolchain struct {
	app.Toolchain
	runConfig domain.RunConfiguration
	goArgs    []string
	runErr    error
}

func (f *fakeToolchain) Run(_ context.Context, c domain.RunConfiguration) error {
	f.runConfig = c
	return f.runErr
}

func (f *fakeToolchain) RunGoCommand(_ context.Context, _ string, args []string) error {
	f.goArgs = args
	return nil
}

func TestRunServiceBuildsTheConfigurationFromTheFile(t *testing.T) {
	root := t.TempDir()
	if err := os.WriteFile(filepath.Join(root, "go.mod"), []byte("module example.com/m\n"), 0o600); err != nil {
		t.Fatal(err)
	}
	fake := &fakeToolchain{}

	config, err := NewRunService(fake).Run(filepath.Join(root, "main.go"), []string{"a"})

	if err != nil || config.Mode != domain.RunPackage || fake.runConfig.Module == nil {
		t.Errorf("config = %+v, err = %v", config, err)
	}
}

func TestRunServicePropagatesBusy(t *testing.T) {
	fake := &fakeToolchain{runErr: app.ErrBusy}
	_, err := NewRunService(fake).Run("main.go", nil)
	if !errors.Is(err, app.ErrBusy) {
		t.Errorf("error = %v, want ErrBusy", err)
	}
}

func TestRunServiceValidatesModuleArguments(t *testing.T) {
	fake := &fakeToolchain{}
	service := NewRunService(fake)
	if err := service.ModGet("dir", "-u"); !errors.Is(err, app.ErrInvalidGoArgument) {
		t.Errorf("ModGet error = %v", err)
	}
	if err := service.ModTidy("dir"); err != nil || len(fake.goArgs) != 2 || fake.goArgs[1] != "tidy" {
		t.Errorf("ModTidy args = %v, err = %v", fake.goArgs, err)
	}
}

type noContext struct{}

func (noContext) Context() context.Context { return nil }

func TestFilesServiceRemembersTheLastFolder(t *testing.T) {
	store := NewMemorySettingsStore()
	service := NewFilesService(noContext{}, store, nil)
	root := t.TempDir()
	if err := os.WriteFile(filepath.Join(root, "main.go"), []byte("package main"), 0o600); err != nil {
		t.Fatal(err)
	}

	tree, err := service.ListTree(root)
	if err != nil || len(tree.Children) != 1 {
		t.Fatalf("tree = %+v, err = %v", tree, err)
	}
	if saved, _ := store.Load(); saved.LastFolder != root {
		t.Errorf("LastFolder = %q, want %q", saved.LastFolder, root)
	}
	again, err := service.OpenFolder() // no Wails context: falls back to the last folder
	if err != nil || again.Path != root {
		t.Errorf("OpenFolder = %+v, err = %v", again, err)
	}
}

func TestFilesServiceWithoutFolderReturnsAnEmptyNode(t *testing.T) {
	tree, err := NewFilesService(noContext{}, NewMemorySettingsStore(), nil).ListTree("")
	if err != nil || tree.Path != "" || tree.Children == nil {
		t.Errorf("tree = %+v, err = %v", tree, err)
	}
}

func TestFilesServiceSavesAndReads(t *testing.T) {
	service := NewFilesService(noContext{}, nil, nil)
	path := filepath.Join(t.TempDir(), "x.go")
	if err := service.SaveFile(path, "package x\n"); err != nil {
		t.Fatal(err)
	}
	if text, err := service.ReadFile(path); err != nil || text != "package x\n" {
		t.Errorf("ReadFile = %q, %v", text, err)
	}
}
