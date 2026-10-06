package lsp

import (
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
	"testing"

	"go.lsp.dev/jsonrpc2"

	"github.com/codeplai/VizcachaIDE/wails/internal/domain"
)

func textsOf(texts map[string]string) func(string) string {
	return func(path string) string { return texts[path] }
}

func TestWorkspaceEditChangesFormAcrossFiles(t *testing.T) {
	dir := t.TempDir()
	a, b := filepath.Join(dir, "a.go"), filepath.Join(dir, "b.go")
	raw := json.RawMessage(`{"changes":{
		"` + string(pathToURI(b)) + `":[{"range":{"start":{"line":0,"character":0},"end":{"line":0,"character":5}},"newText":"hello"}],
		"` + string(pathToURI(a)) + `":[
			{"range":{"start":{"line":0,"character":5},"end":{"line":0,"character":10}},"newText":"hello"},
			{"range":{"start":{"line":1,"character":0},"end":{"line":1,"character":5}},"newText":"hello"}]}}`)
	files, err := toWorkspaceEdit(raw, textsOf(map[string]string{a: "func greet() {}\ngreet()\n", b: "greet()\n"}))
	if err != nil {
		t.Fatal(err)
	}
	if len(files) != 2 || files[0].File != a || files[1].File != b || len(files[0].Edits) != 2 {
		t.Fatalf("files %+v", files)
	}
	if got := files[0].Edits[0].Range; got.Start.Line != 1 || got.Start.Column != 6 || got.End.Column != 11 {
		t.Fatalf("range %+v", got)
	}
}

func TestWorkspaceEditDocumentChangesMergesEntriesOfOneFile(t *testing.T) {
	a := filepath.Join(t.TempDir(), "a.py")
	uri := string(pathToURI(a))
	raw := json.RawMessage(`{"documentChanges":[
		{"textDocument":{"uri":"` + uri + `","version":1},"edits":[{"range":{"start":{"line":0,"character":4},"end":{"line":0,"character":7}},"newText":"x"}]},
		{"textDocument":{"uri":"` + uri + `","version":1},"edits":[{"range":{"start":{"line":1,"character":0},"end":{"line":1,"character":3}},"newText":"x","annotationId":"1"}]},
		{"textDocument":{"uri":"` + string(pathToURI(filepath.Join(filepath.Dir(a), "empty.py"))) + `"},"edits":[]}]}`)
	files, err := toWorkspaceEdit(raw, textsOf(map[string]string{a: "def foo():\nfoo()\n"}))
	if err != nil {
		t.Fatal(err)
	}
	if len(files) != 1 || len(files[0].Edits) != 2 {
		t.Fatalf("files %+v", files)
	}
}

func TestWorkspaceEditConvertsUTF16ColumnsToRunes(t *testing.T) {
	a := filepath.Join(t.TempDir(), "a.go")
	// "😀" takes two UTF-16 units, so "año" after it starts at character 2 and ends at 5.
	raw := json.RawMessage(`{"changes":{"` + string(pathToURI(a)) + `":[{"range":{"start":{"line":0,"character":2},"end":{"line":0,"character":5}},"newText":"x"}]}}`)
	files, err := toWorkspaceEdit(raw, textsOf(map[string]string{a: "😀año\n"}))
	if err != nil {
		t.Fatal(err)
	}
	got := files[0].Edits[0].Range
	if got.Start.Column != 2 || got.End.Column != 5 {
		t.Fatalf("columns in runes: %+v", got)
	}
	edited, err := domain.ApplyTextEdits("😀año\n", files[0].Edits)
	if err != nil || edited != "😀x\n" {
		t.Fatalf("edited %q, %v", edited, err)
	}
}

func TestWorkspaceEditRefusesFileOperations(t *testing.T) {
	raw := json.RawMessage(`{"documentChanges":[{"kind":"rename","oldUri":"file:///a","newUri":"file:///b"}]}`)
	if _, err := toWorkspaceEdit(raw, textsOf(nil)); !errors.Is(err, errResourceOperation) {
		t.Fatalf("got %v", err)
	}
}

