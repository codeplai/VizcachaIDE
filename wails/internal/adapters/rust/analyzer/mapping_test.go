package analyzer

import (
	"context"
	"path/filepath"
	"strings"
	"testing"

	"github.com/codeplai/VizcachaIDE/wails/internal/domain"
)

// These tests replay what the real rust-analyzer answered (testdata/rust-analyzer/session.json)
// through the generic protocol/lsp client and check the domain values the frontend receives.

func TestRecordedMovedValueIsE0382OnTheUse(t *testing.T) {
	s := openReplay(t)
	waitUntil(t, func() bool { _, found := s.sink.withCode(s.file, "E0382"); return found })
	moved, _ := s.sink.withCode(s.file, "E0382")
	if moved.Severity != domain.SeverityError || moved.Source != serverName || !strings.Contains(moved.Message, "borrow of moved value: `s`") {
		t.Errorf("diagnostic = %+v", moved)
	}
	// "println!("{} {} {}", s, t, y);" is line 21: the error is on the s (column 26).
	if moved.Location.Line != 21 || moved.Location.Column != 26 || moved.End == nil || moved.End.Column != 27 {
		t.Errorf("range = %+v to %+v, want 21:26 to 21:27", *moved.Location, moved.End)
	}
}

func TestRecordedUnresolvedNameIsE0425(t *testing.T) {
	s := openReplay(t)
	waitUntil(t, func() bool { _, found := s.sink.withCode(s.file, "E0382"); return found })
	if err := s.server.ChangeDocument(context.Background(), s.file, readTestdata(t, "unresolved.rs"), 2); err != nil {
		t.Fatal(err)
	}
	waitUntil(t, func() bool { _, found := s.sink.withCode(s.file, "E0425"); return found })
	unresolved, _ := s.sink.withCode(s.file, "E0425")
	if unresolved.Severity != domain.SeverityError || !strings.Contains(unresolved.Message, "cannot find value `missing`") {
		t.Errorf("diagnostic = %+v", unresolved)
	}
	if unresolved.Location.Line != 21 || unresolved.Location.Column != 26 {
		t.Errorf("location = %+v, want line 21 column 26", *unresolved.Location)
	}
}

func TestRecordedCompletionAfterVecDot(t *testing.T) {
	s := openReplay(t)
	items, err := s.server.Completion(context.Background(), afterVecDot(s.file))
	if err != nil || len(items) == 0 {
		t.Fatalf("items = %+v, %v", items, err)
	}
	var push *domain.CompletionItem
	for i := range items {
		if items[i].Label == "push" {
			push = &items[i]
		}
	}
	if push == nil || push.Kind != domain.CompletionMethod || !strings.Contains(push.Detail, "fn(&mut self, T)") || push.TextToInsert() != "push" {
		t.Errorf("push = %+v in %+v", push, items)
	}
}

func TestRecordedHoverOnPushHasSignatureAndDocumentation(t *testing.T) {
	s := openReplay(t)
	hover, _ := s.server.Hover(context.Background(), atPush(s.file))
	if !strings.Contains(hover, "pub fn push(&mut self, value: T)") || !strings.Contains(hover, "Appends an element to the back of a collection") {
		t.Errorf("hover = %q", hover)
	}
}

func TestRecordedSignatureHelpOfTwice(t *testing.T) {
	s := openReplay(t)
	help, _ := s.server.SignatureHelp(context.Background(), insideTwice(s.file))
	if help == nil || help.Label != "fn twice(x: i32) -> i32" || len(help.Parameters) != 1 || help.Parameters[0] != "x: i32" || help.ActiveParameter != 0 {
		t.Errorf("signature help = %+v", help)
	}
}

func TestRecordedDefinitionOfTwice(t *testing.T) {
	s := openReplay(t)
	target, _ := s.server.Definition(context.Background(), atTwiceCall(s.file))
	if target == nil || target.Line != 11 || target.Column != 4 || !strings.EqualFold(filepath.Clean(target.File), filepath.Clean(s.file)) {
		t.Errorf("definition = %+v, want 11:4 in main.rs", target)
	}
}

// rust-analyzer lists an impl as a symbol of its own ("impl Counter") that holds the methods;
// the struct keeps its fields.
func TestRecordedSymbolsNestTheMethodUnderItsImpl(t *testing.T) {
	s := openReplay(t)
	symbols, _ := s.server.DocumentSymbols(context.Background(), s.file)
	if len(symbols) != 4 || symbols[0].Name != "Counter" || symbols[1].Name != "impl Counter" || symbols[2].Name != "twice" || symbols[3].Name != "main" {
		t.Fatalf("symbols = %+v", symbols)
	}
	if fields := symbols[0].Children; len(fields) != 1 || fields[0].Name != "total" || fields[0].Kind != domain.SymbolField {
		t.Errorf("children of Counter = %+v", fields)
	}
	methods := symbols[1].Children
	if len(methods) != 1 || methods[0].Name != "add" || methods[0].Kind != domain.SymbolMethod || methods[0].Location.Line != 6 {
		t.Errorf("children of impl Counter = %+v", methods)
	}
}
