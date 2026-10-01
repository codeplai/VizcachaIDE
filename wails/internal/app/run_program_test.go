package app

import (
	"errors"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"

	"github.com/codeplai/VizcachaIDE/wails/internal/domain"
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

func TestConfigurationForFileWithoutModuleRunsTheFile(t *testing.T) {
	dir := t.TempDir()
	write(t, filepath.Join(dir, "main.go"), "package main")
	config := ConfigurationForFile(filepath.Join(dir, "main.go"), nil)
	if config.Mode != domain.RunFile || config.Module != nil || config.WorkingDir != dir {
		t.Errorf("config = %+v", config)
	}
}

func TestConfigurationForFileInsideModuleRunsThePackage(t *testing.T) {
	root := t.TempDir()
	write(t, filepath.Join(root, "go.mod"), "module example.com/app\n\ngo 1.21\n")
	file := filepath.Join(root, "cmd", "main.go")
	write(t, file, "package main")

	config := ConfigurationForFile(file, []string{"-v"})

	if config.Mode != domain.RunPackage || config.WorkingDir != filepath.Dir(file) {
		t.Errorf("config = %+v", config)
	}
	if config.Module == nil || config.Module.Root != root || config.Module.ModulePath != "example.com/app" {
		t.Errorf("module = %+v", config.Module)
	}
	if config.GoTargetArgument() != "." {
		t.Errorf("target = %q", config.GoTargetArgument())
	}
}

func TestParseModulePath(t *testing.T) {
	cases := map[string]string{
		"module example.com/x\n\ngo 1.21": "example.com/x",
		"// c\nmodule \"quoted/mod\"\n":   "quoted/mod",
		"go 1.21":                         "",
	}
	for text, want := range cases {
		if got := ParseModulePath(text); got != want {
			t.Errorf("ParseModulePath(%q) = %q, want %q", text, got, want)
		}
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

func TestGoCommandArgumentsAreValidated(t *testing.T) {
	if got, err := ModInitArguments(" example.com/x "); err != nil || !reflect.DeepEqual(got, []string{"mod", "init", "example.com/x"}) {
		t.Errorf("ModInitArguments = %v, %v", got, err)
	}
	if got, err := GetArguments("github.com/google/uuid"); err != nil || !reflect.DeepEqual(got, []string{"get", "github.com/google/uuid"}) {
		t.Errorf("GetArguments = %v, %v", got, err)
	}
	for _, bad := range []string{"", "  ", "-u", "two words"} {
		if _, err := GetArguments(bad); !errors.Is(err, ErrInvalidGoArgument) {
			t.Errorf("GetArguments(%q) error = %v", bad, err)
		}
		if _, err := ModInitArguments(bad); !errors.Is(err, ErrInvalidGoArgument) {
			t.Errorf("ModInitArguments(%q) error = %v", bad, err)
		}
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
