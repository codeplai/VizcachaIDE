package analyzer

import (
	"context"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"go.lsp.dev/jsonrpc2"
)

const (
	fakeServerVariable  = "VIZCACHA_FAKE_RUST_ANALYZER"
	fakeSessionVariable = "VIZCACHA_FAKE_RUST_ANALYZER_SESSION"
)

// TestMain turns the test binary into a language server that replays the answers recorded
// from the real rust-analyzer (testdata/rust-analyzer/session.json) when the variable is set.
func TestMain(m *testing.M) {
	if os.Getenv(fakeServerVariable) == "1" {
		runReplayServer()
		return
	}
	os.Exit(m.Run())
}

type stdio struct {
	*os.File
	out *os.File
}

func (s stdio) Write(p []byte) (int, error) { return s.out.Write(p) }

// answers maps a request method to the name of the recorded answer.
var answers = map[string]string{
	"initialize":                  "initialize",
	"textDocument/completion":     "memberCompletion",
	"textDocument/hover":          "hover",
	"textDocument/signatureHelp":  "signatureHelp",
	"textDocument/definition":     "definition",
	"textDocument/documentSymbol": "documentSymbol",
}

// notifications maps what the editor sends to the recorded diagnostics: opening the file
// gives the moved value of cargo check, changing it the unresolved name.
var notifications = map[string]string{
	"textDocument/didOpen":   "diagnosticsMoved",
	"textDocument/didChange": "diagnosticsUnresolved",
}

func runReplayServer() {
	raw, err := os.ReadFile(os.Getenv(fakeSessionVariable))
	if err != nil {
		os.Exit(2)
	}
	var session map[string]json.RawMessage
	if json.Unmarshal(raw, &session) != nil {
		os.Exit(2)
	}
	docURI := ""
	conn := jsonrpc2.NewConn(jsonrpc2.NewStream(stdio{File: os.Stdin, out: os.Stdout}))
	conn.Go(context.Background(), func(ctx context.Context, reply jsonrpc2.Replier, req jsonrpc2.Request) error {
		if req.Method() == "exit" {
			os.Exit(0)
		}
		if uri := documentURI(req.Params()); uri != "" {
			docURI = uri
		}
		if name, known := notifications[req.Method()]; known {
			return conn.Notify(ctx, "textDocument/publishDiagnostics", withURI(session[name], docURI))
		}
		if name, known := answers[req.Method()]; known {
			return reply(ctx, withURI(session[name], docURI), nil)
		}
		return reply(ctx, nil, nil)
	})
	<-conn.Done()
}

func documentURI(params json.RawMessage) string {
	var document struct {
		TextDocument struct {
			URI string `json:"uri"`
		} `json:"textDocument"`
	}
	_ = json.Unmarshal(params, &document)
	return document.TextDocument.URI
}

// withURI puts the URI of the opened document where the recording has {DOC_URI}.
func withURI(recorded json.RawMessage, docURI string) json.RawMessage {
	escaped, _ := json.Marshal(docURI)
	return json.RawMessage(strings.ReplaceAll(string(recorded), `"`+docURIPlaceholder+`"`, string(escaped)))
}

// replayFlavor is the rust-analyzer flavor, started as the replay server.
type replayFlavor struct{ *Flavor }

func (replayFlavor) Command(map[string]string) (string, []string, error) {
	return os.Args[0], []string{"-test.run=^$"}, nil
}

func (replayFlavor) Environment() map[string]string {
	session, _ := filepath.Abs(filepath.Join("testdata", "rust-analyzer", "session.json"))
	return map[string]string{fakeServerVariable: "1", fakeSessionVariable: session}
}
