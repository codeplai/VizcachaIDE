//go:build windows

package filesystem

import (
	"os"
	"path/filepath"
	"testing"
)

func TestDoubleNulTerminated(t *testing.T) {
	units, err := doubleNulTerminated(`C:\a b\c.go`)
	if err != nil {
		t.Fatal(err)
	}
	n := len(units)
	if n < 3 || units[n-1] != 0 || units[n-2] != 0 || units[n-3] == 0 {
		t.Fatalf("path must end with exactly two NULs: %v", units)
	}
}

func TestTrashFlags(t *testing.T) {
	const want = 0x0040 | 0x0010 | 0x0004 | 0x0400 // ALLOWUNDO | NOCONFIRMATION | SILENT | NOERRORUI
	if trashFlags != want {
		t.Fatalf("flags = %#x, want %#x", trashFlags, want)
	}
}

func TestRevealCommandLine(t *testing.T) {
	if got := revealCommandLine(`C:\a b\c.go`); got != `explorer.exe /select,"C:\a b\c.go"` {
		t.Fatal(got)
	}
}

func TestMoveToTrashRealFile(t *testing.T) {
	dir := t.TempDir()
	file := filepath.Join(dir, "trash-me.go")
	if err := os.WriteFile(file, []byte("package main"), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := New().MoveToTrash(file); err != nil {
		t.Skipf("Recycle Bin unavailable here: %v", err)
	}
	if _, err := os.Stat(file); !os.IsNotExist(err) {
		t.Fatal("the file should have left its folder")
	}
}

func TestMoveToTrashMissingFile(t *testing.T) {
	if err := New().MoveToTrash(filepath.Join(t.TempDir(), "nope.go")); err == nil {
		t.Fatal("a missing path must be an error")
	}
}
