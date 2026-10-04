package analyzer

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"go.lsp.dev/uri"
)

// rust-analyzer detaches a second loose file only when it is asked for its settings again
// (workspace/configuration) after the file is opened: the answer is Flavor.Configuration(). This
// test is the client that does it; protocol/lsp does not yet (CCR in the track report).
func TestRealRustAnalyzerDetachesASecondLooseFileWhenAskedForTheSettings(t *testing.T) {
	text := readTestdata(t, "loose.rs")
	first := filepath.Join(t.TempDir(), "main.rs")
	second := filepath.Join(t.TempDir(), "other.rs")
	for _, file := range []string{first, second} {
		if err := os.WriteFile(file, []byte(text), 0o600); err != nil {
			t.Fatal(err)
		}
	}
	flavor := NewFlavor(Config{Locator: locatorForTests(t)})
	flavor.RootOf(first)
	c := startRawWith(t, first, flavor.Configuration())
	c.open(first, text)
	time.Sleep(10 * time.Second)

	flavor.RootOf(second)
	c.pulled.Store(flavor.Configuration())
	c.notify("workspace/didChangeWorkspaceFolders", map[string]any{"event": map[string]any{
		"added": []any{map[string]any{"uri": string(uri.File(filepath.Dir(second))), "name": "second"}}, "removed": []any{},
	}})
	c.open(second, text)
	c.notify("workspace/didChangeConfiguration", map[string]any{"settings": nil})
	document := map[string]any{"uri": string(uri.File(second))}
	deadline := time.Now().Add(loadTimeout)
	for time.Now().Before(deadline) {
		answer, err := c.tryCall("textDocument/completion", map[string]any{
			"textDocument": document, "position": map[string]any{"line": 6, "character": 6},
		})
		if err == nil && strings.Contains(string(answer), `"label":"push"`) {
			return
		}
		time.Sleep(3 * time.Second)
	}
	t.Fatal("the second loose file got no completion")
}
