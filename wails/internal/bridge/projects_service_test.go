package bridge

import (
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/codeplai/VizcachaIDE/wails/internal/app"
	"github.com/codeplai/VizcachaIDE/wails/internal/domain"
)

type fakeScaffold struct{}

func (fakeScaffold) Scaffold(name string) (map[string]string, string) {
	return map[string]string{"main.x": "hola " + name, "src/lib.x": "lib"}, "src/lib.x"
}

func newProjectsService(t *testing.T) *ProjectsService {
	t.Helper()
	with := app.UnavailableSupport(domain.LanguageProfile{ID: domain.CodeLanguageGo, Extensions: []string{".go"}})
	with.Scaffold = fakeScaffold{}
	without := app.UnavailableSupport(domain.LanguageProfile{ID: domain.CodeLanguagePython, Extensions: []string{".py"}})
	registry, err := app.NewLanguageRegistry(domain.CodeLanguageGo, with, without)
	if err != nil {
		t.Fatal(err)
	}
	return NewProjectsService(registry)
}

func TestCreateWritesTheTemplateAndReturnsPaths(t *testing.T) {
	location := t.TempDir()
	project, err := newProjectsService(t).Create(domain.CodeLanguageGo, location, "Mi Ñandú")
	if err != nil {
		t.Fatal(err)
	}
	root := filepath.Join(location, "Mi Ñandú")
	if project.Root != root || project.MainFile != filepath.Join(root, "src", "lib.x") {
		t.Fatalf("project = %+v", project)
	}
	text, err := os.ReadFile(filepath.Join(root, "main.x"))
	if err != nil || string(text) != "hola Mi Ñandú" {
		t.Fatalf("main.x = %q, %v", text, err)
	}
	if _, err := os.Stat(project.MainFile); err != nil {
		t.Errorf("main file missing: %v", err)
	}
}

func TestCreateAcceptsAnEmptyFolder(t *testing.T) {
	location := t.TempDir()
	if err := os.Mkdir(filepath.Join(location, "vacia"), 0o755); err != nil {
		t.Fatal(err)
	}
	if _, err := newProjectsService(t).Create(domain.CodeLanguageGo, location, "vacia"); err != nil {
		t.Fatal(err)
	}
}

func TestCreateErrorsCarryTheirKey(t *testing.T) {
	location := t.TempDir()
	taken := filepath.Join(location, "ocupada")
	if err := os.Mkdir(taken, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(taken, "a.txt"), nil, 0o644); err != nil {
		t.Fatal(err)
	}
	cases := []struct {
		name     string
		language domain.CodeLanguage
		location string
		project  string
		want     error
	}{
		{"empty name", domain.CodeLanguageGo, location, "  ", app.ErrProjectNameEmpty},
		{"slash", domain.CodeLanguageGo, location, "a/b", app.ErrProjectNameInvalid},
		{"colon", domain.CodeLanguageGo, location, "a:b", app.ErrProjectNameInvalid},
		{"pipe", domain.CodeLanguageGo, location, "a|b", app.ErrProjectNameInvalid},
		{"only dots", domain.CodeLanguageGo, location, "...", app.ErrProjectNameInvalid},
		{"trailing dot", domain.CodeLanguageGo, location, "hola.", app.ErrProjectNameInvalid},
		{"folder in use", domain.CodeLanguageGo, location, "ocupada", app.ErrProjectFolderTaken},
		{"no location", domain.CodeLanguageGo, "", "hola", app.ErrProjectLocationMissing},
		{"missing location", domain.CodeLanguageGo, filepath.Join(location, "no-existe"), "hola", app.ErrProjectLocationMissing},
		{"no scaffold", domain.CodeLanguagePython, location, "hola", app.ErrProjectNoScaffold},
		{"unknown language", domain.CodeLanguage("zig"), location, "hola", app.ErrUnknownCodeLanguage},
	}
	for _, tc := range cases {
		_, err := newProjectsService(t).Create(tc.language, tc.location, tc.project)
		if !errors.Is(err, tc.want) {
			t.Errorf("%s: err = %v, want %v", tc.name, err, tc.want)
		}
	}
	if !strings.HasPrefix(app.ErrProjectFolderTaken.Error(), "project.") {
		t.Error("the sentinels must be i18n keys")
	}
	if _, err := os.Stat(filepath.Join(location, "a")); err == nil {
		t.Error("an invalid name must not create anything")
	}
}
