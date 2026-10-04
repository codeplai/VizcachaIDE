package runner

import (
	"context"
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/codeplai/VizcachaIDE/wails/internal/adapters/python"
	"github.com/codeplai/VizcachaIDE/wails/internal/app"
	"github.com/codeplai/VizcachaIDE/wails/internal/domain"
	"github.com/codeplai/VizcachaIDE/wails/internal/protocol/process"
)

func TestConfigureRunsTheFileFromItsFolder(t *testing.T) {
	folder := t.TempDir()
	path := filepath.Join(folder, "hello.py")
	config := fakeRunner(t).Configure(path, []string{"a", "b c"})

	if config.CodeLanguage != domain.CodeLanguagePython || config.Mode != domain.RunFile || config.Target != path {
		t.Fatalf("config = %+v", config)
	}
	if config.WorkingDir != folder || len(config.ProgramArgs) != 2 || config.ProgramArgs[1] != "b c" {
		t.Fatalf("config = %+v", config)
	}
	if config.Project == nil || config.Project.Root != folder || config.Project.Kind != domain.ProjectFolder || config.Project.Name != "" {
		t.Fatalf("project = %+v", config.Project)
	}
}

func TestConfigureNamesTheVenv(t *testing.T) {
	folder := fakeVenv(t)
	config := fakeRunner(t).Configure(filepath.Join(folder, "hello.py"), nil)
	if config.Project == nil || config.Project.Name != ".venv" || config.Mode == domain.RunProject {
		t.Fatalf("config = %+v", config)
	}
	if config.ProgramArgs == nil {
		t.Fatal("ProgramArgs must be [] not null")
	}
}

func TestBuildIsUnsupported(t *testing.T) {
	if err := fakeRunner(t).Build(context.Background(), domain.RunConfiguration{}); !errors.Is(err, app.ErrUnsupported) {
		t.Fatalf("Build = %v", err)
	}
}

func TestProgramJobArguments(t *testing.T) {
	interpreter := python.Interpreter{Path: "python"}
	config := domain.NewFileRunConfiguration(domain.CodeLanguagePython, "x.py", []string{"one", "two"})

	terminal := programJob(interpreter, config, nil, process.Terminal)
	if got := strings.Join(terminal.Args, " "); got != "-X utf8 x.py one two" || !terminal.Config.Echo {
		t.Fatalf("terminal job = %v echo=%v", got, terminal.Config.Echo)
	}
	pipes := programJob(interpreter, config, nil, process.Pipes)
	if got := strings.Join(pipes.Args, " "); got != "-u -X utf8 x.py one two" || pipes.Config.Echo {
		t.Fatalf("pipes job = %v echo=%v", got, pipes.Config.Echo)
	}
}

func TestMissingInterpreter(t *testing.T) {
	locator := python.NewLocator(python.Options{
		AppDir:          t.TempDir(),
		BaseEnvironment: []string{"PATH="},
		Probe:           func(context.Context, string) (string, error) { return "", errors.New("no") },
		Launcher:        func(context.Context) string { return "" },
	})
	r := &Runner{supervisor: fakeRunner(t).supervisor, locator: locator}

	err := r.Run(context.Background(), r.Configure(filepath.Join(t.TempDir(), "x.py"), nil))
	if !errors.Is(err, app.ErrToolNotFound) {
		t.Fatalf("Run = %v", err)
	}
	for _, status := range r.Tools(context.Background()) {
		if status.Source != domain.ToolMissing || status.Version != "" {
			t.Errorf("tool %s = %+v", status.ID, status)
		}
	}
}

func TestUntitledFileName(t *testing.T) {
	cases := map[string]string{
		"untitled-1.py": "untitled-1.py", "": "main.py", "notes.txt": "main.py", ".py": "main.py", "a/b.PYW": "b.PYW",
	}
	for path, want := range cases {
		if got := untitledFileName(path); got != want {
			t.Errorf("untitledFileName(%q) = %q, want %q", path, got, want)
		}
	}
}

func TestRunUntitledWritesAndCleansTheFolder(t *testing.T) {
	r, sink := realRunner(t)
	config, err := r.RunUntitled(context.Background(), "untitled-1.py", "print('from untitled')\n", nil)
	if err != nil {
		t.Fatal(err)
	}
	if filepath.Base(config.Target) != "untitled-1.py" {
		t.Fatalf("target = %s", config.Target)
	}
	dir := filepath.Dir(config.Target)
	if !strings.HasPrefix(filepath.Base(dir), "vizcacha_py_") {
		t.Fatalf("folder = %s", dir)
	}
	sink.waitOutput(t, "from untitled")
	if code := sink.waitFinished(t); code != 0 {
		t.Fatalf("exit code %d: %s", code, sink.output())
	}
	waitGone(t, dir)
}

func waitGone(t *testing.T, dir string) {
	t.Helper()
	for range 200 {
		if _, err := os.Stat(dir); os.IsNotExist(err) {
			return
		}
		sleep()
	}
	t.Fatalf("%s was not removed", dir)
}
