package clangd

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/codeplai/VizcachaIDE/wails/internal/adapters/cpp/cpptest"
	"github.com/codeplai/VizcachaIDE/wails/internal/domain"
	"github.com/codeplai/VizcachaIDE/wails/internal/protocol/lsp"
)

const (
	headerSource = "#pragma once\n\ninline int twice(int n) {\n    return n * 2;\n}\n"
	callerSource = "#include \"helper.h\"\n\nint main() {\n    int a = twice(2);\n    return twice(a);\n}\n"
)

// Renames a function declared in a header and used by main.cpp, and lists its references.
func TestRealClangdRenamesAcrossAHeaderAndFindsReferences(t *testing.T) {
	llvm := cpptest.LLVMBin(t)
	dir := t.TempDir()
	header, caller := filepath.Join(dir, "helper.h"), filepath.Join(dir, "main.cpp")
	server := lsp.New(&recordingSink{diagnostics: map[string][]domain.Diagnostic{}}, NewFlavor(Config{Locator: locatorFor(llvm, llvm)}),
		lsp.Options{Name: serverName, LanguageID: languageID, CodeLanguage: domain.CodeLanguageCpp})
	t.Cleanup(func() { _ = server.Shutdown(context.Background()) })
	for path, text := range map[string]string{header: headerSource, caller: callerSource} {
		if err := os.WriteFile(path, []byte(text), 0o600); err != nil {
			t.Fatal(err)
		}
		if err := server.OpenDocument(context.Background(), path, text); err != nil {
			t.Fatal(err)
		}
	}
	at := domain.SourceLocation{File: caller, Line: 5, Column: strings.Index("    return twice(a);", "twice") + 1}
	result := eventually(t, func() (domain.RenameResult, bool) {
		result, _ := server.Rename(context.Background(), at, "doubled")
		return result, len(result.Files) == 2
	})
	want := map[string]string{
		"helper.h": strings.ReplaceAll(headerSource, "twice", "doubled"),
		"main.cpp": strings.ReplaceAll(callerSource, "twice", "doubled"),
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
	refs := eventually(t, func() ([]domain.Reference, bool) {
		refs, _ := server.References(context.Background(), at)
		return refs, len(refs) >= 3
	})
	if len(refs) != 3 || !strings.HasSuffix(refs[0].Range.Start.File, "helper.h") {
		t.Errorf("references = %+v", refs)
	}
}
