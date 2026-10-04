package app

import (
	"errors"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"
)

func write(t *testing.T, path, text string) {
	t.Helper()
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path, []byte(text), 0o600); err != nil {
		t.Fatal(err)
	}
}

func TestSplitProgramArguments(t *testing.T) {
	cases := []struct {
		text string
		want []string
	}{
		{"", []string{}},
		{"a b", []string{"a", "b"}},
		{`uno "dos tres" 'cuatro cinco'`, []string{"uno", "dos tres", "cuatro cinco"}},
		{`C:\Users\ana\file.txt`, []string{`C:\Users\ana\file.txt`}},
		{`"" x`, []string{"", "x"}},
		{"  espacios   varios ", []string{"espacios", "varios"}},
	}
	for _, c := range cases {
		got, err := SplitProgramArguments(c.text)
		if err != nil || !reflect.DeepEqual(got, c.want) {
			t.Errorf("Split(%q) = %q, %v; want %q", c.text, got, err, c.want)
		}
	}
	if _, err := SplitProgramArguments(`abc "sin cerrar`); !errors.Is(err, ErrUnclosedQuote) {
		t.Errorf("error = %v, want ErrUnclosedQuote", err)
	}
}

func TestBuildFileTreeHidesNoiseAndSorts(t *testing.T) {
	root := t.TempDir()
	for _, name := range []string{"b.go", "A.go", "app.exe", "__debug_bin123", "lib.dll", ".git/config", "sub/z.go"} {
		write(t, filepath.Join(root, name), "x")
	}

	tree, err := BuildFileTree(root)
	if err != nil {
		t.Fatal(err)
	}

	var names []string
	for _, child := range tree.Children {
		names = append(names, child.Name)
	}
	if got := strings.Join(names, ","); got != "sub,A.go,b.go" {
		t.Errorf("children = %s, want sub,A.go,b.go", got)
	}
	if !tree.Children[0].IsDir || len(tree.Children[0].Children) != 1 {
		t.Errorf("sub = %+v", tree.Children[0])
	}
}

func TestReadAndWriteSourceFile(t *testing.T) {
	path := filepath.Join(t.TempDir(), "nueva", "hola.go")
	if err := WriteSourceFile(path, "// canción\n"); err != nil {
		t.Fatal(err)
	}
	got, err := ReadSourceFile(path)
	if err != nil || got != "// canción\n" {
		t.Errorf("ReadSourceFile = %q, %v", got, err)
	}
	if _, err := ReadSourceFile(filepath.Join(filepath.Dir(path), "no.go")); !errors.Is(err, os.ErrNotExist) {
		t.Errorf("error = %v, want not exist", err)
	}
}
