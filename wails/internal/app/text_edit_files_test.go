package app

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/codeplai/VizcachaIDE/wails/internal/domain"
)

func oneEdit(file string, line, from, to int, text string) domain.FileEdit {
	return domain.FileEdit{File: file, Edits: []domain.TextEdit{{
		Range: domain.SourceRange{
			Start: domain.SourceLocation{File: file, Line: line, Column: from},
			End:   domain.SourceLocation{File: file, Line: line, Column: to},
		},
		NewText: text,
	}}}
}

func TestApplyEditsToFilesEditsEveryFile(t *testing.T) {
	dir := t.TempDir()
	a, b := filepath.Join(dir, "a.go"), filepath.Join(dir, "b.go")
	for path, text := range map[string]string{a: "año := 1\r\n", b: "x := año\n"} {
		if err := os.WriteFile(path, []byte(text), 0o600); err != nil {
			t.Fatal(err)
		}
	}
	texts, err := ApplyEditsToFiles([]domain.FileEdit{oneEdit(a, 1, 1, 4, "y"), oneEdit(b, 1, 6, 9, "y")})
	if err != nil {
		t.Fatal(err)
	}
	want := map[string]string{a: "y := 1\r\n", b: "x := y\n"}
	for path, expected := range want {
		onDisk, _ := os.ReadFile(path)
		if string(onDisk) != expected || texts[path] != expected {
			t.Errorf("%s: disk %q, returned %q, want %q", path, onDisk, texts[path], expected)
		}
	}
}

func TestApplyEditsToFilesWritesNothingWhenOneEditIsBad(t *testing.T) {
	dir := t.TempDir()
	a, b := filepath.Join(dir, "a.go"), filepath.Join(dir, "b.go")
	for _, path := range []string{a, b} {
		if err := os.WriteFile(path, []byte("abc\n"), 0o600); err != nil {
			t.Fatal(err)
		}
	}
	if _, err := ApplyEditsToFiles([]domain.FileEdit{oneEdit(a, 1, 1, 2, "X"), oneEdit(b, 9, 1, 2, "X")}); err == nil {
		t.Fatal("expected an error")
	}
	if onDisk, _ := os.ReadFile(a); string(onDisk) != "abc\n" {
		t.Errorf("a.go changed: %q", onDisk)
	}
}
