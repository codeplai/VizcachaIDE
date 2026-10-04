package analyzer

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/codeplai/VizcachaIDE/wails/internal/domain"
	"github.com/codeplai/VizcachaIDE/wails/internal/protocol/lsp"
)

// These tests start the real rust-analyzer (they skip without the toolchain of
// docs/PLAN_RUST.md section 0). The first answer takes seconds: it indexes the standard library
// and runs cargo check.

// openCrate writes the sample crate in a temp folder and opens its main.rs.
func openCrate(t *testing.T) sample {
	t.Helper()
	folder := t.TempDir()
	writeManifest(t, folder, readTestdata(t, filepath.Join("crate", "Cargo.toml")))
	text := readTestdata(t, filepath.Join("crate", "src", "main.rs"))
	return openIn(t, NewFlavor(Config{Locator: locatorForTests(t)}), filepath.Join(folder, "src", "main.rs"), text)
}

func TestRealRustAnalyzerReportsTheMovedValueOfACrate(t *testing.T) {
	s := openCrate(t)
	waitUntil(t, func() bool { _, found := s.sink.withCode(s.file, "E0382"); return found })
	moved, _ := s.sink.withCode(s.file, "E0382")
	if moved.Location.Line != 21 || moved.Location.Column != 26 || !strings.Contains(moved.Message, "moved value") {
		t.Errorf("moved = %+v at %+v", moved, *moved.Location)
	}
}

func TestRealRustAnalyzerCompletesAndExplainsPushInACrate(t *testing.T) {
	s := openCrate(t)
	members := eventually(t, func() ([]domain.CompletionItem, bool) {
		items, _ := s.server.Completion(context.Background(), afterVecDot(s.file))
		return items, len(items) > 0
	})
	if !hasLabel(members, "push") {
		t.Errorf("no push in %d members", len(members))
	}
	hover := eventually(t, func() (string, bool) {
		text, _ := s.server.Hover(context.Background(), atPush(s.file))
		return text, text != ""
	})
	if !strings.Contains(hover, "push") {
		t.Errorf("hover = %q", hover)
	}
}

// A loose .rs has no Cargo.toml: rust-analyzer takes it as a detached file and still gives its
// own diagnostics (types, syntax), completion and hover, but not the ones of cargo check.
func TestRealRustAnalyzerServesALooseFile(t *testing.T) {
	file := filepath.Join(t.TempDir(), "main.rs")
	s := openIn(t, NewFlavor(Config{Locator: locatorForTests(t)}), file, readTestdata(t, "loose.rs"))
	waitUntil(t, func() bool { _, found := s.sink.withCode(file, "E0308"); return found })
	mismatch, _ := s.sink.withCode(file, "E0308")
	if mismatch.Location.Line != 8 || mismatch.Location.Column != 18 || !strings.Contains(mismatch.Message, "expected i32") {
		t.Errorf("mismatch = %+v at %+v", mismatch, *mismatch.Location)
	}
	// "    v.push(1);" is line 7.
	members := eventually(t, func() ([]domain.CompletionItem, bool) {
		items, _ := s.server.Completion(context.Background(), domain.SourceLocation{File: file, Line: 7, Column: 7})
		return items, len(items) > 0
	})
	if !hasLabel(members, "push") {
		t.Errorf("no push in %d members", len(members))
	}
	hover := eventually(t, func() (string, bool) {
		text, _ := s.server.Hover(context.Background(), domain.SourceLocation{File: file, Line: 7, Column: 8})
		return text, text != ""
	})
	if !strings.Contains(hover, "push") {
		t.Errorf("hover = %q", hover)
	}
}

// A reloaded window opens the same file again: the diagnostics must be there again.
func TestRealRustAnalyzerRepublishesWhenTheFileIsOpenedAgain(t *testing.T) {
	file := filepath.Join(t.TempDir(), "main.rs")
	s := openIn(t, NewFlavor(Config{Locator: locatorForTests(t)}), file, readTestdata(t, "loose.rs"))
	waitUntil(t, func() bool { _, found := s.sink.withCode(file, "E0308"); return found })
	s.sink.mu.Lock()
	delete(s.sink.diagnostics, file)
	s.sink.mu.Unlock()
	text, err := os.ReadFile(file)
	if err != nil {
		t.Fatal(err)
	}
	if err := s.server.OpenDocument(context.Background(), file, string(text)); err != nil {
		t.Fatal(err)
	}
	waitUntil(t, func() bool { _, found := s.sink.withCode(file, "E0308"); return found })
}

func hasLabel(items []domain.CompletionItem, label string) bool {
	for _, item := range items {
		if item.Label == label {
			return true
		}
	}
	return false
}

// A second loose file in another folder is a new detached file: the client tells rust-analyzer
// its settings changed (lsp.PulledConfiguration) and it serves that file too.
func TestRealRustAnalyzerServesASecondLooseFile(t *testing.T) {
	first := filepath.Join(t.TempDir(), "uno", "main.rs")
	s := openIn(t, NewFlavor(Config{Locator: locatorForTests(t)}), first, readTestdata(t, "loose.rs"))
	waitUntil(t, func() bool { _, found := s.sink.withCode(first, "E0308"); return found })
	second := filepath.Join(t.TempDir(), "dos", "main.rs")
	if err := os.MkdirAll(filepath.Dir(second), 0o750); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(second, []byte(s.text), 0o600); err != nil {
		t.Fatal(err)
	}
	if err := s.server.OpenDocument(context.Background(), second, s.text); err != nil {
		t.Fatal(err)
	}
	// "    v.push(1);" is line 7.
	members := eventually(t, func() ([]domain.CompletionItem, bool) {
		items, _ := s.server.Completion(context.Background(), domain.SourceLocation{File: second, Line: 7, Column: 7})
		return items, len(items) > 0
	})
	if !hasLabel(members, "push") {
		t.Errorf("no push in %d members", len(members))
	}
}

// rust-analyzer started on a loose file does not load a crate opened later (found in the M3 QA):
// the Router gives the crate its own rust-analyzer.
func TestRouterServesACrateOpenedAfterALooseFile(t *testing.T) {
	loose := filepath.Join(t.TempDir(), "main.rs")
	folder := t.TempDir()
	// Registered after the folders: cleanups run last first, rust-analyzer must stop before they go.
	router := NewRouter(&recordingSink{diagnostics: map[string][]domain.Diagnostic{}}, Config{Locator: locatorForTests(t)}, lsp.Options{})
	t.Cleanup(func() { _ = router.Shutdown(context.Background()) })
	looseText := readTestdata(t, "loose.rs")
	if err := os.WriteFile(loose, []byte(looseText), 0o600); err != nil {
		t.Fatal(err)
	}
	if err := router.OpenDocument(context.Background(), loose, looseText); err != nil {
		t.Fatal(err)
	}
	writeManifest(t, folder, readTestdata(t, filepath.Join("crate", "Cargo.toml")))
	crate := filepath.Join(folder, "src", "main.rs")
	crateText := readTestdata(t, filepath.Join("crate", "src", "main.rs"))
	if err := os.MkdirAll(filepath.Dir(crate), 0o750); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(crate, []byte(crateText), 0o600); err != nil {
		t.Fatal(err)
	}
	if err := router.OpenDocument(context.Background(), crate, crateText); err != nil {
		t.Fatal(err)
	}
	members := eventually(t, func() ([]domain.CompletionItem, bool) {
		items, _ := router.Completion(context.Background(), afterVecDot(crate))
		return items, len(items) > 0
	})
	if !hasLabel(members, "push") {
		t.Errorf("no push in %d members", len(members))
	}
}
