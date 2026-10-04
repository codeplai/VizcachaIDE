package pylsp

import (
	"context"
	"path/filepath"
	"strings"
	"testing"

	"github.com/codeplai/VizcachaIDE/wails/internal/domain"
)

// These tests replay what the real pylsp answered (testdata/pylsp/session.json) through the
// generic protocol/lsp client and check the domain values the frontend receives.

func openReplay(t *testing.T) sample { return openSample(t, replayFlavor{NewFlavor(Config{})}, true) }

func TestRecordedDiagnosticsBecomeDomainDiagnostics(t *testing.T) {
	s := openReplay(t)
	found := s.sink.diagnosticsOf(s.file)
	if len(found) != 2 {
		t.Fatalf("diagnostics = %+v", found)
	}
	d := found[0]
	if d.Message != "undefined name 'x'" || d.Source != serverName || d.Severity != domain.SeverityError {
		t.Errorf("diagnostic = %+v", d)
	}
	if d.Location.Line != 6 || d.Location.Column != 7 {
		t.Errorf("location = %+v, want 6:7", *d.Location)
	}
}

func TestRecordedCompletionOffersPrintToInsert(t *testing.T) {
	s := openReplay(t)
	items, err := s.server.Completion(context.Background(), atPri(s.file))
	if err != nil || len(items) != 1 {
		t.Fatalf("items = %+v, %v", items, err)
	}
	// pylsp labels the item with its parameters; the text to insert is just the name.
	if !strings.HasPrefix(items[0].Label, "print") || items[0].InsertText != "print" || items[0].Kind != domain.CompletionFunction {
		t.Errorf("item = %+v", items[0])
	}
}

func TestRecordedHoverAndSignatureHelp(t *testing.T) {
	s := openReplay(t)
	hover, _ := s.server.Hover(context.Background(), atLen(s.file))
	if !strings.Contains(hover, "Return the number of items") {
		t.Errorf("hover = %q", hover)
	}
	help, _ := s.server.SignatureHelp(context.Background(), insideGreet(s.file))
	if help == nil || help.Label != "greet(name)" || len(help.Parameters) != 1 || help.Parameters[0] != "name" || help.ActiveParameter != 0 {
		t.Errorf("signature help = %+v", help)
	}
}

func TestRecordedDefinitionAndSymbols(t *testing.T) {
	s := openReplay(t)
	target, _ := s.server.Definition(context.Background(), atGreetCall(s.file))
	if target == nil || target.Line != 1 || target.Column != 5 || !strings.EqualFold(filepath.Clean(target.File), filepath.Clean(s.file)) {
		t.Errorf("definition = %+v, want 1:5 in main.py", target)
	}
	// pylsp answers flat SymbolInformation (location.range, no range/selectionRange); the generic
	// mapping takes the extent from location.range.
	symbols, _ := s.server.DocumentSymbols(context.Background(), s.file)
	if len(symbols) != 1 || symbols[0].Name != "greet" || symbols[0].Kind != domain.SymbolFunction ||
		symbols[0].Location.Line != 1 || symbols[0].Range == nil || symbols[0].Range.End.Line < 2 {
		t.Errorf("symbols = %+v (range %+v)", symbols, symbols[0].Range)
	}
}
