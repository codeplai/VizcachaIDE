package bridge

import (
	"errors"
	"os"
	"path/filepath"
	"testing"
)

func TestWithExtensionOfTheSuggestedName(t *testing.T) {
	cases := []struct{ suggested, path, want string }{
		{"hola.go", "", ""},
		{"hola.go", "hola", "hola.go"},
		{"hola.py", "hola", "hola.py"},
		{"hola.go", "hola.go", "hola.go"},
		{"hola.go", "notas.txt", "notas.txt"},
		{"hola", "hola", "hola"},
	}
	for _, c := range cases {
		if got := withExtensionOf(c.suggested, c.path); got != c.want {
			t.Errorf("withExtensionOf(%q, %q) = %q, want %q", c.suggested, c.path, got, c.want)
		}
	}
}

func TestFileFiltersListEveryLanguage(t *testing.T) {
	service := NewFilesService(noContext{}, nil, nil).UseLanguages(newTestRegistry(t))
	filters := service.fileFilters()
	if len(filters) != 3 || filters[0].Pattern != "*.go" || filters[1].Pattern != "*.py;*.pyw" || filters[2].Pattern != "*.*" {
		t.Errorf("filters = %+v", filters)
	}
	if only := NewFilesService(noContext{}, nil, nil).fileFilters(); len(only) != 1 || only[0].Pattern != "*.*" {
		t.Errorf("without languages = %+v", only)
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
