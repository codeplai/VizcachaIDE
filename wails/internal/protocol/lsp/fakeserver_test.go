package lsp

import (
	"context"
	"os"
	"path/filepath"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"go.lsp.dev/jsonrpc2"
)

const fakeServerVariable = "VIZCACHA_FAKE_LSP"

// TestMain turns the test binary into a minimal language server when the variable is set:
// it answers every request and exits on "exit".
func TestMain(m *testing.M) {
	if os.Getenv(fakeServerVariable) == "1" {
		runFakeServer()
		return
	}
	os.Exit(m.Run())
}

type fakeStdio struct {
	*os.File
	out *os.File
}

func (f fakeStdio) Write(p []byte) (int, error) { return f.out.Write(p) }

func runFakeServer() {
	conn := jsonrpc2.NewConn(jsonrpc2.NewStream(fakeStdio{File: os.Stdin, out: os.Stdout}))
	conn.Go(context.Background(), func(ctx context.Context, reply jsonrpc2.Replier, req jsonrpc2.Request) error {
		if req.Method() == "exit" {
			os.Exit(0)
		}
		return reply(ctx, fakeAnswer(req.Method()), nil)
	})
	<-conn.Done()
}

// fakeFlavor starts the test binary as the language server and counts the starts.
type fakeFlavor struct{ starts atomic.Int32 }

func (f *fakeFlavor) Command(map[string]string) (string, []string, error) {
	f.starts.Add(1)
	return os.Args[0], []string{"-test.run=^$"}, nil
}
func (*fakeFlavor) RootOf(path string) string  { return filepath.Dir(path) }
func (*fakeFlavor) InitializationOptions() any { return nil }
func (*fakeFlavor) Configuration() any         { return nil }
func (*fakeFlavor) Environment() map[string]string {
	return map[string]string{fakeServerVariable: "1"}
}

// fakeClock is a timer factory that only fires when the test says so.
type fakeClock struct {
	mu      sync.Mutex
	pending []*fakeTimer
}

type fakeTimer struct {
	delay   time.Duration
	action  func()
	stopped atomic.Bool
}

func (t *fakeTimer) Stop() bool { t.stopped.Store(true); return true }

func (c *fakeClock) AfterFunc(d time.Duration, f func()) Timer {
	c.mu.Lock()
	defer c.mu.Unlock()
	timer := &fakeTimer{delay: d, action: f}
	c.pending = append(c.pending, timer)
	return timer
}

// fire runs the timers that were not stopped and returns their delays.
func (c *fakeClock) fire() []time.Duration {
	c.mu.Lock()
	timers := c.pending
	c.pending = nil
	c.mu.Unlock()
	var delays []time.Duration
	for _, timer := range timers {
		if !timer.stopped.Load() {
			delays = append(delays, timer.delay)
			timer.action()
		}
	}
	return delays
}

func waitFor(t *testing.T, what string, condition func() bool) {
	t.Helper()
	deadline := time.Now().Add(15 * time.Second)
	for time.Now().Before(deadline) {
		if condition() {
			return
		}
		time.Sleep(10 * time.Millisecond)
	}
	t.Fatalf("timed out waiting for %s", what)
}
