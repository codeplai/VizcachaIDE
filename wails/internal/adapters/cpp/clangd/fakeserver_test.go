package clangd

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
	fakeServerVariable  = "VIZCACHA_FAKE_CLANGD"
	fakeSessionVariable = "VIZCACHA_FAKE_CLANGD_SESSION"
)

// TestMain turns the test binary into a language server that replays the answers recorded
// from the real clangd (testdata/clangd/session.json) when the variable is set.
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
	"textDocument/completion":     "completion",
	"textDocument/hover":          "hover",
	"textDocument/signatureHelp":  "signatureHelp",
	"textDocument/definition":     "definition",
	"textDocument/documentSymbol": "documentSymbol",
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
		switch req.Method() {
		case "exit":
			os.Exit(0)
		case "textDocument/didOpen":
			var params struct {
				TextDocument struct {
					URI string `json:"uri"`
				} `json:"textDocument"`
			}
			_ = json.Unmarshal(req.Params(), &params)
			docURI = params.TextDocument.URI
			return conn.Notify(ctx, "textDocument/publishDiagnostics", withURI(session["publishDiagnostics"], docURI))
		}
		if name, known := answers[req.Method()]; known {
			if req.Method() == "textDocument/completion" && completionLine(req.Params()) == memberCompletionLine {
				name = "memberCompletion"
			}
			return reply(ctx, withURI(session[name], docURI), nil)
		}
		return reply(ctx, nil, nil)
	})
	<-conn.Done()
}

// memberCompletionLine is the 0-based line of "v.push_back" in main.cpp: the other completion
// is asked on the "std::vec" line.
const memberCompletionLine = 14

func completionLine(params json.RawMessage) int {
	var position struct {
		Position struct {
			Line int `json:"line"`
		} `json:"position"`
	}
	_ = json.Unmarshal(params, &position)
	return position.Position.Line
}

// withURI puts the URI of the opened document where the recording has {DOC_URI}.
func withURI(recorded json.RawMessage, docURI string) json.RawMessage {
	escaped, _ := json.Marshal(docURI)
	return json.RawMessage(strings.ReplaceAll(string(recorded), `"`+docURIPlaceholder+`"`, string(escaped)))
}

// replayFlavor is the clangd flavor, started as the replay server.
type replayFlavor struct{ Flavor }

func (replayFlavor) Command(map[string]string) (string, []string, error) {
	return os.Args[0], []string{"-test.run=^$"}, nil
}

func (replayFlavor) Environment() map[string]string {
	session, _ := filepath.Abs(filepath.Join("testdata", "clangd", "session.json"))
	return map[string]string{fakeServerVariable: "1", fakeSessionVariable: session}
}
