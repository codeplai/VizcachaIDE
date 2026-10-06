package lsp

import (
	"context"
	"encoding/json"
	"os"
	"path/filepath"
	"testing"

	"github.com/codeplai/VizcachaIDE/wails/internal/domain"
)

const fakeInlayVariable = "VIZCACHA_FAKE_INLAY"

// fakeAnswer is what the fake server replies: with VIZCACHA_FAKE_INLAY=1 it announces the
// provider and always has a hint to give (so an answer of [] proves the server was not asked).
func fakeAnswer(method string) any {
	if os.Getenv(fakeRefactorVariable) == "1" {
		return refactorAnswer(method)
	}
	if os.Getenv(fakeInlayVariable) != "1" {
		return nil
	}
	switch method {
	case "initialize":
		return map[string]any{"capabilities": map[string]any{"inlayHintProvider": true}}
	case "textDocument/inlayHint":
		return []map[string]any{{"position": map[string]any{"line": 0, "character": 1}, "label": "int", "kind": 1}}
	}
	return nil
}

type inlayFlavor struct {
	fakeFlavor
	announce bool
}

func (f *inlayFlavor) Environment() map[string]string {
	env := f.fakeFlavor.Environment()
	if f.announce {
		env[fakeInlayVariable] = "1"
	}
	return env
}

func fakeInlayHints(t *testing.T, announce bool) []domain.InlayHint {
	t.Helper()
	file := filepath.Join(t.TempDir(), "main.go") // first: cleanups run in reverse, the server stops before
	server := New(newRecordingSink(), &inlayFlavor{announce: announce}, Options{})
	t.Cleanup(func() { _ = server.Shutdown(context.Background()) })
	if err := server.OpenDocument(context.Background(), file, "x := 5\n"); err != nil {
		t.Fatal(err)
	}
	waitFor(t, "server ready", func() bool { return server.hints.Load() == announce && server.snapshotReady() })
	visible := domain.SourceRange{
		Start: domain.SourceLocation{File: file, Line: 1, Column: 1},
		End:   domain.SourceLocation{File: file, Line: 1, Column: 1},
	}
	hints, err := server.InlayHints(context.Background(), visible)
	if err != nil {
		t.Fatal(err)
	}
	return hints
}

func (s *Server) snapshotReady() bool {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.state == stateReady
}

func TestInlayHintsAreAskedOnlyWhenTheServerAnnouncesThem(t *testing.T) {
	if hints := fakeInlayHints(t, false); len(hints) != 0 {
		t.Errorf("without inlayHintProvider the server is not asked: %+v", hints)
	}
	hints := fakeInlayHints(t, true)
	if len(hints) != 1 || hints[0] != (domain.InlayHint{Line: 1, Column: 2, Label: "int", Kind: domain.InlayHintType}) {
		t.Errorf("hints = %+v", hints)
	}
}

func TestInlayHintsOfAnUnopenedDocumentAreEmpty(t *testing.T) {
	server := New(newRecordingSink(), missingToolFlavor{}, Options{})
	hints, err := server.InlayHints(context.Background(), domain.SourceRange{})
	if err != nil || hints == nil || len(hints) != 0 {
		t.Errorf("hints = %#v, %v", hints, err)
	}
}

func TestAnnouncesInlayHints(t *testing.T) {
	cases := map[string]bool{
		`{"capabilities":{"inlayHintProvider":true}}`:                      true,
		`{"capabilities":{"inlayHintProvider":{"resolveProvider":false}}}`: true,
		`{"capabilities":{"inlayHintProvider":false}}`:                     false,
		`{"capabilities":{"inlayHintProvider":null}}`:                      false,
		`{"capabilities":{"hoverProvider":true}}`:                          false,
		`not json`: false,
	}
	for result, want := range cases {
		if got := announcesInlayHints(json.RawMessage(result)); got != want {
			t.Errorf("%s: %v, want %v", result, got, want)
		}
	}
}

func TestInlayLabelsAreStringsOrJoinedParts(t *testing.T) {
	text := "x := 1\n"
	raw := json.RawMessage(`[
		{"position":{"line":0,"character":1},"label":": int","kind":1,"paddingLeft":true},
		{"position":{"line":0,"character":5},"label":[{"value":"name"},{"value":":","tooltip":"t"}],"kind":2,"paddingRight":true},
		{"position":{"line":0,"character":6},"label":"","kind":1},
		{"position":{"line":0,"character":6},"label":"fin","kind":3}
	]`)
	got := toInlayHints(raw, text)
	want := []domain.InlayHint{
		{Line: 1, Column: 2, Label: ": int", Kind: domain.InlayHintType, PaddingLeft: true},
		{Line: 1, Column: 6, Label: "name:", Kind: domain.InlayHintParameter, PaddingRight: true},
		{Line: 1, Column: 7, Label: "fin", Kind: domain.InlayHintOther},
	}
	if len(got) != len(want) {
		t.Fatalf("hints = %+v", got)
	}
	for i := range want {
		if got[i] != want[i] {
			t.Errorf("hint %d = %+v, want %+v", i, got[i], want[i])
		}
	}
	if empty := toInlayHints(json.RawMessage("null"), text); empty == nil || len(empty) != 0 {
		t.Errorf("null must map to an empty slice, got %#v", empty)
	}
}

func TestInlayColumnsCountRunesNotUTF16Units(t *testing.T) {
	// "é" is one UTF-16 unit, the emoji two: UTF-16 offset 9 is rune 8, the closing quote (column 9).
	text := "s := \"é😀\"\n"
	raw := json.RawMessage(`[{"position":{"line":0,"character":1},"label":"a"},{"position":{"line":0,"character":9},"label":"b"}]`)
	got := toInlayHints(raw, text)
	if got[0].Column != 2 || got[1].Column != 9 {
		t.Errorf("columns = %d, %d", got[0].Column, got[1].Column)
	}
}
