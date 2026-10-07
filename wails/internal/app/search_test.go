package app

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/codeplai/VizcachaIDE/wails/internal/domain"
)

func TestSearchFindsMatchesWithUTF16Columns(t *testing.T) {
	root := writeTree(t, map[string]string{"a.go": "😀 ñandú\nnada\n  ñandú ñandú\n"})
	result, err := SearchFolder(root, "ñandú", domain.SearchOptions{})
	if err != nil || len(result.Files) != 1 {
		t.Fatalf("%v %+v", err, result)
	}
	got := result.Files[0].Matches
	if len(got) != 3 || got[0].Line != 1 || got[0].Column != 4 || got[0].Length != 5 {
		t.Fatalf("matches: %+v", got)
	}
	if got[1].Line != 3 || got[1].Column != 3 || got[2].Column != 9 {
		t.Errorf("matches: %+v", got)
	}
}

func TestSearchSkipsFoldersAndBinaryFiles(t *testing.T) {
	root := writeTree(t, map[string]string{
		"main.go": "needle", "build/o.txt": "needle", "target/x.rs": "needle",
		"node_modules/m/i.js": "needle", ".git/config": "needle", "logo.png": "needle", "sub/b.py": "needle",
	})
	if err := os.WriteFile(filepath.Join(root, "raw.dat"), []byte("needle\x00"), 0o644); err != nil {
		t.Fatal(err)
	}
	result, err := SearchFolder(root, "needle", domain.SearchOptions{})
	if err != nil {
		t.Fatal(err)
	}
	var names []string
	for _, file := range result.Files {
		rel, _ := filepath.Rel(root, file.Path)
		names = append(names, filepath.ToSlash(rel))
	}
	if strings.Join(names, ",") != "main.go,sub/b.py" {
		t.Errorf("searched %v", names)
	}
}

func TestSearchEmptyQueryAndInvalidRegex(t *testing.T) {
	root := writeTree(t, map[string]string{"a.txt": "x"})
	if result, err := SearchFolder(root, "", domain.SearchOptions{}); err != nil || len(result.Files) != 0 {
		t.Errorf("%v %+v", err, result)
	}
	if _, err := SearchFolder(root, "[", domain.SearchOptions{Regex: true}); err == nil {
		t.Error("an invalid regex must fail")
	}
}

func TestSearchStopsAtTheLimit(t *testing.T) {
	root := writeTree(t, map[string]string{"a.txt": strings.Repeat("x\n", maxSearchMatches+50)})
	result, err := SearchFolder(root, "x", domain.SearchOptions{})
	if err != nil || !result.Truncated || len(result.Files[0].Matches) != maxSearchMatches {
		t.Errorf("%v truncated=%v", err, result.Truncated)
	}
}
