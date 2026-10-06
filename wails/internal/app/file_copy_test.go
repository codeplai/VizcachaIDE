package app

import (
	"errors"
	"os"
	"path/filepath"
	"testing"
)

func TestCopyFileKeepsTheSource(t *testing.T) {
	dir := t.TempDir()
	write(t, filepath.Join(dir, "canción número 1.txt"), "hola")
	to := filepath.Join(dir, "canción número 1 (copia).txt")
	if err := NewFileOperations(nil).Copy(filepath.Join(dir, "canción número 1.txt"), to); err != nil {
		t.Fatal(err)
	}
	if data, _ := os.ReadFile(to); string(data) != "hola" {
		t.Fatalf("copy content %q", data)
	}
	if _, err := os.Stat(filepath.Join(dir, "canción número 1.txt")); err != nil {
		t.Fatal("the source must stay")
	}
}

func TestCopyFolderIsRecursive(t *testing.T) {
	dir := t.TempDir()
	write(t, filepath.Join(dir, "src", "a.go"), "a")
	write(t, filepath.Join(dir, "src", "sub dir", "deep", "b.go"), "b")
	if err := os.Mkdir(filepath.Join(dir, "src", "empty"), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := NewFileOperations(nil).Copy(filepath.Join(dir, "src"), filepath.Join(dir, "dst")); err != nil {
		t.Fatal(err)
	}
	for path, want := range map[string]string{"a.go": "a", filepath.Join("sub dir", "deep", "b.go"): "b"} {
		if data, _ := os.ReadFile(filepath.Join(dir, "dst", path)); string(data) != want {
			t.Errorf("%s: got %q", path, data)
		}
	}
	if info, err := os.Stat(filepath.Join(dir, "dst", "empty")); err != nil || !info.IsDir() {
		t.Error("empty folders are copied too")
	}
}

func TestCopyNeverOverwrites(t *testing.T) {
	dir := t.TempDir()
	write(t, filepath.Join(dir, "a.txt"), "new")
	write(t, filepath.Join(dir, "b.txt"), "old")
	err := NewFileOperations(nil).Copy(filepath.Join(dir, "a.txt"), filepath.Join(dir, "b.txt"))
	if !errors.Is(err, ErrAlreadyExists) {
		t.Fatalf("want ErrAlreadyExists, got %v", err)
	}
	if data, _ := os.ReadFile(filepath.Join(dir, "b.txt")); string(data) != "old" {
		t.Fatal("the existing file was overwritten")
	}
}

func TestCopyRefusesTargetInsideSource(t *testing.T) {
	dir := t.TempDir()
	write(t, filepath.Join(dir, "src", "a.go"), "a")
	ops := NewFileOperations(nil)
	for _, to := range []string{filepath.Join(dir, "src", "copy"), filepath.Join(dir, "src", "x", "y")} {
		if err := ops.Copy(filepath.Join(dir, "src"), to); !errors.Is(err, ErrIntoItself) {
			t.Errorf("%s: want ErrIntoItself, got %v", to, err)
		}
	}
	if _, err := os.Stat(filepath.Join(dir, "src", "copy")); err == nil {
		t.Fatal("nothing must be created")
	}
}

func TestCopyOfMissingSourceLeavesNothing(t *testing.T) {
	dir := t.TempDir()
	to := filepath.Join(dir, "to")
	if err := NewFileOperations(nil).Copy(filepath.Join(dir, "gone"), to); err == nil {
		t.Fatal("a missing source must fail")
	}
	if _, err := os.Stat(to); err == nil {
		t.Fatal("the failed copy must not leave the target behind")
	}
}

func TestRenameMovesIntoAnotherFolderButNotIntoItself(t *testing.T) {
	dir := t.TempDir()
	write(t, filepath.Join(dir, "a.txt"), "a")
	write(t, filepath.Join(dir, "src", "b.txt"), "b")
	ops := NewFileOperations(nil)
	if err := ops.Rename(filepath.Join(dir, "a.txt"), filepath.Join(dir, "src", "a.txt")); err != nil {
		t.Fatal(err)
	}
	if _, err := os.Stat(filepath.Join(dir, "src", "a.txt")); err != nil {
		t.Fatal("the file did not move")
	}
	err := ops.Rename(filepath.Join(dir, "src"), filepath.Join(dir, "src", "inner", "src"))
	if !errors.Is(err, ErrIntoItself) {
		t.Fatalf("want ErrIntoItself, got %v", err)
	}
}
