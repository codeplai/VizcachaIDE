package lsp

import (
	"context"
	"path/filepath"
	"sync"
	"testing"
	"time"

	"github.com/codeplai/VizcachaIDE/wails/internal/app"
	"github.com/codeplai/VizcachaIDE/wails/internal/domain"
)

// recordingSink keeps the lsp events. The other EventSink methods come from the nil embed
// and must never be called by this adapter.
type recordingSink struct {
	app.EventSink
	mu          sync.Mutex
	statuses    []domain.ServerStatus
	diagnostics map[string][]domain.Diagnostic
}

func newRecordingSink() *recordingSink {
	return &recordingSink{diagnostics: map[string][]domain.Diagnostic{}}
}

func (r *recordingSink) LanguageServerStatus(status domain.ServerStatus) {
	r.mu.Lock()
	r.statuses = append(r.statuses, status)
	r.mu.Unlock()
}

func (r *recordingSink) Diagnostics(path string, diagnostics []domain.Diagnostic) {
	r.mu.Lock()
	r.diagnostics[path] = diagnostics
	r.mu.Unlock()
}

func (r *recordingSink) lastStatus() domain.ServerStatus {
	r.mu.Lock()
	defer r.mu.Unlock()
	if len(r.statuses) == 0 {
		return ""
	}
	return r.statuses[len(r.statuses)-1]
}

// missingToolFlavor is a language server whose tool is not installed.
type missingToolFlavor struct{}

func (missingToolFlavor) Command(map[string]string) (string, []string, error) {
	return "", nil, app.MissingTool("fake-ls")
}
func (missingToolFlavor) RootOf(path string) string      { return filepath.Dir(path) }
func (missingToolFlavor) InitializationOptions() any     { return nil }
func (missingToolFlavor) Configuration() any             { return nil }
func (missingToolFlavor) Environment() map[string]string { return nil }

func TestMissingToolIsUnavailableAndQueriesAreEmpty(t *testing.T) {
	sink := newRecordingSink()
	server := New(sink, missingToolFlavor{}, Options{})
	ctx := context.Background()
	file := filepath.Join(t.TempDir(), "main.go")

	if err := server.OpenDocument(ctx, file, "package main\n"); err != nil {
		t.Fatal(err)
	}
	if sink.lastStatus() != domain.ServerUnavailable {
		t.Fatalf("status = %q, want unavailable", sink.lastStatus())
	}
	at := domain.SourceLocation{File: file, Line: 1, Column: 1}
	started := time.Now()
	items, err := server.Completion(ctx, at)
	hover, _ := server.Hover(ctx, at)
	definition, _ := server.Definition(ctx, at)
	symbols, _ := server.DocumentSymbols(ctx, file)
	if err != nil || len(items) != 0 || hover != "" || definition != nil || len(symbols) != 0 {
		t.Errorf("queries must be empty: %v %v %q %v %v", err, items, hover, definition, symbols)
	}
	if time.Since(started) > 200*time.Millisecond {
		t.Error("queries to an unavailable server must not wait")
	}
	for _, step := range []error{
		server.ChangeDocument(ctx, file, "package x\n", 2),
		server.CloseDocument(ctx, file),
		server.Shutdown(ctx),
	} {
		if step != nil {
			t.Error(step)
		}
	}
}

func TestQueriesOnUnopenedDocumentsAreEmpty(t *testing.T) {
	server := New(newRecordingSink(), missingToolFlavor{}, Options{})
	at := domain.SourceLocation{File: "nunca-abierto.go", Line: 1, Column: 1}
	if items, err := server.Completion(context.Background(), at); err != nil || len(items) != 0 {
		t.Errorf("Completion = %v, %v", items, err)
	}
}
