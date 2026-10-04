package packages_test

import (
	"context"
	"errors"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/codeplai/VizcachaIDE/wails/internal/adapters/golang"
	"github.com/codeplai/VizcachaIDE/wails/internal/adapters/golang/packages"
	"github.com/codeplai/VizcachaIDE/wails/internal/adapters/golang/runner"
	"github.com/codeplai/VizcachaIDE/wails/internal/app"
	"github.com/codeplai/VizcachaIDE/wails/internal/domain"
	"github.com/codeplai/VizcachaIDE/wails/internal/protocol/process"
)

type sink struct {
	app.EventSink
	mu       sync.Mutex
	stderr   strings.Builder
	started  int
	finished chan int
}

func (s *sink) RunOutput(stream, text string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if stream == "stderr" {
		s.stderr.WriteString(text)
	}
}

func (s *sink) RunStarted(domain.RunConfiguration) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.started++
}

func (s *sink) RunFinished(exitCode int, _ int64) { s.finished <- exitCode }

func (s *sink) wait(t *testing.T) int {
	t.Helper()
	select {
	case code := <-s.finished:
		return code
	case <-time.After(90 * time.Second):
		t.Fatal("the go command did not finish in time")
		return -1
	}
}

func newManager(t *testing.T) (*packages.Manager, *sink) {
	t.Helper()
	if _, err := exec.LookPath("go"); err != nil {
		t.Skip("go is not installed")
	}
	events := &sink{finished: make(chan int, 4)}
	supervisor := process.New(events)
	goRunner := runner.New(supervisor, runner.Options{AppDir: t.TempDir()})
	return packages.New(supervisor, goRunner), events
}

func TestInitAndTidyRunGoModCommands(t *testing.T) {
	manager, events := newManager(t)
	dir := t.TempDir()
	if err := os.WriteFile(filepath.Join(dir, "main.go"), []byte("package main\n\nfunc main() {}\n"), 0o600); err != nil {
		t.Fatal(err)
	}

	if err := manager.Init(context.Background(), dir, "example.com/hola"); err != nil {
		t.Fatal(err)
	}
	if code := events.wait(t); code != 0 {
		t.Fatalf("go mod init exit %d: %s", code, events.stderr.String())
	}
	if err := manager.Tidy(context.Background(), dir); err != nil {
		t.Fatal(err)
	}
	if code := events.wait(t); code != 0 {
		t.Fatalf("go mod tidy exit %d: %s", code, events.stderr.String())
	}

	text, err := app.ReadSourceFile(filepath.Join(dir, "go.mod"))
	if err != nil || golang.ParseModulePath(text) != "example.com/hola" || events.started != 2 {
		t.Errorf("go.mod = %q, %v, run:started %d", text, err, events.started)
	}
}

func TestArgumentsAreValidatedBeforeRunning(t *testing.T) {
	manager, events := newManager(t)
	for _, bad := range []string{"", "-u", "two words"} {
		if err := manager.Add(context.Background(), t.TempDir(), bad); !errors.Is(err, app.ErrInvalidArgument) {
			t.Errorf("Add(%q) error = %v, want ErrInvalidGoArgument", bad, err)
		}
		if err := manager.Init(context.Background(), t.TempDir(), bad); !errors.Is(err, app.ErrInvalidArgument) {
			t.Errorf("Init(%q) error = %v, want ErrInvalidGoArgument", bad, err)
		}
	}
	if events.started != 0 {
		t.Error("nothing may run with an invalid argument")
	}
}

func TestRemoveAndListAreUnsupported(t *testing.T) {
	manager, _ := newManager(t)
	if err := manager.Remove(context.Background(), "dir", "pkg"); !errors.Is(err, app.ErrUnsupported) {
		t.Errorf("Remove error = %v", err)
	}
	if err := manager.List(context.Background(), "dir"); !errors.Is(err, app.ErrUnsupported) {
		t.Errorf("List error = %v", err)
	}
}

func TestSharesTheRunSlotWithTheRunner(t *testing.T) {
	manager, events := newManager(t)
	dir := t.TempDir()
	if err := manager.Init(context.Background(), dir, "example.com/slot"); err != nil {
		t.Fatal(err)
	}

	err := manager.Tidy(context.Background(), dir)

	if !errors.Is(err, app.ErrBusy) {
		t.Errorf("second command error = %v, want ErrBusy", err)
	}
	events.wait(t)
}
