package analyzer

import (
	"context"
	"os"
	"path/filepath"
	"sync"
	"testing"
	"time"

	"github.com/codeplai/VizcachaIDE/wails/internal/app"
	"github.com/codeplai/VizcachaIDE/wails/internal/domain"
	"github.com/codeplai/VizcachaIDE/wails/internal/protocol/lsp"
)

// loadTimeout is generous: rust-analyzer indexes the standard library and runs cargo check.
const loadTimeout = 3 * time.Minute

// recordingSink keeps the lsp events. The other EventSink methods come from the nil embed
// and must never be called by this adapter.
type recordingSink struct {
	app.EventSink
	mu          sync.Mutex
	statuses    []domain.ServerStatus
	diagnostics map[string][]domain.Diagnostic
}

func (r *recordingSink) LanguageServerStatus(_ domain.CodeLanguage, status domain.ServerStatus) {
	r.mu.Lock()
	r.statuses = append(r.statuses, status)
	r.mu.Unlock()
}

func (r *recordingSink) Diagnostics(path string, diagnostics []domain.Diagnostic) {
	r.mu.Lock()
	r.diagnostics[path] = diagnostics
	r.mu.Unlock()
}

func (r *recordingSink) diagnosticsOf(path string) []domain.Diagnostic {
	r.mu.Lock()
	defer r.mu.Unlock()
	return r.diagnostics[path]
}

// withCode is the error with the code, if the file has it now. rust-analyzer also sends the
// notes of an error ("value moved here") as hints with the same code.
func (r *recordingSink) withCode(path, code string) (domain.Diagnostic, bool) {
	for _, d := range r.diagnosticsOf(path) {
		if d.Code == code && d.Severity == domain.SeverityError {
			return d, true
		}
	}
	return domain.Diagnostic{}, false
}

// sample is a server with main.rs open. Domain positions are 1-based.
type sample struct {
	server *lsp.Server
	sink   *recordingSink
	file   string
	text   string
}

// Positions (1-based) of the recorded queries in testdata/rust-analyzer/crate/src/main.rs.
var (
	afterVecDot = func(file string) domain.SourceLocation { return domain.SourceLocation{File: file, Line: 17, Column: 7} }
	atPush      = func(file string) domain.SourceLocation { return domain.SourceLocation{File: file, Line: 17, Column: 8} }
	insideTwice = func(file string) domain.SourceLocation {
		return domain.SourceLocation{File: file, Line: 18, Column: 19}
	}
	atTwiceCall = func(file string) domain.SourceLocation {
		return domain.SourceLocation{File: file, Line: 18, Column: 14}
	}
)

// openIn writes text as file and opens it in a server of the flavor.
func openIn(t *testing.T, flavor lsp.Flavor, file, text string) sample {
	t.Helper()
	if err := os.MkdirAll(filepath.Dir(file), 0o750); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(file, []byte(text), 0o600); err != nil {
		t.Fatal(err)
	}
	sink := &recordingSink{diagnostics: map[string][]domain.Diagnostic{}}
	server := lsp.New(sink, flavor, lsp.Options{Name: serverName, LanguageID: languageID, CodeLanguage: domain.CodeLanguageRust})
	t.Cleanup(func() { _ = server.Shutdown(context.Background()) })
	if err := server.OpenDocument(context.Background(), file, text); err != nil {
		t.Fatal(err)
	}
	return sample{server: server, sink: sink, file: file, text: text}
}

// openReplay opens the crate sample in the server that replays the recording.
func openReplay(t *testing.T) sample {
	t.Helper()
	text, err := os.ReadFile(filepath.Join("testdata", "rust-analyzer", "crate", "src", "main.rs"))
	if err != nil {
		t.Fatal(err)
	}
	flavor := replayFlavor{NewFlavor(Config{})}
	return openIn(t, flavor, filepath.Join(t.TempDir(), "main.rs"), string(text))
}

func readTestdata(t *testing.T, name string) string {
	t.Helper()
	text, err := os.ReadFile(filepath.Join("testdata", "rust-analyzer", name))
	if err != nil {
		t.Fatal(err)
	}
	return string(text)
}

func waitUntil(t *testing.T, condition func() bool) {
	t.Helper()
	deadline := time.Now().Add(loadTimeout)
	for time.Now().Before(deadline) {
		if condition() {
			return
		}
		time.Sleep(100 * time.Millisecond)
	}
	t.Fatal("timed out waiting for the language server")
}

// eventually retries a query until the server, still loading, answers.
func eventually[T any](t *testing.T, query func() (T, bool)) T {
	t.Helper()
	var last T
	waitUntil(t, func() bool {
		var ok bool
		last, ok = query()
		return ok
	})
	return last
}
