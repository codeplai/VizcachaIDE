package process_test

import (
	"bufio"
	"fmt"
	"os"
	"os/exec"
	"os/signal"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/codeplai/VizcachaIDE/wails/internal/app"
	"github.com/codeplai/VizcachaIDE/wails/internal/domain"
	"github.com/codeplai/VizcachaIDE/wails/internal/protocol/process"
)

const waitLimit = 30 * time.Second

type finishedEvent struct {
	ExitCode   int
	DurationMs int64
}

// recordingSink records the run events. The other EventSink methods are never called here.
type recordingSink struct {
	app.EventSink
	mu       sync.Mutex
	stdout   strings.Builder
	stderr   strings.Builder
	started  []domain.RunConfiguration
	finished chan finishedEvent
}

func newSink() *recordingSink { return &recordingSink{finished: make(chan finishedEvent, 8)} }

func (s *recordingSink) RunOutput(stream, text string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if stream == "stderr" {
		s.stderr.WriteString(text)
		return
	}
	s.stdout.WriteString(text)
}

func (s *recordingSink) RunStarted(config domain.RunConfiguration) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.started = append(s.started, config)
}

func (s *recordingSink) RunFinished(exitCode int, durationMs int64) {
	s.finished <- finishedEvent{exitCode, durationMs}
}

func (s *recordingSink) Stdout() string {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.stdout.String()
}

func (s *recordingSink) Stderr() string {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.stderr.String()
}

func (s *recordingSink) startedCount() int {
	s.mu.Lock()
	defer s.mu.Unlock()
	return len(s.started)
}

func (s *recordingSink) waitFinished(t *testing.T) finishedEvent {
	t.Helper()
	select {
	case event := <-s.finished:
		return event
	case <-time.After(waitLimit):
		t.Fatal("the process did not finish in time")
		return finishedEvent{}
	}
}

func (s *recordingSink) waitStdout(t *testing.T, want string) {
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

// helperJob runs this test binary as a small program: see TestHelperProcess for its modes.
func helperJob(mode string, args ...string) process.Job {
	all := append([]string{"-test.run=^TestHelperProcess$", "--", mode}, args...)
	return process.Job{
		Config:  domain.RunConfiguration{Target: mode, ProgramArgs: []string{}},
		Command: os.Args[0],
		Args:    all,
		Dir:     os.TempDir(),
	}
}

// TestHelperProcess is not a test: the jobs above run it as a child process.
func TestHelperProcess(t *testing.T) {
	index := -1
	for i, arg := range os.Args {
		if arg == "--" {
			index = i
		}
	}
	if index < 0 || index+1 >= len(os.Args) {
		t.Skip("helper for the other tests")
	}
	mode, args := os.Args[index+1], os.Args[index+2:]
	switch mode {
	case "print":
		fmt.Println(args[0])
		os.Exit(parseExit(args[1:]))
	case "chatty": // prints at once and stays alive past the notice delay used by the test
		fmt.Println("hi")
		time.Sleep(1500 * time.Millisecond)
	case "silent":
		time.Sleep(600 * time.Millisecond)
		fmt.Println("done")
	case "echo":
		line, _ := bufio.NewReader(os.Stdin).ReadString('\n')
		fmt.Println("got:" + strings.TrimSpace(line))
	case "beat":
		beat(args[0])
	case "parent":
		parent(args[0])
	}
	os.Exit(0)
}

func parseExit(args []string) int {
	code := 0
	if len(args) > 0 {
		_, _ = fmt.Sscan(args[0], &code)
	}
	return code
}

// beat appends to a file every 20 ms and ignores the interrupt, so only a kill ends it.
func beat(path string) {
	signal.Notify(make(chan os.Signal, 1), os.Interrupt)
	fmt.Println("ready")
	for {
		if file, err := os.OpenFile(path, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0o600); err == nil {
			_, _ = file.WriteString("x")
			_ = file.Close()
		}
		time.Sleep(20 * time.Millisecond)
	}
}

// parent starts a beating child and waits for it, like "go run" does with the user's program.
func parent(path string) {
	signal.Notify(make(chan os.Signal, 1), os.Interrupt)
	child := exec.Command(os.Args[0], "-test.run=^TestHelperProcess$", "--", "beat", path)
	child.Stdout = os.Stdout
	_ = child.Run()
}
