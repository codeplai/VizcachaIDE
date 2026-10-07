package pty_test

import (
	"runtime"
	"testing"
	"time"

	"github.com/codeplai/VizcachaIDE/wails/internal/protocol/pty"
)

// A program that never ends on its own must stop on Ctrl+C. On Windows this failed when the IDE
// had inherited the "ignore Ctrl+C" flag (a loop in the integrated terminal could not be stopped).
func TestCtrlCStopsAProgramThatNeverEnds(t *testing.T) {
	program := pty.Program{Command: "sleep", Args: []string{"60"}, Interactive: true}
	if runtime.GOOS == "windows" {
		program = pty.Program{Command: "ping", Args: []string{"-t", "127.0.0.1"}, Interactive: true}
	}
	terminal, err := pty.Start(program)
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = terminal.Close() }()
	go func() { _, _ = drain(terminal) }()
	time.Sleep(time.Second)
	if err := terminal.Interrupt(); err != nil {
		t.Fatal(err)
	}
	ended := make(chan struct{})
	go func() { _ = terminal.Wait(); close(ended) }()
	select {
	case <-ended:
	case <-time.After(10 * time.Second):
		t.Fatal("the program kept running after Ctrl+C")
	}
}

// drain keeps reading the output so the program never blocks on a full pipe.
func drain(terminal *pty.Terminal) (int, error) {
	buffer := make([]byte, 4096)
	total := 0
	for {
		n, err := terminal.Read(buffer)
		total += n
		if err != nil {
			return total, err
		}
	}
}
