package runner

import (
	"context"
	"errors"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/codeplai/VizcachaIDE/wails/internal/adapters/golang"
	"github.com/codeplai/VizcachaIDE/wails/internal/app"
)

// beatProgram prints "ready" and then appends to a file every 20 ms, forever.
const beatProgram = `package main

import (
	"fmt"
	"os"
	"time"
)

func main() {
	fmt.Println("ready")
	for {
		f, err := os.OpenFile(os.Args[1], os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0o600)
		if err == nil {
			f.WriteString("x")
			f.Close()
		}
		time.Sleep(20 * time.Millisecond)
	}
}
`

func fileSize(t *testing.T, path string) int64 {
	t.Helper()
	info, err := os.Stat(path)
	if err != nil {
		t.Fatal(err)
	}
	return info.Size()
}

func startBeat(t *testing.T) (*Runner, *testSink, string) {
	t.Helper()
	tc, sink := newTestRunner(t)
	dir := writeFiles(t, map[string]string{"main.go": beatProgram})
	beat := filepath.Join(dir, "beat.txt")
	config := golang.ConfigurationForFile(filepath.Join(dir, "main.go"), []string{beat})
	if err := tc.Run(context.Background(), config); err != nil {
		t.Fatal(err)
	}
	sink.waitStdout(t, "ready")
	return tc, sink, beat
}

func TestStopEndsTheWholeProcessTree(t *testing.T) {
	tc, sink, beat := startBeat(t)

	if err := tc.Stop(); err != nil {
		t.Fatal(err)
	}
	sink.waitFinished(t)

	// If the user's program (a child of "go run") survived, the file would keep growing.
	before := fileSize(t, beat)
	time.Sleep(500 * time.Millisecond)
	if after := fileSize(t, beat); after != before {
		t.Errorf("the program is still alive: beat file grew from %d to %d bytes", before, after)
	}
	if tc.IsRunning() {
		t.Error("IsRunning must be false after Stop")
	}
}

func TestSecondRunWhileRunningIsBusy(t *testing.T) {
	tc, sink, _ := startBeat(t)
	defer func() { _ = tc.Stop(); sink.waitFinished(t) }()

	err := tc.Run(context.Background(), golang.ConfigurationForFile("other.go", nil))
	if !errors.Is(err, app.ErrBusy) {
		t.Errorf("Run error = %v, want ErrBusy", err)
	}
	_, err = tc.RunUntitled(context.Background(), "untitled-1.go", helloProgram, nil)
	if !errors.Is(err, app.ErrBusy) {
		t.Errorf("RunUntitled error = %v, want ErrBusy", err)
	}
}

func TestCancelingTheContextStopsTheProgram(t *testing.T) {
	tc, sink := newTestRunner(t)
	dir := writeFiles(t, map[string]string{"main.go": beatProgram})
	ctx, cancel := context.WithCancel(context.Background())
	config := golang.ConfigurationForFile(filepath.Join(dir, "main.go"), []string{filepath.Join(dir, "b.txt")})
	if err := tc.Run(ctx, config); err != nil {
		t.Fatal(err)
	}
	sink.waitStdout(t, "ready")

	cancel()

	sink.waitFinished(t)
}
