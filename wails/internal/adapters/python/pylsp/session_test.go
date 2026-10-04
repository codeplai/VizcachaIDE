package pylsp

import (
	"context"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/codeplai/VizcachaIDE/wails/internal/domain"
	"github.com/codeplai/VizcachaIDE/wails/internal/protocol/lsp"
)

const loadTimeout = time.Minute

// sample is testdata/pylsp/main.py: print(x) on line 6, len("abc") on line 7 and "pri" on line 8.
type sample struct {
	server *lsp.Server
	sink   *recordingSink
	file   string
}

// Domain positions (1-based) of the queries recorded in session.json.
var (
	atPri       = func(file string) domain.SourceLocation { return domain.SourceLocation{File: file, Line: 8, Column: 4} }
	atLen       = func(file string) domain.SourceLocation { return domain.SourceLocation{File: file, Line: 7, Column: 2} }
	atGreetCall = func(file string) domain.SourceLocation { return domain.SourceLocation{File: file, Line: 5, Column: 8} }
	insideGreet = func(file string) domain.SourceLocation { return domain.SourceLocation{File: file, Line: 5, Column: 13} }
)

// openSample writes main.py in a temp folder and opens it in a server of the given flavor,
// optionally waiting for its first diagnostics.
func openSample(t *testing.T, flavor lsp.Flavor, withDiagnostics bool) sample {
	t.Helper()
	text, err := os.ReadFile(filepath.Join("testdata", "pylsp", "main.py"))
	if err != nil {
		t.Fatal(err)
	}
	file := filepath.Join(t.TempDir(), "main.py")
	if err := os.WriteFile(file, text, 0o600); err != nil {
		t.Fatal(err)
	}
	sink := newRecordingSink()
	server := lsp.New(sink, flavor, lsp.Options{Name: serverName, LanguageID: languageID})
	t.Cleanup(func() { _ = server.Shutdown(context.Background()) })
	if err := server.OpenDocument(context.Background(), file, string(text)); err != nil {
		t.Fatal(err)
	}
	if withDiagnostics {
		waitUntil(t, func() bool { return len(sink.diagnosticsOf(file)) > 0 })
	}
	return sample{server: server, sink: sink, file: file}
}

func waitUntil(t *testing.T, condition func() bool) {
	t.Helper()
	deadline := time.Now().Add(loadTimeout)
	for time.Now().Before(deadline) {
		if condition() {
			return
		}
		time.Sleep(50 * time.Millisecond)
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
