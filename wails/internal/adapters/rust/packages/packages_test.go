package packages_test

import (
	"context"
	"errors"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/codeplai/VizcachaIDE/wails/internal/adapters/rust/packages"
	"github.com/codeplai/VizcachaIDE/wails/internal/adapters/rust/runner"
	"github.com/codeplai/VizcachaIDE/wails/internal/adapters/rust/rusttest"
	"github.com/codeplai/VizcachaIDE/wails/internal/app"
	"github.com/codeplai/VizcachaIDE/wails/internal/domain"
	"github.com/codeplai/VizcachaIDE/wails/internal/protocol/process"
)

type sink struct {
	app.EventSink
	mu       sync.Mutex
	output   strings.Builder
	finished chan int
}

func (s *sink) RunOutput(_, text string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.output.WriteString(text)
}

func (s *sink) RunStarted(domain.RunConfiguration) {}

func (s *sink) RunFinished(exitCode int, _ int64) { s.finished <- exitCode }

func (s *sink) text() string {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.output.String()
}

// wait returns the exit code and what the command printed.
func (s *sink) wait(t *testing.T) (int, string) {
	t.Helper()
	select {
	case code := <-s.finished:
		return code, s.text()
	case <-time.After(120 * time.Second):
		t.Fatal("cargo did not finish in time")
		return -1, ""
	}
}

func newManager(t *testing.T) (*packages.Manager, *sink) {
	t.Helper()
	events := &sink{finished: make(chan int, 4)}
	supervisor := process.New(events)
	rustRunner := runner.New(supervisor, runner.Options{BaseEnvironment: rusttest.Environment(t), AppDir: t.TempDir()})
	return packages.New(supervisor, rustRunner), events
}

func read(t *testing.T, path string) string {
	t.Helper()
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	return string(data)
}

func do(t *testing.T, events *sink, call func() error) string {
	t.Helper()
	if err := call(); err != nil {
		t.Fatal(err)
	}
	code, output := events.wait(t)
	if code != 0 {
		t.Fatalf("cargo exit %d: %s", code, output)
	}
	return output
}

func TestInitAddListAndRemoveRunCargoCommands(t *testing.T) {
	manager, events := newManager(t)
	ctx := context.Background()
	dir := filepath.Join(t.TempDir(), "Mi Programa")
	if err := os.MkdirAll(dir, 0o755); err != nil {
		t.Fatal(err)
	}

	do(t, events, func() error { return manager.Init(ctx, dir, "") })
	manifest := filepath.Join(dir, "Cargo.toml")
	if !strings.Contains(read(t, manifest), `name = "mi_programa"`) {
		t.Fatalf("Cargo.toml = %s", read(t, manifest))
	}
	if _, err := os.Stat(filepath.Join(dir, ".git")); err == nil {
		t.Fatal("init must not create a git repository")
	}
	if err := manager.Init(ctx, dir, "otra"); !errors.Is(err, packages.ErrAlreadyProject) {
		t.Fatalf("second init: %v, want ErrAlreadyProject", err)
	}

	do(t, events, func() error { return manager.Add(ctx, dir, "itoa") })
	if !strings.Contains(read(t, manifest), "itoa") {
		t.Fatalf("Cargo.toml = %s", read(t, manifest))
	}
	// The active file of the crate is as good as its folder.
	if listing := do(t, events, func() error { return manager.List(ctx, filepath.Join(dir, "src", "main.rs")) }); !strings.Contains(listing, "itoa") {
		t.Fatalf("list = %q, want itoa", listing)
	}
	do(t, events, func() error { return manager.Remove(ctx, dir, "itoa") })
	if strings.Contains(read(t, manifest), "itoa") {
		t.Fatalf("Cargo.toml = %s", read(t, manifest))
	}
}

func TestWorkspaceCommandsActOnTheMemberOfTheFile(t *testing.T) {
	manager, events := newManager(t)
	root := t.TempDir()
	member := filepath.Join(root, "crates", "uno")
	for name, content := range map[string]string{
		"Cargo.toml":             "[workspace]\nmembers = [\"crates/*\"]\nresolver = \"2\"\n",
		"crates/uno/Cargo.toml":  "[package]\nname = \"uno\"\nversion = \"0.1.0\"\nedition = \"2021\"\n",
		"crates/uno/src/main.rs": "fn main() {}\n",
		"crates/otro/Cargo.toml": "[package]\nname = \"otro\"\nversion = \"0.1.0\"\nedition = \"2021\"\n",
		"crates/otro/src/lib.rs": "",
	} {
		path := filepath.Join(root, filepath.FromSlash(name))
		if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(path, []byte(content), 0o600); err != nil {
			t.Fatal(err)
		}
	}
	do(t, events, func() error {
		return manager.Add(context.Background(), filepath.Join(member, "src", "main.rs"), "itoa")
	})
	if !strings.Contains(read(t, filepath.Join(member, "Cargo.toml")), "itoa") ||
		strings.Contains(read(t, filepath.Join(root, "crates", "otro", "Cargo.toml")), "itoa") {
		t.Fatal("itoa must be added to the member of the file only")
	}
}

func TestOutsideACrateTheVerbsAskForAProject(t *testing.T) {
	manager, _ := newManager(t)
	dir := t.TempDir()
	ctx := context.Background()
	for name, call := range map[string]func() error{
		"add":    func() error { return manager.Add(ctx, dir, "itoa") },
		"remove": func() error { return manager.Remove(ctx, dir, "itoa") },
		"list":   func() error { return manager.List(ctx, dir) },
	} {
		if err := call(); !errors.Is(err, packages.ErrNeedsProject) || !strings.Contains(err.Error(), "errors.rustNeedsProject") {
			t.Errorf("%s: %v, want ErrNeedsProject", name, err)
		}
	}
	if err := manager.Tidy(ctx, dir); !errors.Is(err, app.ErrUnsupported) {
		t.Errorf("tidy: %v, want ErrUnsupported", err)
	}
	if err := manager.Add(ctx, dir, "--git"); !errors.Is(err, app.ErrInvalidArgument) {
		t.Errorf("a flag as package: %v", err)
	}
}

func TestCrateNameIsAlwaysValid(t *testing.T) {
	tests := map[string]string{
		"Mi Programa": "mi_programa", "hola-mundo": "hola-mundo", "3d": "app_3d", "test": "app_test",
		"___": "app", "": "app", "Ñandú": "and",
	}
	for folder, want := range tests {
		if got := packages.CrateName(folder); got != want {
			t.Errorf("%q: %q, want %q", folder, got, want)
		}
	}
}