func TestRenameTargetForms(t *testing.T) {
	doc := document{path: "main.go", text: "func greet() {}\n"}
	cases := map[string]struct {
		refusal     domain.RenameRefusal
		placeholder string
		hasRange    bool
	}{
		`null`: {refusal: domain.RenameNotRenameable},
		`{"start":{"line":0,"character":5},"end":{"line":0,"character":10}}`:                                 {hasRange: true},
		`{"range":{"start":{"line":0,"character":5},"end":{"line":0,"character":10}},"placeholder":"greet"}`: {placeholder: "greet", hasRange: true},
		`{"defaultBehavior":true}`: {},
	}
	for raw, want := range cases {
		got := toRenameTarget(json.RawMessage(raw), doc)
		if got.Refusal != want.refusal || got.Placeholder != want.placeholder || (got.Range != nil) != want.hasRange {
			t.Errorf("%s: got %+v", raw, got)
		}
	}
}

func TestRefusalMapsServerErrors(t *testing.T) {
	if reason, _ := refusal(errNotAnnounced); reason != domain.RenameUnsupported {
		t.Errorf("not announced: %v", reason)
	}
	if reason, _ := refusal(jsonrpc2.NewError(jsonrpc2.MethodNotFound, "no")); reason != domain.RenameUnsupported {
		t.Errorf("method not found: %v", reason)
	}
	if reason, detail := refusal(jsonrpc2.NewError(jsonrpc2.InternalError, "boom")); reason != domain.RenameFailed || detail != "boom" {
		t.Errorf("failed: %v %q", reason, detail)
	}
}

func TestReferencesAreMappedWithPreviewsAndSorted(t *testing.T) {
	dir := t.TempDir()
	a, b := filepath.Join(dir, "a.go"), filepath.Join(dir, "b.go")
	texts := map[string]string{a: "func greet() {}\n\tgreet()\n", b: "package b\n\n\n\tgreet()\n"}
	raw := json.RawMessage(`[
		{"uri":"` + string(pathToURI(b)) + `","range":{"start":{"line":3,"character":1},"end":{"line":3,"character":6}}},
		{"uri":"` + string(pathToURI(a)) + `","range":{"start":{"line":1,"character":1},"end":{"line":1,"character":6}}},
		{"uri":"` + string(pathToURI(a)) + `","range":{"start":{"line":0,"character":5},"end":{"line":0,"character":10}}}]`)
	refs := toReferences(raw, textsOf(texts))
	if len(refs) != 3 || refs[0].Range.Start.File != a || refs[0].Range.Start.Line != 1 || refs[2].Range.Start.File != b {
		t.Fatalf("refs %+v", refs)
	}
	if refs[1].Preview != "greet()" || refs[0].Preview != "func greet() {}" {
		t.Fatalf("previews %q %q", refs[0].Preview, refs[1].Preview)
	}
	if got := toReferences(json.RawMessage("null"), textsOf(nil)); got == nil || len(got) != 0 {
		t.Fatalf("null result: %#v", got)
	}
}

func TestCapabilitiesOfRenameAndReferences(t *testing.T) {
	var caps refactorCaps
	caps.store(json.RawMessage(`{"capabilities":{"renameProvider":{"prepareProvider":true},"referencesProvider":true}}`))
	if !caps.rename.Load() || !caps.prepare.Load() || !caps.references.Load() {
		t.Fatal("expected all announced")
	}
	caps.store(json.RawMessage(`{"capabilities":{"renameProvider":true}}`))
	if !caps.rename.Load() || caps.prepare.Load() || caps.references.Load() {
		t.Fatal("expected rename without prepare or references")
	}
}

func TestRenameOfAClosedDocumentIsRefused(t *testing.T) {
	server := New(newRecordingSink(), &fakeFlavor{}, Options{})
	file := filepath.Join(t.TempDir(), "x.go")
	if err := os.WriteFile(file, []byte("x"), 0o600); err != nil {
		t.Fatal(err)
	}
	at := domain.SourceLocation{File: file, Line: 1, Column: 1}
	result, err := server.Rename(t.Context(), at, "y")
	if err != nil || result.Refusal != domain.RenameNotRenameable || result.Files == nil {
		t.Fatalf("%+v %v", result, err)
	}
	refs, err := server.References(t.Context(), at)
	if err != nil || refs == nil || len(refs) != 0 {
		t.Fatalf("%#v %v", refs, err)
	}
}
