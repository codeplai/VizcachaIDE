package bridge

import (
	"errors"
	"os"
	"path/filepath"
	"testing"
)

func TestWithGoExtension(t *testing.T) {
	cases := map[string]string{"": "", "hola": "hola.go", "hola.go": "hola.go", "notas.txt": "notas.txt"}
	for in, want := range cases {
		if got := withGoExtension(in); got != want {
			t.Errorf("withGoExtension(%q) = %q, want %q", in, got, want)
		}
	}
}

func TestDefaultDialogFolder(t *testing.T) {
	dir := t.TempDir()
	if got := defaultDialogFolder(dir, "b"); got != dir {
		t.Errorf("existing folder: got %q", got)
	}
	if got := defaultDialogFolder("untitled", "b"); got != "b" {
		t.Errorf("relative (untitled) folder: got %q", got)
	}
	if got := defaultDialogFolder("", "b"); got != "b" {
		t.Errorf("empty folder: got %q", got)
	}
}

func TestExistingFolderSkipsDeletedFolders(t *testing.T) {
	dir := t.TempDir()
	deleted := filepath.Join(dir, "dist", "release", "VizcachaIDE")
	if got := existingFolder(deleted); got != dir {
		t.Errorf("deleted folder: got %q, want its closest existing parent %q", got, dir)
	}
	file := filepath.Join(dir, "main.go")
	if err := os.WriteFile(file, []byte("package main"), 0o600); err != nil {
		t.Fatal(err)
	}
	if got := existingFolder(file); got != dir {
		t.Errorf("a file is not a folder: got %q", got)
	}
	if got := existingFolder("untitled"); got != "" {
		t.Errorf("relative path: got %q", got)
	}
}

func TestWithStartFolderRetriesWithoutFolder(t *testing.T) {
	dir := t.TempDir()
	var starts []string
	path, err := withStartFolder(dir, func(start string) (string, error) {
		starts = append(starts, start)
		if start != "" {
			return "", errors.New("the system refused the folder")
		}
		return "C:/elegido.go", nil
	})
	if err != nil || path != "C:/elegido.go" || len(starts) != 2 || starts[1] != "" {
		t.Errorf("path=%q err=%v starts=%q", path, err, starts)
	}
}
