package clangd

import (
	"context"
	"path/filepath"
	"strings"
	"testing"

	"github.com/codeplai/VizcachaIDE/wails/internal/domain"
)

// These tests replay what the real clangd answered (testdata/clangd/session.json) through the
// generic protocol/lsp client and check the domain values the frontend receives.

func openReplay(t *testing.T) sample { return openSample(t, replayFlavor{NewFlavor(Config{})}, true) }

func TestRecordedDiagnosticsNameTheUndeclaredIdentifier(t *testing.T) {
	s := openReplay(t)
	var undeclared *domain.Diagnostic
	for _, d := range s.sink.diagnosticsOf(s.file) {
		if strings.Contains(d.Message, "undeclared identifier 'missing'") {
			undeclared = &d
		}
	}
	if undeclared == nil {
		t.Fatalf("no diagnostic for missing: %+v", s.sink.diagnosticsOf(s.file))
	}
	if undeclared.Severity != domain.SeverityError || undeclared.Source != serverName || undeclared.Code != "undeclared_var_use" {
		t.Errorf("diagnostic = %+v", *undeclared)
	}
	if undeclared.Location.Line != 17 || undeclared.Location.Column != 18 || undeclared.End == nil || undeclared.End.Column != 25 {
		t.Errorf("range = %+v to %+v, want 17:18 to 17:25", *undeclared.Location, undeclared.End)
	}
}

func TestRecordedCompletionOffersVectorAndPushBack(t *testing.T) {
	s := openReplay(t)
	items, err := s.server.Completion(context.Background(), atStdVec(s.file))
	if err != nil || len(items) == 0 {
		t.Fatalf("items = %+v, %v", items, err)
	}
	// clangd's detailed style puts the template parameters in the label; the text to insert
	// is just the name.
	if !strings.Contains(items[0].Label, "vector") || items[0].InsertText != "vector" || items[0].Kind != domain.CompletionType {
		t.Errorf("item = %+v", items[0])
	}
	members, _ := s.server.Completion(context.Background(), atMemberPush(s.file))
	if len(members) == 0 || !strings.Contains(members[0].Label, "push_back(") || members[0].Detail != "void" ||
		members[0].InsertText != "push_back" || members[0].Kind != domain.CompletionMethod {
		t.Errorf("members = %+v", members)
	}
}

func TestRecordedHoverAndSignatureHelp(t *testing.T) {
	s := openReplay(t)
	hover, _ := s.server.Hover(context.Background(), atPushBack(s.file))
	if !strings.Contains(hover, "push_back") || !strings.Contains(hover, "provided by <vector>") {
		t.Errorf("hover = %q", hover)
	}
	help, _ := s.server.SignatureHelp(context.Background(), insideTwice(s.file))
	if help == nil || help.Label != "twice(int x) -> int" || len(help.Parameters) != 1 || help.Parameters[0] != "int x" || help.ActiveParameter != 0 {
		t.Errorf("signature help = %+v", help)
	}
}

func TestRecordedDefinitionAndNestedSymbols(t *testing.T) {
	s := openReplay(t)
	target, _ := s.server.Definition(context.Background(), atTwiceCall(s.file))
	if target == nil || target.Line != 11 || target.Column != 5 || !strings.EqualFold(filepath.Clean(target.File), filepath.Clean(s.file)) {
		t.Errorf("definition = %+v, want 11:5 in main.cpp", target)
	}
	symbols, _ := s.server.DocumentSymbols(context.Background(), s.file)
	if len(symbols) != 3 || symbols[0].Name != "Counter" || symbols[1].Name != "twice" || symbols[2].Name != "main" {
		t.Fatalf("symbols = %+v", symbols)
	}
	class := symbols[0]
	if len(class.Children) != 2 || class.Children[0].Name != "add" || class.Children[0].Kind != domain.SymbolMethod ||
		class.Children[0].Location.Line != 6 || class.Children[1].Kind != domain.SymbolField {
		t.Errorf("children of Counter = %+v", class.Children)
	}
}
