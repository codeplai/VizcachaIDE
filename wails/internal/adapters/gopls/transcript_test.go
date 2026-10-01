package gopls

import (
	"bytes"
	"context"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"go.lsp.dev/jsonrpc2"
	"go.lsp.dev/protocol"

	"github.com/codeplai/VizcachaIDE/wails/internal/domain"
)

// recorded is the session recorded from gopls v0.23 (shared with the 1.0 tests), re-rooted
// in a temp folder where main.go is written.
type recorded struct {
	messages map[string]json.RawMessage
	source   string
	text     string
}

func loadRecorded(t *testing.T) recorded {
	t.Helper()
	dir := t.TempDir()
	text, err := os.ReadFile(filepath.Join("testdata", "session_main.go"))
	if err != nil {
		t.Fatal(err)
	}
	source := filepath.Join(dir, "main.go")
	if err := os.WriteFile(source, text, 0o600); err != nil {
		t.Fatal(err)
	}
	raw, err := os.ReadFile(filepath.Join("testdata", "gopls_session.json"))
	if err != nil {
		t.Fatal(err)
	}
	rooted := strings.ReplaceAll(string(raw), "{ROOT_URI}", string(pathToURI(dir)))
	var messages map[string]json.RawMessage
	if err := json.Unmarshal([]byte(rooted), &messages); err != nil {
		t.Fatal(err)
	}
	return recorded{messages: messages, source: source, text: string(text)}
}

// result returns the "result" of a recorded response, or its "params" for notifications.
func (r recorded) result(t *testing.T, name string) json.RawMessage {
	t.Helper()
	var message struct {
		Result json.RawMessage `json:"result"`
		Params json.RawMessage `json:"params"`
	}
	if err := json.Unmarshal(r.messages[name], &message); err != nil {
		t.Fatal(err)
	}
	if message.Result != nil {
		return message.Result
	}
	return message.Params
}

func TestFramingRoundTripsTheRecordedSessionInSmallChunks(t *testing.T) {
	session := loadRecorded(t)
	var wire bytes.Buffer
	writer := jsonrpc2.NewStream(nopCloser{&wire})
	var sent []jsonrpc2.Message
	for _, name := range []string{"showMessage", "publishDiagnostics", "completion"} {
		message, err := jsonrpc2.DecodeMessage(session.messages[name])
		if err != nil {
			t.Fatalf("%s: %v", name, err)
		}
		sent = append(sent, message)
		if _, err := writer.Write(context.Background(), message); err != nil {
			t.Fatal(err)
		}
	}
	reader := jsonrpc2.NewStream(nopCloser{&oneByteReader{data: wire.Bytes()}})
	for i := range sent {
		got, _, err := reader.Read(context.Background())
		if err != nil {
			t.Fatalf("message %d: %v", i, err)
		}
		want, _ := json.Marshal(sent[i])
		have, _ := json.Marshal(got)
		if !bytes.Equal(want, have) {
			t.Errorf("message %d changed:\n%s\n%s", i, want, have)
		}
	}
}

func TestPublishedDiagnosticsBecomeDomainDiagnostics(t *testing.T) {
	session := loadRecorded(t)
	var params protocol.PublishDiagnosticsParams
	if err := json.Unmarshal(session.result(t, "publishDiagnostics"), &params); err != nil {
		t.Fatal(err)
	}
	got := toDiagnostics(params, session.source, session.text)
	if len(got) != 1 {
		t.Fatalf("diagnostics = %+v", got)
	}
	d := got[0]
	if *d.Location != (domain.SourceLocation{File: session.source, Line: 10, Column: 2}) ||
		*d.End != (domain.SourceLocation{File: session.source, Line: 10, Column: 3}) {
		t.Errorf("range = %+v - %+v", *d.Location, *d.End)
	}
	if d.Severity != domain.SeverityError || d.Message != "declared and not used: x" ||
		d.RawText != d.Message || d.Source != "gopls" || d.Code != "UnusedVar" {
		t.Errorf("diagnostic = %+v", d)
	}
}

func TestCompletionHoverAndSignatureMapping(t *testing.T) {
	session := loadRecorded(t)
	items := toCompletionItems(session.result(t, "completion"))
	if len(items) == 0 || items[0].Label != "Print" || items[0].Kind != domain.CompletionFunction {
		t.Fatalf("items = %+v", items)
	}
	if items[0].InsertText != "" || !strings.Contains(items[0].Documentation, "standard output") {
		t.Errorf("first item = %+v", items[0])
	}
	if hover := toHoverText(session.result(t, "hover")); !strings.HasPrefix(hover, "func fmt.Println") {
		t.Errorf("hover = %q", hover)
	}
	help := toSignatureHelp(session.result(t, "signatureHelp"))
	if help == nil || help.Label != "greet(name string) string" || len(help.Parameters) != 1 ||
		help.Parameters[0] != "name string" || help.ActiveParameter != 0 {
		t.Errorf("signature = %+v", help)
	}
}

func TestDefinitionHighlightsAndSymbolsMapping(t *testing.T) {
	session := loadRecorded(t)
	target := toDefinition(session.result(t, "definition"), func(string) string { return session.text })
	if target == nil || target.Line != 5 || target.Column != 6 {
		t.Errorf("definition = %+v", target)
	}
	ranges := toHighlightRanges(session.result(t, "documentHighlight"), session.source, session.text)
	if len(ranges) != 2 || ranges[0].Start.Line != 5 || ranges[0].End.Column != 11 {
		t.Errorf("highlights = %+v", ranges)
	}
	symbols := toDocumentSymbols(session.result(t, "documentSymbol"), session.source, session.text)
	if len(symbols) != 2 || symbols[0].Name != "greet" || symbols[0].Kind != domain.SymbolFunction ||
		symbols[0].Location.Line != 5 || symbols[0].Range.End.Line != 7 || symbols[1].Location.Line != 9 {
		t.Errorf("symbols = %+v", symbols)
	}
}

func TestEmptyResultsMapToNothing(t *testing.T) {
	null := json.RawMessage("null")
	if toHoverText(null) != "" || toSignatureHelp(null) != nil || len(toCompletionItems(null)) != 0 ||
		toDefinition(null, nil) != nil || len(toHighlightRanges(null, "f", "")) != 0 {
		t.Error("null results must map to empty values")
	}
}
