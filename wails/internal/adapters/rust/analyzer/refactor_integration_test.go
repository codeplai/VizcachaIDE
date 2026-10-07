package analyzer

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/codeplai/VizcachaIDE/wails/internal/domain"
)

const (
	utilSource = "pub fn twice(n: i32) -> i32 {\n    n * 2\n}\n"
	mainSource = "mod util;\n\nfn main() {\n    println!(\"año {}\", util::twice(2));\n    println!(\"{}\", util::twice(3));\n}\n"
)

// These need the toolchain of docs/PLAN_RUST.md (they skip without it): rename a function of a
// module used from main.rs and list its references.
func openRefactorCrate(t *testing.T) (s sample, util string) {
	t.Helper()
	folder := t.TempDir()
	writeManifest(t, folder, readTestdata(t, filepath.Join("crate", "Cargo.toml")))
	util = filepath.Join(folder, "src", "util.rs")
	if err := os.MkdirAll(filepath.Dir(util), 0o750); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(util, []byte(utilSource), 0o600); err != nil {
		t.Fatal(err)
	}
	s = openIn(t, NewFlavor(Config{Locator: locatorForTests(t)}), filepath.Join(folder, "src", "main.rs"), mainSource)
	if err := s.server.OpenDocument(context.Background(), util, utilSource); err != nil {
		t.Fatal(err)
	}
	return s, util
}

func TestRealRustAnalyzerRenamesAFunctionAcrossTwoFiles(t *testing.T) {
	s, _ := openRefactorCrate(t)
	at := domain.SourceLocation{File: s.file, Line: 5, Column: strings.Index("    println!(\"{}\", util::twice(3));", "twice") + 1}
	target := eventually(t, func() (domain.RenameTarget, bool) {
		target, _ := s.server.PrepareRename(context.Background(), at)
		return target, target.Range != nil
	})
	if target.Refusal != domain.RenameAllowed {
		t.Fatalf("prepareRename = %+v", target)
	}
	result := eventually(t, func() (domain.RenameResult, bool) {
		result, _ := s.server.Rename(context.Background(), at, "doubled")
		return result, len(result.Files) == 2
	})
	want := map[string]string{
		"util.rs": strings.ReplaceAll(utilSource, "twice", "doubled"),
		"main.rs": strings.ReplaceAll(mainSource, "twice", "doubled"),
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

func TestRealRustAnalyzerFindsReferencesAcrossFiles(t *testing.T) {
	s, util := openRefactorCrate(t)
	declaration := domain.SourceLocation{File: util, Line: 1, Column: 9}
	refs := eventually(t, func() ([]domain.Reference, bool) {
		refs, _ := s.server.References(context.Background(), declaration)
		return refs, len(refs) >= 3
	})
	if len(refs) != 3 {
		t.Fatalf("references = %+v", refs)
	}
	if !strings.HasSuffix(refs[0].Range.Start.File, "main.rs") || refs[0].Range.Start.Line != 4 {
		t.Errorf("first reference = %+v", refs[0])
	}
	if !strings.HasPrefix(refs[2].Preview, "pub fn twice") {
		t.Errorf("declaration preview = %q", refs[2].Preview)
	}
}
