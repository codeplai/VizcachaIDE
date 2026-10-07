package lsp

import (
	"testing"

	"github.com/codeplai/VizcachaIDE/wails/internal/domain"
)

func span(line, from, to int) *domain.SourceRange {
	return &domain.SourceRange{
		Start: domain.SourceLocation{Line: line, Column: from},
		End:   domain.SourceLocation{Line: line, Column: to},
	}
}

func TestVetRenameTarget(t *testing.T) {
	server := &Server{opts: Options{CodeLanguage: domain.CodeLanguageGo}}
	doc := document{path: "main.go", text: "func año(x int) int {\n\treturn x + 1\n}\n"}
	at := func(line, col int) domain.SourceLocation {
		return domain.SourceLocation{File: "main.go", Line: line, Column: col}
	}

	// gopls answers the range of the keyword and the placeholder "func()".
	got := server.vetRenameTarget(doc, at(1, 2), domain.RenameTarget{Range: span(1, 1, 5), Placeholder: "func()"})
	if got.Refusal != domain.RenameNotRenameable {
		t.Errorf("keyword: %+v", got)
	}
	if got := server.vetRenameTarget(doc, at(2, 2), domain.RenameTarget{}); got.Refusal != domain.RenameNotRenameable {
		t.Errorf("return: %+v", got)
	}
	if got := server.vetRenameTarget(doc, at(2, 11), domain.RenameTarget{}); got.Refusal != domain.RenameNotRenameable {
		t.Errorf("operator (no word under the cursor): %+v", got)
	}
	got = server.vetRenameTarget(doc, at(1, 6), domain.RenameTarget{Range: span(1, 6, 9), Placeholder: "año"})
	if got.Refusal != domain.RenameAllowed || got.Placeholder != "año" {
		t.Errorf("identifier: %+v", got)
	}
	// A server without prepareRename: the word under the cursor, in runes.
	got = server.vetRenameTarget(doc, at(1, 7), domain.RenameTarget{})
	if got.Refusal != domain.RenameAllowed || got.Placeholder != "año" || got.Range.Start.Column != 6 || got.Range.End.Column != 9 {
		t.Errorf("word: %+v %+v", got, got.Range)
	}
	if got := server.vetRenameTarget(doc, at(1, 1), domain.RenameTarget{Refusal: domain.RenameUnsupported}); got.Refusal != domain.RenameUnsupported {
		t.Errorf("a refusal stays: %+v", got)
	}
}
