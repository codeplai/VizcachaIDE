package lsp

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	"go.lsp.dev/protocol"
)

func write(t *testing.T, path, text string) {
	t.Helper()
	if err := os.MkdirAll(filepath.Dir(path), 0o750); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path, []byte(text), 0o600); err != nil {
		t.Fatal(err)
	}
}

func TestScanSourcesOnlyTakesTheLanguageAndSkipsDependencies(t *testing.T) {
	root := t.TempDir()
	write(t, filepath.Join(root, "main.go"), "x")
	write(t, filepath.Join(root, "pkg", "util.GO"), "x")
	write(t, filepath.Join(root, "notes.txt"), "x")
	write(t, filepath.Join(root, "vendor", "dep.go"), "x")
	write(t, filepath.Join(root, ".git", "hook.go"), "x")
	found := scanSources(root, []string{".go"})
	if len(found) != 2 {
		t.Fatalf("found %v", found)
	}
}

func TestSourceDiffReportsCreatedChangedAndDeletedClosedFiles(t *testing.T) {
	root := t.TempDir()
	a, b, c, open := filepath.Join(root, "a.go"), filepath.Join(root, "b.go"), filepath.Join(root, "c.go"), filepath.Join(root, "open.go")
	t0 := time.Now()
	isOpen := func(path string) (document, bool) { return document{}, path == open }
	w := &sourceWatch{seen: map[string]time.Time{}, roots: map[string]bool{}}
	if got := w.diff(root, map[string]time.Time{a: t0, b: t0, open: t0}, isOpen); len(got) != 0 {
		t.Fatalf("the first scan only remembers: %v", got)
	}
	got := w.diff(root, map[string]time.Time{a: t0.Add(time.Second), c: t0, open: t0.Add(time.Second)}, isOpen)
	kinds := map[string]protocol.FileChangeType{}
	for _, event := range got {
		kinds[uriToPath(event.URI)] = event.Type
	}
	if len(kinds) != 3 || kinds[a] != protocol.FileChangeTypeChanged || kinds[b] != protocol.FileChangeTypeDeleted || kinds[c] != protocol.FileChangeTypeCreated {
		t.Fatalf("events = %v (open files are skipped)", kinds)
	}
	if again := w.diff(root, map[string]time.Time{a: t0.Add(time.Second), c: t0, open: t0.Add(time.Second)}, isOpen); len(again) != 0 {
		t.Fatalf("nothing changed: %v", again)
	}
}
