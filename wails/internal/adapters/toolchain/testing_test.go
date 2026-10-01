package toolchain

import (
	"context"
	"os"
	"os/exec"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/codeplai/VizcachaIDE/wails/internal/app"
	"github.com/codeplai/VizcachaIDE/wails/internal/domain"
)

const waitLimit = 90 * time.Second

type finishedEvent struct {
	ExitCode   int
	DurationMs int64
}

// testSink records the run events. The other EventSink methods are never called here.
type testSink struct {
	app.EventSink
	mu       sync.Mutex
	stdout   strings.Builder
	stderr   strings.Builder
	started  []domain.RunConfiguration
	finished chan finishedEvent
}

func newTestSink() *testSink { return &testSink{finished: make(chan finishedEvent, 8)} }

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

func (s *testSink) RunFinished(exitCode int, durationMs int64) {
	s.finished <- finishedEvent{exitCode, durationMs}
}

func (s *testSink) Stdout() string {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.stdout.String()
}

func (s *testSink) Stderr() string {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.stderr.String()
}

func (s *testSink) waitFinished(t *testing.T) finishedEvent {
	t.Helper()
	select {
	case event := <-s.finished:
		return event
	case <-time.After(waitLimit):
		t.Fatal("the process did not finish in time")
		return finishedEvent{}
	}
}

func (s *testSink) waitStdout(t *testing.T, want string) {
	t.Helper()
	deadline := time.Now().Add(waitLimit)
	for time.Now().Before(deadline) {
		if strings.Contains(s.Stdout(), want) {
			return
		}
		time.Sleep(20 * time.Millisecond)
	}
	t.Fatalf("stdout never contained %q, got %q", want, s.Stdout())
}

// requireGo skips the test when there is no Go on this machine.
func requireGo(t *testing.T) {
	t.Helper()
	if _, err := exec.LookPath("go"); err != nil {
		t.Skip("go is not installed")
	}
}

func newTestToolchain(t *testing.T) (*Toolchain, *testSink) {
	t.Helper()
	requireGo(t)
	sink := newTestSink()
	tc := New(Options{Sink: sink, AppDir: t.TempDir()})
	return tc, sink
}

// writeFiles creates the files (name to content) in a new temporary folder.
func writeFiles(t *testing.T, files map[string]string) string {
	t.Helper()
	dir := t.TempDir()
	for name, content := range files {
		if err := os.WriteFile(dir+string(os.PathSeparator)+name, []byte(content), 0o600); err != nil {
			t.Fatal(err)
		}
	}
	return dir
}

func runAndWait(t *testing.T, tc *Toolchain, sink *testSink, config domain.RunConfiguration) finishedEvent {
	t.Helper()
	if err := tc.Run(context.Background(), config); err != nil {
		t.Fatalf("Run: %v", err)
	}
	return sink.waitFinished(t)
}
