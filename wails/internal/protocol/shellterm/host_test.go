package shellterm

import (
	"context"
	"os"
	"os/exec"
	"runtime"
	"strings"
	"sync"
	"testing"
	"time"
)

type recordingSink struct {
	mu      sync.Mutex
	output  map[string]string
	exits   map[string]int
	exitSet map[string]bool
}

func newRecordingSink() *recordingSink {
	return &recordingSink{output: map[string]string{}, exits: map[string]int{}, exitSet: map[string]bool{}}
}

func (s *recordingSink) TerminalOutput(id, data string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.output[id] += data
}

func (s *recordingSink) TerminalExit(id string, code int) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.exits[id], s.exitSet[id] = code, true
}

func (s *recordingSink) text(id string) string {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.output[id]
}

func (s *recordingSink) exited(id string) bool {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.exitSet[id]
}

func eventually(t *testing.T, what string, ok func() bool) {
	t.Helper()
	deadline := time.Now().Add(20 * time.Second)
	for time.Now().Before(deadline) {
		if ok() {
			return
		}
		time.Sleep(50 * time.Millisecond)
	}
	t.Fatalf("timeout waiting for %s", what)
}

func testShell() Command {
	if runtime.GOOS == "windows" {
		return ChooseShell(runtime.GOOS, exec.LookPath, os.Getenv)
	}
	return Command{Path: "/bin/sh"}
}

func newTestHost(t *testing.T, sink *recordingSink, folders ...string) *Host {
	t.Helper()
	host := NewHost(Options{
		Sink:    sink,
		Folders: func(context.Context) []string { return folders },
		Shell:   testShell,
	})
	t.Cleanup(func() {
		host.CloseAll()
		eventually(t, "every session to end", func() bool {
			host.mu.Lock()
			defer host.mu.Unlock()
			return len(host.live) == 0
		})
	})
	return host
}

func TestSessionRunsCommandsResizesAndCloses(t *testing.T) {
	sink := newRecordingSink()
	dir := t.TempDir()
	host := newTestHost(t, sink)
	id, err := host.Start(dir, 100, 30)
	if err != nil {
		t.Skipf("no pseudoterminal on this machine: %v", err)
	}
	// The sum keeps the answer apart from the echo of what was typed.
	command := "echo vizcacha-$((20+1))"
	if runtime.GOOS == "windows" {
		command = "echo ('vizcacha-' + (20+1))"
	}
	if err := host.Write(id, command); err != nil {
		t.Fatal(err)
	}
	eventually(t, "the echo output", func() bool { return strings.Contains(sink.text(id), "vizcacha-21") })
	if err := host.Resize(id, 120, 40); err != nil {
		t.Fatalf("resize: %v", err)
	}
	if err := host.Close(id); err != nil {
		t.Fatal(err)
	}
	eventually(t, "the session to disappear", func() bool { _, err := host.find(id); return err != nil })
	if sink.exited(id) {
		t.Error("a session the user closed must not report an exit")
	}
	if err := host.Write(id, "x"); err == nil {
		t.Error("writing to a closed session must fail")
	}
}

func TestShellExitReportsTheExitCode(t *testing.T) {
	sink := newRecordingSink()
	dir := t.TempDir()
	host := newTestHost(t, sink)
	id, err := host.Start(dir, 80, 24)
	if err != nil {
		t.Skipf("no pseudoterminal on this machine: %v", err)
	}
	if err := host.Write(id, "exit 3\r"); err != nil {
		t.Fatal(err)
	}
	eventually(t, "the exit event", func() bool { return sink.exited(id) })
	if sink.exits[id] != 3 {
		t.Errorf("exit code = %d, want 3", sink.exits[id])
	}
}

func TestSeveralSessionsAreIndependentAndStartInTheFolder(t *testing.T) {
	sink := newRecordingSink()
	dir := t.TempDir()
	host := newTestHost(t, sink)
	first, err := host.Start(dir, 80, 24)
	if err != nil {
		t.Skipf("no pseudoterminal on this machine: %v", err)
	}
	second, err := host.Start(dir, 80, 24)
	if err != nil || first == second {
		t.Fatalf("second session: %q %v", second, err)
	}
	pwd := "pwd\r"
	if runtime.GOOS == "windows" {
		pwd = "(Get-Location).Path\r"
	}
	for _, id := range []string{first, second} {
		if err := host.Write(id, pwd); err != nil {
			t.Fatal(err)
		}
	}
	base := dir[strings.LastIndexAny(dir, `\/`)+1:]
	for _, id := range []string{first, second} {
		eventually(t, "the folder in "+id, func() bool { return strings.Contains(sink.text(id), base) })
	}
}

func TestStartFallsBackToHomeForAMissingFolder(t *testing.T) {
	home, err := os.UserHomeDir()
	if err != nil {
		t.Skip("no home folder")
	}
	if got := startDirectory(`Z:\does\not\exist`); got != home {
		t.Errorf("got %q, want %q", got, home)
	}
}
