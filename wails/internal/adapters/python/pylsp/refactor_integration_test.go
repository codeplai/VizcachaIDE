package pylsp

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/codeplai/VizcachaIDE/wails/internal/adapters/python/pythontest"
	"github.com/codeplai/VizcachaIDE/wails/internal/domain"
	"github.com/codeplai/VizcachaIDE/wails/internal/protocol/lsp"
)

const (
	libSource = "def saludar(nombre):\n    return \"hola \" + nombre\n"
	appSource = "from lib import saludar\n\nprint(\"año\", saludar(\"vizcacha\"))\nprint(saludar(\"otra\"))\n"
)

// refactorServer opens lib.py and app.py (which imports it) in the real python-lsp-server.
func refactorServer(t *testing.T) (server *lsp.Server, library, app, interpreter string) {
	t.Helper()
	interpreter = pythontest.Interpreter(t)
	dir := t.TempDir()
	library, app = filepath.Join(dir, "lib.py"), filepath.Join(dir, "app.py")
	server = lsp.New(newRecordingSink(), NewFlavor(configAt(interpreter)), lsp.Options{Name: serverName, LanguageID: languageID})
	t.Cleanup(func() { _ = server.Shutdown(context.Background()) })
	for path, text := range map[string]string{library: libSource, app: appSource} {
		if err := os.WriteFile(path, []byte(text), 0o600); err != nil {
			t.Fatal(err)
		}
		if err := server.OpenDocument(context.Background(), path, text); err != nil {
			t.Fatal(err)
		}
	}
	return server, library, app, interpreter
}

func TestRealPylspFindsReferencesAcrossFiles(t *testing.T) {
	server, library, app, _ := refactorServer(t)
	declaration := domain.SourceLocation{File: library, Line: 1, Column: 6}
	refs := eventually(t, func() ([]domain.Reference, bool) {
		refs, _ := server.References(context.Background(), declaration)
		return refs, len(refs) >= 4
	})
	inApp := 0
	for _, ref := range refs {
		if strings.EqualFold(filepath.Clean(ref.Range.Start.File), filepath.Clean(app)) {
			inApp++
			if ref.Range.Start.Line == 3 && ref.Range.Start.Column != 14 {
				t.Errorf("columns are in runes: %+v", ref.Range)
			}
		}
	}
	if inApp < 3 {
		t.Errorf("want the import and two calls in app.py, got %+v", refs)
	}
}

// pylsp 1.15 renames with jedi (or with rope when installed) and answers one edit per file that
// replaces the whole text. An older pylsp without either answers no edit, and the editor then
// shows its translated message.
func TestRealPylspRenamesAcrossFilesOrSaysItCannot(t *testing.T) {
	server, library, _, _ := refactorServer(t)
	at := domain.SourceLocation{File: library, Line: 1, Column: 6}
	var result domain.RenameResult
	eventually(t, func() (bool, bool) {
		result, _ = server.Rename(context.Background(), at, "saludo")
		return true, len(result.Files) > 0 || result.Refusal != domain.RenameAllowed
	})
	if len(result.Files) == 0 {
		return // the refusal is what the editor translates
	}
	if len(result.Files) != 2 {
		t.Fatalf("rename = %+v", result)
	}
	for _, file := range result.Files {
		current, err := os.ReadFile(file.File)
		if err != nil {
			t.Fatal(err)
		}
		edited, err := domain.ApplyTextEdits(string(current), file.Edits)
		if err != nil || strings.Contains(edited, "saludar") || !strings.Contains(edited, "saludo") {
			t.Errorf("%s: %q, %v", file.File, edited, err)
		}
	}
}
