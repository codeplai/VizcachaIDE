package runner

import (
	"context"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/codeplai/VizcachaIDE/wails/internal/adapters/python"
	"github.com/codeplai/VizcachaIDE/wails/internal/adapters/python/pythontest"
	"github.com/codeplai/VizcachaIDE/wails/internal/app"
	"github.com/codeplai/VizcachaIDE/wails/internal/domain"
	"github.com/codeplai/VizcachaIDE/wails/internal/protocol/process"
)

const waitLimit = 60 * time.Second

// testSink records the run events. The other EventSink methods are never called here.
type testSink struct {
	app.EventSink
	mu       sync.Mutex
	stdout   strings.Builder
	stderr   strings.Builder
	started  []domain.RunConfiguration
	finished chan int
}

func newTestSink() *testSink { return &testSink{finished: make(chan int, 4)} }

func (s *testSink) RunOutput(stream, text string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if stream == "stderr" {
		s.stderr.WriteString(text)
		return
	}
	s.stdout.WriteString(text)
}

func (s *testSink) RunStarted(config domain.RunConfiguration) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.started = append(s.started, config)
}

func (s *testSink) RunFinished(exitCode int, _ int64) { s.finished <- exitCode }

func (s *testSink) output() string {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.stdout.String() + s.stderr.String()
}

func (s *testSink) waitOutput(t *testing.T, text string) {
	t.Helper()
	deadline := time.Now().Add(waitLimit)
	for time.Now().Before(deadline) {
		if strings.Contains(s.output(), text) {
			return
		}
		time.Sleep(20 * time.Millisecond)
	}
	t.Fatalf("output never contained %q: %q", text, s.output())
}

func (s *testSink) waitFinished(t *testing.T) int {
	t.Helper()
	select {
	case code := <-s.finished:
		return code
	case <-time.After(waitLimit):
		t.Fatal("the program did not finish in time")
		return -1
	}
}

// settings is a fixed python path chosen by the user.
type settings struct{ path string }

func (s settings) Load() (domain.Settings, error) {
	loaded := domain.DefaultSettings()
	loaded.ToolPaths = map[string]string{python.ToolPython: s.path}
	return loaded, nil
}
func (settings) Save(domain.Settings) error { return nil }

// realRunner runs with the test interpreter, skipping the test without one.
func realRunner(t *testing.T) (*Runner, *testSink) {
	t.Helper()
	path := pythontest.Interpreter(t)
	sink := newTestSink()
	r := New(process.New(sink), Options{Settings: settings{path: path}, AppDir: t.TempDir()})
	return r, sink
}

// fakeRunner finds only what the fake probe accepts: no PATH, no launcher, no toolchain.
func fakeRunner(t *testing.T) *Runner {
	t.Helper()
	locator := python.NewLocator(python.Options{
		AppDir:          t.TempDir(),
		BaseEnvironment: []string{"PATH="},
		Probe:           func(context.Context, string) (string, error) { return "3.12.0", nil },
		Launcher:        func(context.Context) string { return "" },
	})
	return &Runner{supervisor: process.New(newTestSink()), locator: locator, base: []string{}}
}

// fakeVenv creates folder/.venv with an empty interpreter file and returns the folder.
func fakeVenv(t *testing.T) string {
	t.Helper()
	folder := t.TempDir()
	dir, name := filepath.Join(folder, ".venv", "bin"), "python"
	if runtime.GOOS == "windows" {
		dir, name = filepath.Join(folder, ".venv", "Scripts"), "python.exe"
	}
	if err := os.MkdirAll(dir, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(dir, name), nil, 0o755); err != nil {
		t.Fatal(err)
	}
	return folder
}

func writeProgram(t *testing.T, source string) string {
	t.Helper()
	path := filepath.Join(t.TempDir(), "program.py")
	if err := os.WriteFile(path, []byte(source), 0o600); err != nil {
		t.Fatal(err)
	}
	return path
}
