package gopls

import (
	"context"
	"strings"
	"testing"

	"github.com/codeplai/VizcachaIDE/wails/internal/domain"
)

const inlaySource = "package main\n\nfunc add(a, b int) int {\n\treturn a + b\n}\n\nfunc main() {\n\tx := 5\n\t_ = add(x, 2)\n}\n"

func TestFlavorTurnsOnTheHintsABeginnerReads(t *testing.T) {
	options, _ := Flavor{}.InitializationOptions().(map[string]any)
	hints, _ := options["hints"].(map[string]any)
	for _, name := range []string{"assignVariableTypes", "parameterNames", "rangeVariableTypes"} {
		if hints[name] != true {
			t.Errorf("hint %s must be on: %v", name, hints)
		}
	}
	if hints["compositeLiteralFields"] != false {
		t.Errorf("compositeLiteralFields adds noise: %v", hints)
	}
}

func TestRealGoplsGivesTheTypeOfAnAssignmentAndTheParameterNames(t *testing.T) {
	session := startSession(t)
	if err := session.server.ChangeDocument(context.Background(), session.file, inlaySource, 1); err != nil {
		t.Fatal(err)
	}
	visible := domain.SourceRange{
		Start: domain.SourceLocation{File: session.file, Line: 1, Column: 1},
		End:   domain.SourceLocation{File: session.file, Line: 10, Column: 1},
	}
	hints := eventually(t, func() ([]domain.InlayHint, bool) {
		hints, _ := session.server.InlayHints(context.Background(), visible)
		return hints, len(hints) >= 3
	})
	var typed, named []string
	for _, hint := range hints {
		switch hint.Kind {
		case domain.InlayHintType:
			typed = append(typed, hint.Label)
			if hint.Line != 8 || hint.Column != 3 {
				t.Errorf("type hint at %d:%d, want 8:3 (after x)", hint.Line, hint.Column)
			}
		case domain.InlayHintParameter:
			named = append(named, strings.TrimSpace(hint.Label))
		}
	}
	if len(typed) != 1 || !strings.Contains(typed[0], "int") {
		t.Errorf("type hints = %v (%+v)", typed, hints)
	}
	if strings.Join(named, ",") != "a:,b:" {
		t.Errorf("parameter hints = %v (%+v)", named, hints)
	}
}
