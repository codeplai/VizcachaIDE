package runner

import (
	"context"
	"errors"
	"os"
	"path/filepath"
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

func TestConfigureOneSourceRunsTheFile(t *testing.T) {
	folder := t.TempDir()
	touch(t, folder, "hola.cpp", "util.h")
	path := filepath.Join(folder, "hola.cpp")
	config := bareRunner().Configure(path, []string{"a", "b c"})

	if config.CodeLanguage != domain.CodeLanguageCpp || config.Mode != domain.RunFile || config.Target != path {
		t.Fatalf("config = %+v", config)
	}
	if config.WorkingDir != folder || len(config.ProgramArgs) != 2 || config.ProgramArgs[1] != "b c" {
		t.Fatalf("config = %+v", config)
	}
	if config.Project == nil || config.Project.Root != folder || config.Project.Kind != domain.ProjectFolder {
		t.Fatalf("project = %+v", config.Project)
	}
}

func TestConfigureTwoSourcesIsAProject(t *testing.T) {
	folder := filepath.Join(t.TempDir(), "calculadora")
	if err := os.Mkdir(folder, 0o755); err != nil {
		t.Fatal(err)
	}
	touch(t, folder, "main.cpp", "suma.cc")
	config := bareRunner().Configure(filepath.Join(folder, "suma.cc"), nil)

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
	config := bareRunner().Configure(filepath.Join(folder, "util.hpp"), nil)
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
