package domain

import (
	"errors"
	"testing"
)

func edit(line, col, endLine, endCol int, text string) TextEdit {
	return TextEdit{
		Range:   SourceRange{Start: SourceLocation{Line: line, Column: col}, End: SourceLocation{Line: endLine, Column: endCol}},
		NewText: text,
	}
}

func TestApplyTextEditsRenamesEveryOccurrence(t *testing.T) {
	text := "func greet() {}\ngreet()\ngreet()\n"
	got, err := ApplyTextEdits(text, []TextEdit{edit(3, 1, 3, 6, "hello"), edit(1, 6, 1, 11, "hello"), edit(2, 1, 2, 6, "hello")})
	if err != nil {
		t.Fatal(err)
	}
	if want := "func hello() {}\nhello()\nhello()\n"; got != want {
		t.Fatalf("got %q, want %q", got, want)
	}
}

func TestApplyTextEditsCountsRunesNotBytes(t *testing.T) {
	// "año" has a two-byte rune; the emoji is one rune (two UTF-16 units, the caller converts).
	text := "año := 1\nprintln(año, \"😀\", año)\n"
	got, err := ApplyTextEdits(text, []TextEdit{edit(1, 1, 1, 4, "x"), edit(2, 9, 2, 12, "x"), edit(2, 19, 2, 22, "x")})
	if err != nil {
		t.Fatal(err)
	}
	if want := "x := 1\nprintln(x, \"😀\", x)\n"; got != want {
		t.Fatalf("got %q, want %q", got, want)
	}
}

func TestApplyTextEditsKeepsCRLFAndAllowsEndOfLine(t *testing.T) {
	got, err := ApplyTextEdits("a\r\nb\r\n", []TextEdit{edit(1, 2, 1, 2, "!"), edit(2, 2, 2, 2, "?")})
	if err != nil {
		t.Fatal(err)
	}
	if want := "a!\r\nb?\r\n"; got != want {
		t.Fatalf("got %q, want %q", got, want)
	}
}

func TestApplyTextEditsRejectsOverlapAndOutOfRange(t *testing.T) {
	if _, err := ApplyTextEdits("abcdef\n", []TextEdit{edit(1, 1, 1, 4, "x"), edit(1, 3, 1, 5, "y")}); !errors.Is(err, ErrOverlappingEdits) {
		t.Fatalf("overlap: %v", err)
	}
	if _, err := ApplyTextEdits("abc\n", []TextEdit{edit(1, 9, 1, 10, "x")}); !errors.Is(err, ErrEditOutOfRange) {
		t.Fatalf("column: %v", err)
	}
	if _, err := ApplyTextEdits("abc\n", []TextEdit{edit(5, 1, 5, 2, "x")}); !errors.Is(err, ErrEditOutOfRange) {
		t.Fatalf("line: %v", err)
	}
}

func TestApplyTextEditsKeepsInsertOrderAtTheSamePoint(t *testing.T) {
	got, err := ApplyTextEdits("ab", []TextEdit{edit(1, 2, 1, 2, "1"), edit(1, 2, 1, 2, "2")})
	if err != nil || got != "a12b" {
		t.Fatalf("got %q, %v", got, err)
	}
}
