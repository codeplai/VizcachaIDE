package clangd

import (
	"context"
	"strings"
	"testing"

	"github.com/codeplai/VizcachaIDE/wails/internal/adapters/cpp/cpptest"
	"github.com/codeplai/VizcachaIDE/wails/internal/domain"
)

const inlaySource = "int twice(int value) { return value * 2; }\n\nint main() {\n    auto x = 5;\n    return twice(3) + x;\n}\n"

// clangd serves inlay hints by default (deduced types, parameter names, designators): the flavor
// only has to announce the client capability, which protocol/lsp does.
func TestRealClangdGivesTheDeducedTypeAndTheParameterName(t *testing.T) {
	llvm := cpptest.LLVMBin(t)
	s := openSample(t, NewFlavor(Config{Locator: locatorFor(llvm, llvm)}), true)
	if err := s.server.ChangeDocument(context.Background(), s.file, inlaySource, 1); err != nil {
		t.Fatal(err)
	}
	visible := domain.SourceRange{
		Start: domain.SourceLocation{File: s.file, Line: 1, Column: 1},
		End:   domain.SourceLocation{File: s.file, Line: 6, Column: 1},
	}
	hints := eventually(t, func() ([]domain.InlayHint, bool) {
		hints, _ := s.server.InlayHints(context.Background(), visible)
		return hints, len(hints) >= 2
	})
	var typed, named bool
	for _, hint := range hints {
		switch hint.Kind {
		case domain.InlayHintType:
			typed = typed || (hint.Line == 4 && strings.Contains(hint.Label, "int"))
		case domain.InlayHintParameter:
			named = named || (hint.Line == 5 && strings.Contains(hint.Label, "value"))
		}
	}
	if !typed || !named {
		t.Errorf("want ': int' on line 4 and 'value:' on line 5, got %+v", hints)
	}
}
