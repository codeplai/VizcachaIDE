package app

import (
	"errors"
	"os"
	"path/filepath"
	"testing"

	"github.com/codeplai/VizcachaIDE/wails/internal/domain"
)

func writeTree(t *testing.T, files map[string]string) string {
	t.Helper()
	root := t.TempDir()
	for name, text := range files {
		path := filepath.Join(root, name)
		if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(path, []byte(text), 0o644); err != nil {
			t.Fatal(err)
		}
	}
	return root
}

func readBack(t *testing.T, path string) string {
	t.Helper()
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	return string(data)
}

func replaceOne(t *testing.T, text, query, replacement string, options domain.SearchOptions) (string, int) {
	t.Helper()
	root := writeTree(t, map[string]string{"a.txt": text})
	path := filepath.Join(root, "a.txt")
	result, err := ReplaceInFolder(ReplaceRequest{Root: root, Query: query, Options: options, Replacement: replacement, Paths: []string{path}}, nil)
	if err != nil {
		t.Fatal(err)
	}
	return readBack(t, path), result.Matches
}

func TestReplaceLiteralCaseOptions(t *testing.T) {
	got, n := replaceOne(t, "Foo foo FOO\n", "foo", "bar", domain.SearchOptions{})
	if got != "bar bar bar\n" || n != 3 {
		t.Errorf("insensitive: %q, %d", got, n)
	}
	got, n = replaceOne(t, "Foo foo FOO\n", "foo", "bar", domain.SearchOptions{CaseSensitive: true})
	if got != "Foo bar FOO\n" || n != 1 {
		t.Errorf("sensitive: %q, %d", got, n)
	}
}

func TestReplaceLiteralDoesNotExpandDollars(t *testing.T) {
	got, _ := replaceOne(t, "a.b\n", "a.b", "$1", domain.SearchOptions{})
	if got != "$1\n" {
		t.Errorf("got %q", got)
	}
}

func TestReplaceWholeWordWithAccents(t *testing.T) {
	got, n := replaceOne(t, "año años ñoño año_x año\n", "año", "dia", domain.SearchOptions{WholeWord: true})
	if got != "dia años ñoño año_x dia\n" || n != 2 {
		t.Errorf("got %q, %d", got, n)
	}
}

func TestReplaceRegexGroups(t *testing.T) {
	options := domain.SearchOptions{Regex: true, CaseSensitive: true}
	got, n := replaceOne(t, "x = 1\ny = 22\n", `(\w) = (\d+)`, "$2 <- $1 [$&] $$ $3", options)
	if got != "1 <- x [x = 1] $ $3\n22 <- y [y = 22] $ $3\n" || n != 2 {
		t.Errorf("got %q, %d", got, n)
	}
}

func TestReplaceInvalidRegex(t *testing.T) {
	root := writeTree(t, map[string]string{"a.txt": "x"})
	_, err := ReplaceInFolder(ReplaceRequest{Root: root, Query: "(", Options: domain.SearchOptions{Regex: true}, Paths: []string{filepath.Join(root, "a.txt")}}, nil)
	if !errors.Is(err, ErrInvalidPattern) {
		t.Errorf("got %v", err)
	}
}

func TestReplaceKeepsCRLFAndBOM(t *testing.T) {
	got, _ := replaceOne(t, "Feffone\r\ntwo one\r\n", "one$", "1", domain.SearchOptions{Regex: true})
	if got != "Feff1\r\ntwo 1\r\n" {
		t.Errorf("$ must match before the carriage return: %q", got)
	}
}

func TestReplaceOnlyTheOccurrenceAtAPlace(t *testing.T) {
	root := writeTree(t, map[string]string{"a.txt": "é cat cat\ncat\n"})
	path := filepath.Join(root, "a.txt")
	result, err := ReplaceInFolder(ReplaceRequest{Root: root, Query: "cat", Replacement: "dog", Paths: []string{path}, Line: 1, Column: 7}, nil)
	if err != nil || result.Matches != 1 {
		t.Fatalf("%v %+v", err, result)
	}
	if got := readBack(t, path); got != "é cat dog\ncat\n" {
		t.Errorf("got %q", got)
	}
	result, _ = ReplaceInFolder(ReplaceRequest{Root: root, Query: "cat", Replacement: "dog", Paths: []string{path}, Line: 2, Column: 5}, nil)
	if result.Matches != 0 {
		t.Errorf("a place with no match must change nothing: %+v", result)
	}
}

func TestReplaceRefusesWhatIsOutsideOrSkipped(t *testing.T) {
	root := writeTree(t, map[string]string{"src/a.txt": "cat", "node_modules/p/b.txt": "cat", "img.png": "cat"})
	outside := writeTree(t, map[string]string{"c.txt": "cat"})
	binary := filepath.Join(root, "bin.dat")
	if err := os.WriteFile(binary, []byte("cat\x00cat"), 0o644); err != nil {
		t.Fatal(err)
	}
	paths := []string{
		filepath.Join(root, "node_modules", "p", "b.txt"), filepath.Join(root, "img.png"),
		filepath.Join(outside, "c.txt"), binary, filepath.Join(root, "..", filepath.Base(outside), "c.txt"),
	}
	result, err := ReplaceInFolder(ReplaceRequest{Root: root, Query: "cat", Replacement: "dog", Paths: paths}, nil)
	if err != nil || result.Files != 0 {
		t.Fatalf("%v %+v", err, result)
	}
	for _, path := range paths[:4] {
		if readBack(t, path) == "dog" || readBack(t, path) == "dog\x00dog" {
			t.Errorf("%s was changed", path)
		}
	}
}

func TestReplaceRemembersWhatItWrote(t *testing.T) {
	root := writeTree(t, map[string]string{"a.txt": "cat"})
	path := filepath.Join(root, "a.txt")
	remembered := map[string]string{}
	_, err := ReplaceInFolder(ReplaceRequest{Root: root, Query: "cat", Replacement: "dog", Paths: []string{path}},
		func(p, text string) { remembered[p] = text })
	if err != nil || remembered[path] != "dog" {
		t.Errorf("%v %v", err, remembered)
	}
}
