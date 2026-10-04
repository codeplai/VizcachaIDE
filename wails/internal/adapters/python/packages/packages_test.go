package packages_test

import (
	"context"
	"errors"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/codeplai/VizcachaIDE/wails/internal/adapters/python"
	"github.com/codeplai/VizcachaIDE/wails/internal/adapters/python/packages"
	"github.com/codeplai/VizcachaIDE/wails/internal/adapters/python/pythontest"
	"github.com/codeplai/VizcachaIDE/wails/internal/adapters/python/runner"
	"github.com/codeplai/VizcachaIDE/wails/internal/app"
	"github.com/codeplai/VizcachaIDE/wails/internal/domain"
	"github.com/codeplai/VizcachaIDE/wails/internal/protocol/process"
)

type sink struct {
	app.EventSink
	mu       sync.Mutex
	stdout   strings.Builder
	started  []domain.RunConfiguration
	finished chan int
}

func (s *sink) RunOutput(_, text string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.stdout.WriteString(text)
}

func (s *sink) RunStarted(config domain.RunConfiguration) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.started = append(s.started, config)
}

func (s *sink) RunFinished(exitCode int, _ int64) { s.finished <- exitCode }

type settings struct{ path string }

func (s settings) Load() (domain.Settings, error) {
	loaded := domain.DefaultSettings()
	loaded.ToolPaths = map[string]string{python.ToolPython: s.path}
	return loaded, nil
}
func (settings) Save(domain.Settings) error { return nil }

func newManager(t *testing.T, path string) (*packages.Manager, *sink) {
	t.Helper()
	events := &sink{finished: make(chan int, 4)}
	supervisor := process.New(events)
	pythonRunner := runner.New(supervisor, runner.Options{Settings: settings{path: path}, AppDir: t.TempDir()})
	return packages.New(supervisor, pythonRunner), events
}

func TestUnsupportedVerbs(t *testing.T) {
	manager, _ := newManager(t, "")
	if err := manager.Init(context.Background(), t.TempDir(), "x"); !errors.Is(err, app.ErrUnsupported) {
		t.Errorf("Init = %v", err)
	}
	if err := manager.Tidy(context.Background(), t.TempDir()); !errors.Is(err, app.ErrUnsupported) {
		t.Errorf("Tidy = %v", err)
	}
}

func TestPackageNameIsValidated(t *testing.T) {
	manager, events := newManager(t, "")
	for _, pkg := range []string{"", "  ", "two words", "-r", "requests; rm -rf /"} {
		if err := manager.Add(context.Background(), t.TempDir(), pkg); err == nil || errors.Is(err, app.ErrToolNotFound) {
			t.Errorf("Add(%q) = %v, want a validation error", pkg, err)
		}
		if err := manager.Remove(context.Background(), t.TempDir(), pkg); err == nil || errors.Is(err, app.ErrToolNotFound) {
			t.Errorf("Remove(%q) = %v, want a validation error", pkg, err)
		}
	}
	if len(events.started) != 0 {
		t.Fatal("no command should have started")
	}
}

func TestMissingPython(t *testing.T) {
	manager, events := newManager(t, "")
	err := manager.List(context.Background(), t.TempDir())
	if err != nil && !errors.Is(err, app.ErrToolNotFound) {
		t.Fatalf("List = %v", err)
	}
	if err == nil { // a Python on PATH ran pip: wait for it, Windows cannot delete a folder in use
		select {
		case <-events.finished:
		case <-time.After(90 * time.Second):
			t.Fatal("pip list did not finish in time")
		}
	}
}

func TestListRunsPipListWithEvents(t *testing.T) {
	manager, events := newManager(t, pythontest.Interpreter(t))
	if err := manager.List(context.Background(), t.TempDir()); err != nil {
		t.Fatal(err)
	}
	select {
	case code := <-events.finished:
		if code != 0 {
			t.Fatalf("pip list exit code %d: %s", code, events.stdout.String())
		}
	case <-time.After(90 * time.Second):
		t.Fatal("pip list did not finish in time")
	}
	events.mu.Lock()
	defer events.mu.Unlock()
	if !strings.Contains(strings.ToLower(events.stdout.String()), "debugpy") {
		t.Fatalf("pip list output = %q", events.stdout.String())
	}
	if len(events.started) != 1 || events.started[0].Mode != domain.RunProject || events.started[0].Echo {
		t.Fatalf("started = %+v", events.started)
	}
}
