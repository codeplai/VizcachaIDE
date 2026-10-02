package toolchain

import (
	"context"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/codeplai/VizcachaIDE/wails/internal/app"
)

// politeProgram cleans up when it is interrupted: the defer and the signal handler must run.
const politeProgram = `package main

import (
	"fmt"
	"os"
	"os/signal"
)

func main() {
	defer fmt.Println("deferred ran")
	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt)
	fmt.Println("ready")
	<-stop
	fmt.Println("interrupt received")
}
`

// stubbornProgram ignores the interrupt, so Stop has to kill it after the grace period.
const stubbornProgram = `package main

import (
	"fmt"
	"os"
	"os/signal"
	"time"
)

func main() {
	signal.Notify(make(chan os.Signal, 1), os.Interrupt) // caught and never answered
	fmt.Println("ready")
	for {
		time.Sleep(50 * time.Millisecond)
	}
}
`

func startProgram(t *testing.T, source string) (*Toolchain, *testSink) {
	t.Helper()
	tc, sink := newTestToolchain(t)
	dir := writeFiles(t, map[string]string{"main.go": source})
	config := app.ConfigurationForFile(filepath.Join(dir, "main.go"), nil)
	if err := tc.Run(context.Background(), config); err != nil {
		t.Fatal(err)
	}
	sink.waitStdout(t, "ready")
	return tc, sink
}

func TestStopLetsTheProgramCleanUp(t *testing.T) {
	if os.Getenv("CI") != "" {
		t.Skip("console signals are not reliable on shared CI runners")
	}
	tc, sink := startProgram(t, politeProgram)

	if err := tc.Stop(); err != nil {
		t.Fatal(err)
	}
	sink.waitFinished(t)

	out := sink.Stdout()
	for _, want := range []string{"interrupt received", "deferred ran"} {
		if !containsLine(out, want) {
			t.Errorf("stdout = %q, want the line %q: Stop must interrupt before killing", out, want)
		}
	}
}

func TestStopKillsAProgramThatIgnoresTheInterrupt(t *testing.T) {
	if os.Getenv("CI") != "" {
		t.Skip("console signals are not reliable on shared CI runners")
	}
	tc, sink := startProgram(t, stubbornProgram)
	began := time.Now()

	if err := tc.Stop(); err != nil {
		t.Fatal(err)
	}
	sink.waitFinished(t)

	if took := time.Since(began); took < killGrace/2 {
		t.Errorf("the program ended after %v: it should have had the grace period first", took)
	}
	if tc.IsRunning() {
		t.Error("IsRunning must be false after the kill")
	}
}

func containsLine(output, line string) bool {
	for _, got := range splitLines(output) {
		if got == line {
			return true
		}
	}
	return false
}

func splitLines(text string) []string {
	lines := []string{}
	start := 0
	for i := 0; i < len(text); i++ {
		if text[i] != '\n' {
			continue
		}
		end := i
		if end > start && text[end-1] == '\r' {
			end--
		}
		lines = append(lines, text[start:end])
		start = i + 1
	}
	return append(lines, text[start:])
}
