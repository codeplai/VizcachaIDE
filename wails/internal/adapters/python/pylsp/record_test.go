package pylsp

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"
	"testing"
	"time"

	"go.lsp.dev/jsonrpc2"
	"go.lsp.dev/uri"

	"github.com/codeplai/VizcachaIDE/wails/internal/adapters/python/pythontest"
)

// recordVariable turns TestRecordPylspSession on. It rewrites testdata/pylsp/session.json from
// the real python-lsp-server:  VIZCACHA_RECORD_PYLSP=1 go test -run Record ./internal/adapters/python/pylsp
const recordVariable = "VIZCACHA_RECORD_PYLSP"

const docURIPlaceholder = "{DOC_URI}"

type pipes struct {
	io.Reader
	io.WriteCloser
}

// Positions are 0-based LSP positions in testdata/pylsp/main.py.
var recordedQueries = []struct {
	name, method string
	line, column int
}{
	{"completion", "textDocument/completion", 7, 3},
	{"hover", "textDocument/hover", 6, 1},
	{"signatureHelp", "textDocument/signatureHelp", 4, 12},
	{"definition", "textDocument/definition", 4, 7},
}

func TestRecordPylspSession(t *testing.T) {
	if os.Getenv(recordVariable) != "1" {
		t.Skip("set " + recordVariable + "=1 to record the fixtures from the real pylsp")
	}
	python := pythontest.Interpreter(t)
	source, err := os.ReadFile(filepath.Join("testdata", "pylsp", "main.py"))
	if err != nil {
		t.Fatal(err)
	}
	folder := t.TempDir()
	file := filepath.Join(folder, "main.py")
	if err := os.WriteFile(file, source, 0o600); err != nil {
		t.Fatal(err)
	}
	docURI := string(uri.File(file))
	session := normalized(t, recordSession(t, python, folder, docURI, string(source)))
	var raw bytes.Buffer
	encoder := json.NewEncoder(&raw)
	encoder.SetEscapeHTML(false)
	encoder.SetIndent("", " ")
	if err := encoder.Encode(session); err != nil { // normalizes the server's "\/" escapes
		t.Fatal(err)
	}
	// pylsp spells the drive letter in lower case in some answers.
	text := regexp.MustCompile("(?i)"+regexp.QuoteMeta(docURI)).ReplaceAllString(raw.String(), docURIPlaceholder)
	if err := os.WriteFile(filepath.Join("testdata", "pylsp", "session.json"), []byte(text), 0o600); err != nil {
		t.Fatal(err)
	}
}

func recordSession(t *testing.T, python, folder, docURI, source string) map[string]json.RawMessage {
	t.Helper()
	cmd := exec.Command(python, "-m", "pylsp")
	stdin, _ := cmd.StdinPipe()
	stdout, _ := cmd.StdoutPipe()
	if err := cmd.Start(); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = cmd.Process.Kill() })
	diagnostics := make(chan json.RawMessage, 8)
	conn := jsonrpc2.NewConn(jsonrpc2.NewStream(pipes{Reader: stdout, WriteCloser: stdin}))
	conn.Go(context.Background(), func(ctx context.Context, reply jsonrpc2.Replier, req jsonrpc2.Request) error {
		if req.Method() == "textDocument/publishDiagnostics" {
			diagnostics <- req.Params()
		}
		if _, isCall := req.(*jsonrpc2.Call); isCall {
			return reply(ctx, nil, nil)
		}
		return nil
	})
	ctx, cancel := context.WithTimeout(context.Background(), time.Minute)
	defer cancel()
	session := map[string]json.RawMessage{}
	call := func(name, method string, params any) {
		var result json.RawMessage
		if _, err := conn.Call(ctx, method, params, &result); err != nil {
			t.Fatalf("%s: %v", method, err)
		}
		session[name] = result
	}
	call("initialize", "initialize", map[string]any{
		"processId": os.Getpid(), "rootUri": string(uri.File(folder)), "capabilities": recordedCapabilities(),
		"workspaceFolders": []any{map[string]any{"uri": string(uri.File(folder)), "name": filepath.Base(folder)}},
	})
	_ = conn.Notify(ctx, "initialized", map[string]any{})
	_ = conn.Notify(ctx, "workspace/didChangeConfiguration", map[string]any{"settings": Flavor{}.Configuration()})
	_ = conn.Notify(ctx, "textDocument/didOpen", map[string]any{"textDocument": map[string]any{
		"uri": docURI, "languageId": "python", "version": 1, "text": source,
	}})
	session["publishDiagnostics"] = firstWithFindings(t, diagnostics)
	document := map[string]any{"uri": docURI}
	for _, q := range recordedQueries {
		call(q.name, q.method, map[string]any{
			"textDocument": document, "position": map[string]any{"line": q.line, "character": q.column},
		})
	}
	call("documentSymbol", "textDocument/documentSymbol", map[string]any{"textDocument": document})
	return session
}

// firstWithFindings skips the empty publishDiagnostics pylsp sends while it still starts.
func firstWithFindings(t *testing.T, published <-chan json.RawMessage) json.RawMessage {
	t.Helper()
	for {
		select {
		case params := <-published:
			if !strings.Contains(string(params), `"diagnostics":[]`) && !strings.Contains(string(params), `"diagnostics": []`) {
				return params
			}
		case <-time.After(30 * time.Second):
			t.Fatal("pylsp reported no diagnostics")
		}
	}
}

// normalized decodes every answer so the encoder drops the "\/" escapes of pylsp's ujson.
func normalized(t *testing.T, session map[string]json.RawMessage) map[string]any {
	t.Helper()
	result := map[string]any{}
	for name, raw := range session {
		var value any
		if err := json.Unmarshal(raw, &value); err != nil {
			t.Fatal(err)
		}
		result[name] = value
	}
	return result
}

func recordedCapabilities() map[string]any {
	plain := []string{"plaintext"}
	return map[string]any{"textDocument": map[string]any{
		"completion": map[string]any{"completionItem": map[string]any{"snippetSupport": false, "documentationFormat": plain}},
		"hover":      map[string]any{"contentFormat": plain},
		"signatureHelp": map[string]any{"signatureInformation": map[string]any{
			"documentationFormat": plain, "parameterInformation": map[string]any{"labelOffsetSupport": true}, "activeParameterSupport": true,
		}},
		"documentSymbol": map[string]any{"hierarchicalDocumentSymbolSupport": true},
	}}
}
