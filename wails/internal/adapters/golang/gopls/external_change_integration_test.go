package gopls

import (
	"context"
	"os"
	"strings"
	"testing"

	"github.com/codeplai/VizcachaIDE/wails/internal/domain"
)

// gopls offers to rename the "func" of a declaration (placeholder "func()"): the editor refuses.
func TestRealGoplsKeywordFuncIsNotRenameable(t *testing.T) {
	server, _, caller := refactorProject(t)
	at := domain.SourceLocation{File: caller, Line: 5, Column: 2}
	got := eventually(t, func() (domain.RenameTarget, bool) {
		target, _ := server.PrepareRename(context.Background(), at)
		return target, target.Refusal != domain.RenameFailed
	})
	if got.Refusal != domain.RenameNotRenameable {
		t.Errorf("func = %+v", got)
	}
}

// A closed file changed by another program (git, a text editor) reaches gopls without reopening.
func TestRealGoplsSeesAClosedFileChangedOutsideTheEditor(t *testing.T) {
	server, helper, caller := refactorProject(t)
	if err := server.CloseDocument(context.Background(), helper); err != nil {
		t.Fatal(err)
	}
	line := "\tfmt.Println(\"😀 año\", Greet(\"vizcacha\"))"
	call := domain.SourceLocation{File: caller, Line: 6, Column: len([]rune(line[:strings.Index(line, "Greet")])) + 1}
	eventually(t, func() (bool, bool) {
		target, _ := server.Definition(context.Background(), call)
		return true, target != nil
	})
	if err := os.WriteFile(helper, []byte(strings.ReplaceAll(helperSource, "Greet", "Other")), 0o600); err != nil {
		t.Fatal(err)
	}
	eventually(t, func() (bool, bool) {
		target, _ := server.Definition(context.Background(), call)
		return true, target == nil
	})
}
