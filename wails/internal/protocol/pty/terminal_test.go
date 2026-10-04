package pty_test

import (
	"bytes"
	"io"
	"runtime"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/codeplai/VizcachaIDE/wails/internal/protocol/pty"
)

type lockedBuffer struct {
	mu  sync.Mutex
	buf bytes.Buffer
}

func (b *lockedBuffer) Write(p []byte) (int, error) {
	b.mu.Lock()
	defer b.mu.Unlock()
	return b.buf.Write(p)
}

func (b *lockedBuffer) String() string {
	b.mu.Lock()
	defer b.mu.Unlock()
	return b.buf.String()
}

// echoProgram reads one line and prints it back, then prints a line longer than 120 columns.
func echoProgram() pty.Program {
	long := strings.Repeat("x", 300)
	if runtime.GOOS == "windows" {
		return pty.Program{Command: "cmd", Args: []string{"/c", "set /p name=Name: & call echo hello %name% & echo " + long}}
	}
	return pty.Program{Command: "sh", Args: []string{"-c", "printf 'Name: '; read name; echo hello $name; echo " + long}}
}

func TestTerminalReadsInputAndKeepsLongLines(t *testing.T) {
	terminal, err := pty.Start(echoProgram())
	if err != nil {
		t.Skipf("no pseudoterminal on this machine: %v", err)
	}
	defer func() { _ = terminal.Close() }()
	output := &lockedBuffer{}
	go func() { _, _ = io.Copy(output, terminal) }()
	time.Sleep(time.Second)
	if _, err := terminal.Write([]byte("Ana\r")); err != nil {
		t.Fatal(err)
	}
	done := make(chan error, 1)
	go func() { done <- terminal.Wait() }()
	select {
	case <-done:
	case <-time.After(10 * time.Second):
		t.Fatal("the program did not end")
	}
	time.Sleep(500 * time.Millisecond)
	text := pty.Clean(output.String())
	if !strings.Contains(text, "hello Ana") {
		t.Fatalf("input not read: %q", text)
	}
	if !strings.Contains(text, strings.Repeat("x", 300)) {
		t.Fatalf("long line was wrapped: %q", text)
	}
}

func TestCleanRemovesEscapeSequences(t *testing.T) {
	raw := "\x1b[2J\x1b[m\x1b[Hline\r\n\x1b]0;title\a\x1b[?25hName: Ana\x1b[17X\r\n\x1b[31mred\x1b[m\x1b[K"
	want := "line\r\nName: Ana\r\nred"
	if got := pty.Clean(raw); got != want {
		t.Fatalf("Clean = %q, want %q", got, want)
	}
}
