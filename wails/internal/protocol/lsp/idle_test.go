package lsp

import (
	"context"
	"path/filepath"
	"testing"
	"time"

	"github.com/codeplai/VizcachaIDE/wails/internal/domain"
)

func (s *Server) currentState() serverState {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.state
}

func (r *recordingSink) allStatuses() []domain.ServerStatus {
	r.mu.Lock()
	defer r.mu.Unlock()
	return append([]domain.ServerStatus(nil), r.statuses...)
}

// startFakeServer returns the server and a file path; the folder is created before the
// server cleanup is registered so the process (whose cwd is that folder) ends before it.
func startFakeServer(t *testing.T) (*Server, *fakeFlavor, *fakeClock, *recordingSink, string) {
	t.Helper()
	file := filepath.Join(t.TempDir(), "main.go")
	flavor, clock, sink := &fakeFlavor{}, &fakeClock{}, newRecordingSink()
	server := New(sink, flavor, Options{IdleTimeout: 5 * time.Minute, AfterFunc: clock.AfterFunc})
	t.Cleanup(func() { _ = server.Shutdown(context.Background()) })
	return server, flavor, clock, sink, file
}

func TestServerShutsDownWhenNoDocumentStaysOpenAndRestartsOnDemand(t *testing.T) {
	server, flavor, clock, sink, file := startFakeServer(t)
	ctx := context.Background()

	_ = server.OpenDocument(ctx, file, "package main\n")
	waitFor(t, "ready", func() bool { return sink.lastStatus() == domain.ServerReady })
	_ = server.CloseDocument(ctx, file)
	if got := clock.fire(); len(got) != 1 || got[0] != 5*time.Minute {
		t.Fatalf("idle timers = %v, want one of 5m", got)
	}
	if server.currentState() != stateIdle {
		t.Fatalf("state = %v, want idle", server.currentState())
	}
	for _, status := range sink.allStatuses() {
		if status == domain.ServerUnavailable {
			t.Fatalf("idleness must not report unavailable: %v", sink.allStatuses())
		}
	}

	_ = server.OpenDocument(ctx, file, "package main\n")
	waitFor(t, "ready again", func() bool { return server.currentState() == stateReady })
	if flavor.starts.Load() != 2 {
		t.Errorf("starts = %d, want 2", flavor.starts.Load())
	}
}

func TestReopeningADocumentCancelsTheIdleShutdown(t *testing.T) {
	server, flavor, clock, sink, file := startFakeServer(t)
	ctx := context.Background()

	_ = server.OpenDocument(ctx, file, "package main\n")
	waitFor(t, "ready", func() bool { return sink.lastStatus() == domain.ServerReady })
	_ = server.CloseDocument(ctx, file)
	_ = server.OpenDocument(ctx, file, "package main\n")
	if got := clock.fire(); len(got) != 0 {
		t.Errorf("a stopped timer fired: %v", got)
	}
	if server.currentState() != stateReady || flavor.starts.Load() != 1 {
		t.Errorf("state = %v, starts = %d", server.currentState(), flavor.starts.Load())
	}
}

func TestMissingToolStaysUnavailableAfterTheIdleTimeout(t *testing.T) {
	clock, sink := &fakeClock{}, newRecordingSink()
	server := New(sink, missingToolFlavor{}, Options{IdleTimeout: time.Minute, AfterFunc: clock.AfterFunc})
	file := filepath.Join(t.TempDir(), "main.go")
	_ = server.OpenDocument(context.Background(), file, "package main\n")
	_ = server.CloseDocument(context.Background(), file)
	clock.fire()
	if server.currentState() != stateUnavailable || sink.lastStatus() != domain.ServerUnavailable {
		t.Errorf("state = %v, status = %q", server.currentState(), sink.lastStatus())
	}
}
