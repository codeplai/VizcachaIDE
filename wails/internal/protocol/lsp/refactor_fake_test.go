package lsp

import (
	"context"
	"path/filepath"
	"testing"

	"github.com/codeplai/VizcachaIDE/wails/internal/domain"
)

const fakeRefactorVariable = "VIZCACHA_FAKE_REFACTOR"

// refactorAnswer is the fake server of the rename tests: it announces rename with prepare
// support, refuses to prepare anything (null) and renames with a documentChanges edit.
func refactorAnswer(method string) any {
	switch method {
	case "initialize":
		return map[string]any{"capabilities": map[string]any{
			"renameProvider": map[string]any{"prepareProvider": true}, "referencesProvider": true,
		}}
	case "textDocument/rename":
		return map[string]any{"documentChanges": []any{map[string]any{
			"textDocument": map[string]any{"uri": "file:///nowhere/a.go", "version": 1},
			"edits": []any{map[string]any{
				"range":   map[string]any{"start": map[string]any{"line": 0, "character": 0}, "end": map[string]any{"line": 0, "character": 1}},
				"newText": "y",
			}},
		}}}
	case "textDocument/references":
		return []any{}
	}
	return nil
}

type refactorFlavor struct{ fakeFlavor }

func (f *refactorFlavor) Environment() map[string]string {
	env := f.fakeFlavor.Environment()
	env[fakeRefactorVariable] = "1"
	return env
}

func TestFakeServerRenameFlow(t *testing.T) {
	file := filepath.Join(t.TempDir(), "main.go")
	server := New(newRecordingSink(), &refactorFlavor{}, Options{})
	t.Cleanup(func() { _ = server.Shutdown(context.Background()) })
	if err := server.OpenDocument(context.Background(), file, "x := 5\n"); err != nil {
		t.Fatal(err)
	}
	waitFor(t, "server ready", func() bool { return server.refactor.prepare.Load() && server.snapshotReady() })
	at := domain.SourceLocation{File: file, Line: 1, Column: 1}
	target, err := server.PrepareRename(context.Background(), at)
	if err != nil || target.Refusal != domain.RenameNotRenameable {
		t.Fatalf("prepare must refuse a null answer: %+v %v", target, err)
	}
	result, err := server.Rename(context.Background(), at, "y")
	if err != nil || result.Refusal != domain.RenameAllowed || len(result.Files) != 1 || result.Files[0].Edits[0].NewText != "y" {
		t.Fatalf("rename: %+v %v", result, err)
	}
	refs, err := server.References(context.Background(), at)
	if err != nil || len(refs) != 0 {
		t.Fatalf("references: %v %v", refs, err)
	}
}
