package analyzer

import (
	"bytes"
	"encoding/json"
	"os"
	"path/filepath"
	"regexp"
	"slices"
	"strings"
	"testing"
	"time"

	"go.lsp.dev/uri"
)

// recordVariable turns TestRecordRustAnalyzerSession on. It rewrites testdata/rust-analyzer/session.json
// from the real rust-analyzer:  VIZCACHA_RECORD_RA=1 go test -run Record ./internal/adapters/rust/analyzer
// (it takes about a minute: rust-analyzer indexes the standard library and runs cargo check).
const recordVariable = "VIZCACHA_RECORD_RA"

const docURIPlaceholder = "{DOC_URI}"

// Positions are 0-based LSP positions in testdata/rust-analyzer/crate/src/main.rs.
var recordedQueries = []struct {
	name, method string
	line, column int
}{
	{"memberCompletion", "textDocument/completion", 16, 6},
	{"hover", "textDocument/hover", 16, 7},
	{"signatureHelp", "textDocument/signatureHelp", 17, 18},
	{"definition", "textDocument/definition", 17, 13},
}

func TestRecordRustAnalyzerSession(t *testing.T) {
	if os.Getenv(recordVariable) != "1" {
		t.Skip("set " + recordVariable + "=1 to record the fixtures from the real rust-analyzer")
	}
	folder := t.TempDir()
	file := copyCrate(t, folder)
	docURI := string(uri.File(file))
	session := recordSession(t, file, docURI)
	var raw bytes.Buffer
	encoder := json.NewEncoder(&raw)
	encoder.SetEscapeHTML(false)
	encoder.SetIndent("", " ")
	if err := encoder.Encode(session); err != nil {
		t.Fatal(err)
	}
	// rust-analyzer spells the URI with a lower case drive letter; accept every spelling.
	text := raw.String()
	for _, spelling := range []string{docURI, strings.Replace(docURI, ":", "%3A", 1)} {
		text = regexp.MustCompile("(?i)"+regexp.QuoteMeta(spelling)).ReplaceAllString(text, docURIPlaceholder)
	}
	if err := os.WriteFile(filepath.Join("testdata", "rust-analyzer", "session.json"), []byte(text), 0o600); err != nil {
		t.Fatal(err)
	}
}

func recordSession(t *testing.T, file, docURI string) map[string]any {
	t.Helper()
	source, _ := os.ReadFile(file)
	c := startRaw(t, file)
	session := map[string]any{"initialize": decoded(t, c.initResult)}
	c.open(file, string(source))
	session["diagnosticsMoved"] = c.diagnosticsWith(docURI, "E0382")
	document := map[string]any{"uri": docURI}
	for _, q := range recordedQueries {
		session[q.name] = c.answer(q.method, map[string]any{
			"textDocument": document, "position": map[string]any{"line": q.line, "character": q.column},
		}, q.method == "textDocument/completion")
	}
	session["memberCompletion"] = onlyLabels(session["memberCompletion"], "push", "pop", "len")
	session["documentSymbol"] = c.answer("textDocument/documentSymbol", map[string]any{"textDocument": document}, false)
	unresolved, _ := os.ReadFile(filepath.Join("testdata", "rust-analyzer", "unresolved.rs"))
	_ = os.WriteFile(file, unresolved, 0o600) // cargo check reads the disk
	c.notify("textDocument/didChange", map[string]any{
		"textDocument":   map[string]any{"uri": docURI, "version": 2},
		"contentChanges": []any{map[string]any{"text": string(unresolved)}},
	})
	c.notify("textDocument/didSave", map[string]any{"textDocument": document})
	session["diagnosticsUnresolved"] = c.diagnosticsWith(docURI, "E0425")
	return session
}

// diagnosticsWith waits for the publishDiagnostics of the document that holds the error code.
// Only the diagnostics of the document stay: the macro expansion in the standard library that
// cargo also reports would put a path of the recording machine in the fixture.
func (c *rawClient) diagnosticsWith(docURI, code string) any {
	c.t.Helper()
	timeout := time.After(3 * time.Minute)
	for {
		select {
		case raw := <-c.diagnostics:
			var params struct {
				URI         string           `json:"uri"`
				Diagnostics []map[string]any `json:"diagnostics"`
			}
			if json.Unmarshal(raw, &params) != nil || !sameURI(params.URI, docURI) {
				continue
			}
			for _, d := range params.Diagnostics {
				if d["code"] == code {
					return withoutForeignRelated(decoded(c.t, raw), docURI)
				}
			}
		case <-timeout:
			c.t.Fatalf("rust-analyzer did not report %s", code)
		}
	}
}

// answer repeats a query until rust-analyzer, still loading, gives a real answer.
func (c *rawClient) answer(method string, params any, needsItems bool) any {
	c.t.Helper()
	deadline := time.Now().Add(3 * time.Minute)
	for time.Now().Before(deadline) {
		raw, err := c.tryCall(method, params)
		text := strings.TrimSpace(string(raw))
		empty := text == "" || text == "null" || text == "[]" || (needsItems && strings.Contains(text, `"items":[]`))
		if err == nil && !empty {
			return decoded(c.t, raw)
		}
		time.Sleep(2 * time.Second)
	}
	c.t.Fatalf("%s: no answer", method)
	return nil
}

// onlyLabels keeps a few completion items: the whole list of Vec has 100 of them and 280 KB.
func onlyLabels(completion any, labels ...string) any {
	list, _ := completion.(map[string]any)
	items, _ := list["items"].([]any)
	kept := []any{}
	for _, item := range items {
		if label, _ := item.(map[string]any)["label"].(string); slices.Contains(labels, label) {
			kept = append(kept, item)
		}
	}
	list["items"] = kept
	return list
}

func decoded(t *testing.T, raw json.RawMessage) any {
	t.Helper()
	var value any
	if err := json.Unmarshal(raw, &value); err != nil {
		t.Fatal(err)
	}
	return value
}

// withoutForeignRelated drops the related information that points outside the document.
func withoutForeignRelated(params any, docURI string) any {
	published, _ := params.(map[string]any)
	list, _ := published["diagnostics"].([]any)
	for _, item := range list {
		diagnostic, _ := item.(map[string]any)
		related, _ := diagnostic["relatedInformation"].([]any)
		kept := []any{}
		for _, entry := range related {
			location, _ := entry.(map[string]any)["location"].(map[string]any)
			if link, _ := location["uri"].(string); sameURI(link, docURI) {
				kept = append(kept, entry)
			}
		}
		diagnostic["relatedInformation"] = kept
	}
	return params
}

func sameURI(a, b string) bool {
	return strings.EqualFold(strings.ReplaceAll(a, "%3A", ":"), strings.ReplaceAll(b, "%3A", ":"))
}

// copyCrate copies testdata/rust-analyzer/crate into folder and returns its src/main.rs.
func copyCrate(t *testing.T, folder string) string {
	t.Helper()
	if err := os.MkdirAll(filepath.Join(folder, "src"), 0o750); err != nil {
		t.Fatal(err)
	}
	for _, name := range []string{"Cargo.toml", filepath.Join("src", "main.rs")} {
		content, err := os.ReadFile(filepath.Join("testdata", "rust-analyzer", "crate", name))
		if err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(filepath.Join(folder, name), content, 0o600); err != nil {
			t.Fatal(err)
		}
	}
	return filepath.Join(folder, "src", "main.rs")
}
