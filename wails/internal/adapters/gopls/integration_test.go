package gopls

import (
	"context"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/codeplai/VizcachaIDE/wails/internal/domain"
)

const integrationSource = `package main

import "fmt"

func greet(name string) string {
	return "hola " + name
}

func main() {
	unused := 1
	fmt.Println(greet("vizcacha"))
	fmt.Pr
}
`

const (
	loadTimeout  = 4 * time.Minute
	queryTimeout = time.Minute
)

type goplsSession struct {
	server *Server
	sink   *recordingSink
	file   string
}

// startSession opens integrationSource in a temp module against the real gopls.
// It skips the test when gopls or go are not installed.
func startSession(t *testing.T) goplsSession {
	t.Helper()
	if _, err := locateGopls("", nil); err != nil {
		t.Skip("gopls is not installed")
	}
	if _, err := exec.LookPath("go"); err != nil {
		t.Skip("go is not installed")
	}
	dir := t.TempDir()
	if err := os.WriteFile(filepath.Join(dir, "go.mod"), []byte("module example.com/demo\n\ngo 1.21\n"), 0o600); err != nil {
		t.Fatal(err)
	}
	file := filepath.Join(dir, "main.go")
	if err := os.WriteFile(file, []byte(integrationSource), 0o600); err != nil {
		t.Fatal(err)
	}
	sink := newRecordingSink()
	server := New(sink, Config{})
	t.Cleanup(func() { _ = server.Shutdown(context.Background()) })
	if err := server.OpenDocument(context.Background(), file, integrationSource); err != nil {
		t.Fatal(err)
	}
	waitUntil(t, loadTimeout, func() bool { return len(sink.diagnosticsOf(file)) > 0 })
	return goplsSession{server: server, sink: sink, file: file}
}

func waitUntil(t *testing.T, limit time.Duration, condition func() bool) {
	t.Helper()
	deadline := time.Now().Add(limit)
	for time.Now().Before(deadline) {
		if condition() {
			return
		}
		time.Sleep(50 * time.Millisecond)
	}
	t.Fatal("timed out waiting for gopls")
}

// eventually retries a (1.5 s bounded) query until gopls, still loading, answers.
func eventually[T any](t *testing.T, query func() (T, bool)) T {
	t.Helper()
	var last T
	waitUntil(t, queryTimeout, func() bool {
		var ok bool
		last, ok = query()
		return ok
	})
	return last
}

func TestRealGoplsReportsTheUnusedVariableWithItsExactRange(t *testing.T) {
	session := startSession(t)
	var unused *domain.Diagnostic
	for _, d := range session.sink.diagnosticsOf(session.file) {
		if strings.Contains(d.Message, "unused") {
			unused = &d
		}
	}
	if unused == nil {
		t.Fatalf("no unused-variable diagnostic: %+v", session.sink.diagnosticsOf(session.file))
	}
	if unused.Location.Line != 10 || unused.Location.Column != 2 || unused.End.Line != 10 || unused.End.Column != 8 {
		t.Errorf("range = %+v - %+v, want 10:2-10:8", *unused.Location, *unused.End)
	}
	if unused.Severity != domain.SeverityError || unused.Source != "gopls" {
		t.Errorf("diagnostic = %+v", *unused)
	}
	if session.sink.lastStatus() != domain.ServerReady {
		t.Errorf("status = %q, want ready", session.sink.lastStatus())
	}
}

func TestRealGoplsCompletesFmtPrWithPrintln(t *testing.T) {
	session := startSession(t)
	at := domain.SourceLocation{File: session.file, Line: 12, Column: len("\tfmt.Pr") + 1}
	items := eventually(t, func() ([]domain.CompletionItem, bool) {
		items, _ := session.server.Completion(context.Background(), at)
		return items, len(items) > 0
	})
	for _, item := range items {
		if item.Label == "Println" {
			return
		}
	}
	t.Errorf("Println not suggested: %+v", items)
}

func TestRealGoplsFindsTheLocalDefinitionAndSymbols(t *testing.T) {
	session := startSession(t)
	call := domain.SourceLocation{File: session.file, Line: 11, Column: strings.Index(strings.Split(integrationSource, "\n")[10], "greet") + 1}
	target := eventually(t, func() (*domain.SourceLocation, bool) {
		target, _ := session.server.Definition(context.Background(), call)
		return target, target != nil
	})
	if target.Line != 5 || target.Column != 6 || pathKey(target.File) != pathKey(session.file) {
		t.Errorf("definition = %+v, want 5:6 in main.go", target)
	}
	symbols := eventually(t, func() ([]domain.DocumentSymbol, bool) {
		symbols, _ := session.server.DocumentSymbols(context.Background(), session.file)
		return symbols, len(symbols) > 0
	})
	if len(symbols) != 2 || symbols[0].Name != "greet" || symbols[1].Name != "main" ||
		symbols[0].Kind != domain.SymbolFunction || symbols[0].Range.End.Line != 7 {
		t.Errorf("symbols = %+v", symbols)
	}
	ranges := eventually(t, func() ([]domain.SourceRange, bool) {
		ranges, _ := session.server.DocumentHighlights(context.Background(), call)
		return ranges, len(ranges) > 0
	})
	if len(ranges) != 2 {
		t.Errorf("highlights = %+v", ranges)
	}
}
