package runner

import (
	"os"
	"path/filepath"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/codeplai/VizcachaIDE/wails/internal/adapters/rust/rusttest"
	"github.com/codeplai/VizcachaIDE/wails/internal/app"
	"github.com/codeplai/VizcachaIDE/wails/internal/domain"
	"github.com/codeplai/VizcachaIDE/wails/internal/protocol/process"
)

const waitLimit = 120 * time.Second

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

func (s *testSink) text() (stdout, stderr string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.stdout.String(), s.stderr.String()
}

// all is everything the run printed, whatever the stream.
func (s *testSink) all() string {
	out, errText := s.text()
	return out + errText
}

func (s *testSink) waitOutput(t *testing.T, text string) {
	t.Helper()
	deadline := time.Now().Add(waitLimit)
	for time.Now().Before(deadline) {
		if strings.Contains(s.all(), text) {
			return
		}
		time.Sleep(20 * time.Millisecond)
	}
	t.Fatalf("output never contained %q: %q", text, s.all())
}

func (s *testSink) waitFinished(t *testing.T) int {
	t.Helper()
	select {
	case code := <-s.finished:
		return code
	case <-time.After(waitLimit):
		t.Fatal("the run did not finish in time")
		return -1
	}
}

// newRunner is a runner on the real toolchain (the test is skipped without it).
func newRunner(t *testing.T) (*Runner, *testSink) {
	t.Helper()
	env := rusttest.Environment(t)
	sink := newTestSink()
	r := New(process.New(sink), Options{
		BaseEnvironment: env, AppDir: t.TempDir(), CacheDir: t.TempDir(),
		CompilingNotice: func() string { return "Compiling..." },
	})
	return r, sink
}

// write creates a file (and its folders) under root.
func write(t *testing.T, root, name, content string) string {
	t.Helper()
	path := filepath.Join(root, filepath.FromSlash(name))
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path, []byte(content), 0o600); err != nil {
		t.Fatal(err)
	}
	return path
}

// looseFile writes main.rs in a folder with spaces and an accent, like a student's.
func looseFile(t *testing.T, source string) string {
	t.Helper()
	return write(t, filepath.Join(t.TempDir(), "mis programas ñandú"), "main.rs", source)
}

// crate writes a Cargo.toml with the package name and the sources of files (path → content).
func crate(t *testing.T, root, name string, files map[string]string) {
	t.Helper()
	write(t, root, "Cargo.toml", "[package]\nname = \""+name+"\"\nversion = \"0.1.0\"\nedition = \"2021\"\n")
	for file, content := range files {
		write(t, root, file, content)
	}
}

const hello = "fn main() { println!(\"hola\"); }\n"
