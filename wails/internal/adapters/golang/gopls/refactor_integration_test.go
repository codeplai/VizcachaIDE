package gopls

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/codeplai/VizcachaIDE/wails/internal/domain"
	"github.com/codeplai/VizcachaIDE/wails/internal/protocol/lsp"
)

const (
	helperSource = "package main\n\n// Greet says hello to a name.\nfunc Greet(name string) string {\n\treturn \"hola \" + name\n}\n"
	callerSource = "package main\n\nimport \"fmt\"\n\nfunc main() {\n\tfmt.Println(\"😀 año\", Greet(\"vizcacha\"))\n\tfmt.Println(Greet(\"otra\"))\n}\n"
)

// refactorProject writes a two-file module (helper.go declares Greet, main.go calls it twice, the
// first call after an emoji and an accent so the columns need UTF-16 care) and opens both files.
func refactorProject(t *testing.T) (server *lsp.Server, helper, caller string) {
	t.Helper()
	if _, _, err := NewFlavor(Config{}).Command(nil); err != nil {
		t.Skip("gopls is not installed")
	}
	dir := t.TempDir()
	helper, caller = filepath.Join(dir, "helper.go"), filepath.Join(dir, "main.go")
	for name, text := range map[string]string{"go.mod": "module example.com/demo\n\ngo 1.21\n", helper: helperSource, caller: callerSource} {
		path := name
		if !filepath.IsAbs(path) {
			path = filepath.Join(dir, name)
		}
		if err := os.WriteFile(path, []byte(text), 0o600); err != nil {
			t.Fatal(err)
		}
	}
	server = New(newRecordingSink(), Config{}, lsp.Options{})
	t.Cleanup(func() { _ = server.Shutdown(context.Background()) })
	for path, text := range map[string]string{helper: helperSource, caller: callerSource} {
		if err := server.OpenDocument(context.Background(), path, text); err != nil {
			t.Fatal(err)
		}
	}
	return server, helper, caller
}

func TestRealGoplsRenamesAFunctionAcrossTwoFiles(t *testing.T) {
	server, _, caller := refactorProject(t)
	at := domain.SourceLocation{File: caller, Line: 7, Column: strings.Index("\tfmt.Println(Greet(\"otra\"))", "Greet") + 1}
	target := eventually(t, func() (domain.RenameTarget, bool) {
		target, _ := server.PrepareRename(context.Background(), at)
		return target, target.Range != nil
	})
	if target.Refusal != domain.RenameAllowed || target.Placeholder != "Greet" {
		t.Fatalf("prepareRename = %+v", target)
	}
	result, err := server.Rename(context.Background(), at, "Salute")
	if err != nil || result.Refusal != domain.RenameAllowed || len(result.Files) != 2 {
		t.Fatalf("rename = %+v, %v", result, err)
	}
	want := map[string]string{
		"helper.go": strings.ReplaceAll(helperSource, "Greet", "Salute"),
		"main.go":   strings.ReplaceAll(callerSource, "Greet", "Salute"),
	}
	for _, file := range result.Files {
		current, err := os.ReadFile(file.File)
		if err != nil {
			t.Fatal(err)
		}
		edited, err := domain.ApplyTextEdits(string(current), file.Edits)
		if err != nil {
			t.Fatal(err)
		}
		if expected := want[filepath.Base(file.File)]; edited != expected {
			t.Errorf("%s:\n%s\nwant:\n%s", file.File, edited, expected)
		}
	}
}

func TestRealGoplsRefusesToRenameAKeywordAndFindsReferences(t *testing.T) {
	server, helper, _ := refactorProject(t)
	keyword := domain.SourceLocation{File: helper, Line: 5, Column: 3} // "return"
	refused := eventually(t, func() (domain.RenameTarget, bool) {
		target, _ := server.PrepareRename(context.Background(), keyword)
		return target, target.Refusal != domain.RenameAllowed
	})
	if refused.Refusal != domain.RenameNotRenameable {
		t.Errorf("keyword refusal = %+v", refused)
	}
	declaration := domain.SourceLocation{File: helper, Line: 4, Column: 7}
	refs := eventually(t, func() ([]domain.Reference, bool) {
		refs, _ := server.References(context.Background(), declaration)
		return refs, len(refs) >= 3
	})
	if len(refs) != 3 {
		t.Fatalf("references = %+v", refs)
	}
	if refs[0].Preview != "func Greet(name string) string {" || !strings.HasSuffix(refs[0].Range.Start.File, "helper.go") {
		t.Errorf("first reference = %+v", refs[0])
	}
	if refs[1].Range.Start.Line != 6 || refs[1].Range.Start.Column != 23 {
		t.Errorf("columns are in runes (emoji counts once): %+v", refs[1].Range)
	}
}
