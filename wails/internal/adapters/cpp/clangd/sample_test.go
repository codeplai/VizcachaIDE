package clangd

import (
	"context"
	"os"
	"path/filepath"
	"sync"
	"testing"
	"time"

	"github.com/codeplai/VizcachaIDE/wails/internal/adapters/cpp"
	"github.com/codeplai/VizcachaIDE/wails/internal/adapters/cpp/cpptest"
	"github.com/codeplai/VizcachaIDE/wails/internal/app"
	"github.com/codeplai/VizcachaIDE/wails/internal/domain"
	"github.com/codeplai/VizcachaIDE/wails/internal/protocol/lsp"
)

const loadTimeout = 90 * time.Second

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

// sample is testdata/clangd/main.cpp. Domain positions (1-based) of the recorded queries.
type sample struct {
	server *lsp.Server
	sink   *recordingSink
	file   string
}

var (
	atStdVec = func(file string) domain.SourceLocation {
		return domain.SourceLocation{File: file, Line: 18, Column: 13}
	}
	atMemberPush = func(file string) domain.SourceLocation {
		return domain.SourceLocation{File: file, Line: 15, Column: 11}
	}
	atPushBack  = func(file string) domain.SourceLocation { return domain.SourceLocation{File: file, Line: 15, Column: 9} }
	insideTwice = func(file string) domain.SourceLocation {
		return domain.SourceLocation{File: file, Line: 16, Column: 19}
	}
	atTwiceCall = func(file string) domain.SourceLocation {
		return domain.SourceLocation{File: file, Line: 16, Column: 14}
	}
	sampleSource = filepath.Join("testdata", "clangd", "main.cpp")
)

// openSample writes main.cpp in a temp folder and opens it in a server of the given flavor,
// optionally waiting for its first diagnostics.
func openSample(t *testing.T, flavor lsp.Flavor, withDiagnostics bool) sample {
	t.Helper()
	text, err := os.ReadFile(sampleSource)
	if err != nil {
		t.Fatal(err)
	}
	file := filepath.Join(t.TempDir(), "main.cpp")
	if err := os.WriteFile(file, text, 0o600); err != nil {
		t.Fatal(err)
	}
	sink := &recordingSink{diagnostics: map[string][]domain.Diagnostic{}}
	server := lsp.New(sink, flavor, lsp.Options{Name: serverName, LanguageID: languageID, CodeLanguage: domain.CodeLanguageCpp})
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

// locatorFor finds the compiler on compilerBin only and clangd/clang-format in llvmBin.
func locatorFor(compilerBin, llvmBin string) *cpp.Locator {
	return cpp.NewLocator(cpp.Options{
		Settings:        settingsStub{paths: map[string]string{cpp.ToolClangd: filepath.Join(llvmBin, cpptest.Exe("clangd"))}},
		AppDir:          os.TempDir(),
		BaseEnvironment: []string{"PATH=" + compilerBin},
	})
}

type settingsStub struct{ paths map[string]string }

func (s settingsStub) Load() (domain.Settings, error) {
	settings := domain.DefaultSettings()
	settings.ToolPaths = s.paths
	return settings, nil
}
func (settingsStub) Save(domain.Settings) error { return nil }
