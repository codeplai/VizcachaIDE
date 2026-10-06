package runner

import (
	"context"
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/codeplai/VizcachaIDE/wails/internal/app"
	"github.com/codeplai/VizcachaIDE/wails/internal/domain"
	"github.com/codeplai/VizcachaIDE/wails/internal/protocol/process"
)

func touch(t *testing.T, folder string, names ...string) {
	t.Helper()
	for _, name := range names {
		if err := os.WriteFile(filepath.Join(folder, name), []byte("int x;\n"), 0o600); err != nil {
			t.Fatal(err)
		}
	}
}

func bareRunner() *Runner { return New(process.New(newTestSink()), Options{AppDir: "."}) }

func TestConfigureOneSourceCreatesTheCMakeProject(t *testing.T) {
	folder := filepath.Join(t.TempDir(), "Ñandú")
	if err := os.Mkdir(folder, 0o755); err != nil {
		t.Fatal(err)
	}
	touch(t, folder, "hola.cpp", "util.h")
	path := filepath.Join(folder, "hola.cpp")
	r := New(process.New(newTestSink()), Options{AppDir: ".", Language: func() string { return "es" }})
	config := r.Configure(path, []string{"a", "b c"})

	if config.CodeLanguage != domain.CodeLanguageCpp || config.Mode != domain.RunProject || config.Target != folder {
		t.Fatalf("config = %+v", config)
	}
	if config.WorkingDir != folder || len(config.ProgramArgs) != 2 || config.ProgramArgs[1] != "b c" {
		t.Fatalf("config = %+v", config)
	}
	if config.Project == nil || config.Project.Root != folder || config.Project.Kind != domain.ProjectFolder {
		t.Fatalf("project = %+v", config.Project)
	}
	lists, err := os.ReadFile(filepath.Join(folder, "CMakeLists.txt"))
	if err != nil || !strings.Contains(string(lists), "add_executable(nandu ") || !strings.Contains(string(lists), "Todo .cpp") {
		t.Fatalf("CMakeLists.txt = %q, %v, want the Spanish template for target nandu", lists, err)
	}
	for _, name := range []string{"vcpkg.json", "CMakePresets.json", ".gitignore"} {
		if _, err := os.Stat(filepath.Join(folder, name)); err != nil {
			t.Errorf("%s: %v", name, err)
		}
	}
}

func TestConfigureKeepsAnExistingCMakeLists(t *testing.T) {
	root := t.TempDir()
	sub := filepath.Join(root, "src")
	if err := os.Mkdir(sub, 0o755); err != nil {
		t.Fatal(err)
	}
	mine := "project(mio)\nadd_executable(mio src/a.cpp)\n"
	if err := os.WriteFile(filepath.Join(root, "CMakeLists.txt"), []byte(mine), 0o600); err != nil {
		t.Fatal(err)
	}
	touch(t, sub, "a.cpp")
	config := bareRunner().Configure(filepath.Join(sub, "a.cpp"), nil)
	if config.Mode != domain.RunProject || config.Target != root {
		t.Fatalf("config = %+v, want the project at %s", config, root)
	}
	if text, _ := os.ReadFile(filepath.Join(root, "CMakeLists.txt")); string(text) != mine {
		t.Fatalf("CMakeLists.txt was changed: %q", text)
	}
	if _, err := os.Stat(filepath.Join(sub, "CMakeLists.txt")); err == nil {
		t.Fatal("a subfolder of a project must not get its own CMakeLists.txt")
	}
}

func TestConfigureTwoSourcesIsAProject(t *testing.T) {
	folder := filepath.Join(t.TempDir(), "calculadora")
	if err := os.Mkdir(folder, 0o755); err != nil {
		t.Fatal(err)
	}
	touch(t, folder, "main.cpp", "suma.cc")
	config := configureDirect(filepath.Join(folder, "suma.cc"), nil)

	if config.Mode != domain.RunProject || config.Target != folder {
		t.Fatalf("config = %+v", config)
	}
	if config.Project == nil || config.Project.Root != folder || config.Project.Kind != domain.ProjectFolder || config.Project.Name != "calculadora" {
		t.Fatalf("project = %+v", config.Project)
	}
	if config.ProgramArgs == nil {
		t.Fatal("ProgramArgs must be [] not null")
	}
}

func TestConfigureHeaderResolvesToItsFolder(t *testing.T) {
	folder := t.TempDir()
	touch(t, folder, "main.cpp", "util.hpp")
	config := configureDirect(filepath.Join(folder, "util.hpp"), nil)
	if config.Mode != domain.RunProject || config.Target != folder {
		t.Fatalf("config = %+v", config)
	}
	sources, err := sourcesOf(config)
	if err != nil || len(sources) != 1 || filepath.Base(sources[0]) != "main.cpp" {
		t.Fatalf("sources = %v, %v", sources, err)
	}
}

func TestHeaderWithoutSourcesIsAnError(t *testing.T) {
	folder := t.TempDir()
	touch(t, folder, "util.h")
	r := bareRunner()
	config := r.Configure(filepath.Join(folder, "util.h"), nil)
	for name, call := range map[string]func() error{
		"Run":   func() error { return r.Run(context.Background(), config) },
		"Build": func() error { return r.Build(context.Background(), config) },
		"Check": func() error { _, err := r.Check(context.Background(), config); return err },
	} {
		err := call()
		if !errors.Is(err, ErrHeaderOnly) || !errors.Is(err, app.ErrUnsupported) {
			t.Errorf("%s = %v, want ErrHeaderOnly", name, err)
		}
	}
	if r.IsRunning() {
		t.Fatal("nothing should have started")
	}
}

func TestMissingCompiler(t *testing.T) {
	folder := t.TempDir()
	touch(t, folder, "main.cpp")
	r := New(process.New(newTestSink()), Options{AppDir: t.TempDir(), BaseEnvironment: []string{"PATH="}, CacheDir: t.TempDir()})
	err := r.Run(context.Background(), r.Configure(filepath.Join(folder, "main.cpp"), nil))
	if !errors.Is(err, app.ErrToolNotFound) {
		t.Fatalf("Run = %v, want a missing tool", err)
	}
}
