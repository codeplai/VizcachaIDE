package app

import (
	"errors"
	"os"
	"path/filepath"
	"testing"
)

type fakeShell struct{ trashed, revealed []string }

func (f *fakeShell) MoveToTrash(path string) error { f.trashed = append(f.trashed, path); return nil }
func (f *fakeShell) Reveal(path string) error      { f.revealed = append(f.revealed, path); return nil }

func TestValidateEntryName(t *testing.T) {
	for _, name := range []string{"main.go", "my folder", ".gitignore", "a-b_c.txt"} {
		if err := ValidateEntryName(name); err != nil {
			t.Errorf("%q should be valid: %v", name, err)
		}
	}
	for _, name := range []string{"", "   ", ".", "..", "a/b", `a\b`, "a:b", "a*", "a?", `a"`, "<a>", "a|b", "name."} {
		if err := ValidateEntryName(name); !errors.Is(err, ErrInvalidName) {
			t.Errorf("%q should be invalid, got %v", name, err)
		}
	}
}

func TestCreateFileAndFolder(t *testing.T) {
	dir := t.TempDir()
	ops := NewFileOperations(nil)
	file := filepath.Join(dir, "main.go")
	if err := ops.CreateFile(file, "package main\n"); err != nil {
		t.Fatal(err)
	}
	if data, _ := os.ReadFile(file); string(data) != "package main\n" {
		t.Fatalf("unexpected content %q", data)
	}
	if err := ops.CreateFile(file, "other"); !errors.Is(err, ErrAlreadyExists) {
		t.Fatalf("want ErrAlreadyExists, got %v", err)
	}
	if data, _ := os.ReadFile(file); string(data) != "package main\n" {
		t.Fatal("an existing file must not be overwritten")
	}
	folder := filepath.Join(dir, "pkg")
	if err := ops.CreateFolder(folder); err != nil {
		t.Fatal(err)
	}
	if err := ops.CreateFolder(folder); !errors.Is(err, ErrAlreadyExists) {
		t.Fatalf("want ErrAlreadyExists, got %v", err)
	}
	if err := ops.CreateFile(filepath.Join(dir, "a?.go"), ""); !errors.Is(err, ErrInvalidName) {
		t.Fatalf("want ErrInvalidName, got %v", err)
	}
}

func TestRename(t *testing.T) {
	dir := t.TempDir()
	ops := NewFileOperations(nil)
	a, b := filepath.Join(dir, "a.go"), filepath.Join(dir, "b.go")
	_ = os.WriteFile(a, []byte("a"), 0o644)
	_ = os.WriteFile(b, []byte("b"), 0o644)
	if err := ops.Rename(a, b); !errors.Is(err, ErrAlreadyExists) {
		t.Fatalf("want ErrAlreadyExists, got %v", err)
	}
	if err := ops.Rename(a, filepath.Join(dir, "bad|.go")); !errors.Is(err, ErrInvalidName) {
		t.Fatalf("want ErrInvalidName, got %v", err)
	}
	c := filepath.Join(dir, "c.go")
	if err := ops.Rename(a, c); err != nil {
		t.Fatal(err)
	}
	if _, err := os.Stat(c); err != nil {
		t.Fatal("renamed file is missing")
	}
	if err := ops.Rename(c, filepath.Join(dir, "C.go")); err != nil {
		t.Fatalf("a change of letter case must be allowed: %v", err)
	}
}

func TestTrashAndRevealUseTheShell(t *testing.T) {
	shell := &fakeShell{}
	ops := NewFileOperations(shell)
	if err := ops.MoveToTrash("x.go"); err != nil || len(shell.trashed) != 1 {
		t.Fatalf("trash not delegated: %v", err)
	}
	if err := ops.RevealInExplorer("x.go"); err != nil || len(shell.revealed) != 1 {
		t.Fatalf("reveal not delegated: %v", err)
	}
	if err := NewFileOperations(nil).MoveToTrash("x.go"); !errors.Is(err, ErrTrashUnavailable) {
		t.Fatalf("without a shell nothing may be deleted, got %v", err)
	}
}
