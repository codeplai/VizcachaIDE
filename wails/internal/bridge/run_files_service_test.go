package bridge

import (
	"context"
	"os"
	"path/filepath"
	"testing"
)

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
