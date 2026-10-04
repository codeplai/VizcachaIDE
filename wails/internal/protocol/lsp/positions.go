package lsp

import (
	"path/filepath"
	"runtime"
	"strings"
	"unicode/utf16"
	"unicode/utf8"

	"go.lsp.dev/protocol"
	"go.lsp.dev/uri"

	"github.com/codeplai/VizcachaIDE/wails/internal/domain"
)

// pathToURI turns a file path into a file:// URI.
func pathToURI(path string) protocol.DocumentURI {
	abs, err := filepath.Abs(path)
	if err != nil {
		abs = path
	}
	return uri.File(abs)
}

// uriToPath is the inverse of pathToURI.
func uriToPath(documentURI protocol.DocumentURI) string {
	return uri.URI(documentURI).Filename()
}

// pathKey normalises a path so the same file always maps to the same key
// (Windows paths are case-insensitive).
func pathKey(path string) string {
	abs, err := filepath.Abs(path)
	if err != nil {
		abs = path
	}
	abs = filepath.Clean(abs)
	if runtime.GOOS == "windows" {
		return strings.ToLower(abs)
	}
	return abs
}

// lineOf returns line number lineIndex (0-based) of text without its line ending.
func lineOf(text string, lineIndex int) string {
	if lineIndex < 0 {
		return ""
	}
	for i := 0; i < lineIndex; i++ {
		next := strings.IndexByte(text, '\n')
		if next < 0 {
			return ""
		}
		text = text[next+1:]
	}
	if end := strings.IndexByte(text, '\n'); end >= 0 {
		text = text[:end]
	}
	return strings.TrimSuffix(text, "\r")
}

// utf16Offset counts the UTF-16 code units before the rune index runeIndex.
func utf16Offset(line string, runeIndex int) int {
	units := 0
	for _, r := range line {
		if runeIndex <= 0 {
			break
		}
		runeIndex--
		units += utf16Units(r)
	}
	return units
}

// runeIndex is the inverse of utf16Offset (positions inside a surrogate pair round down).
func runeIndex(line string, character int) int {
	units, index := 0, 0
	for _, r := range line {
		units += utf16Units(r)
		if units > character {
			return index
		}
		index++
	}
	return utf8.RuneCountInString(line)
}

func utf16Units(r rune) int {
	if utf16.IsSurrogate(r) || r > 0xFFFF {
		return 2
	}
	return 1
}

// toLSPPosition converts a 1-based (line, column in runes) position into an LSP one.
func toLSPPosition(text string, line, column int) protocol.Position {
	lineIndex := max(line-1, 0)
	character := utf16Offset(lineOf(text, lineIndex), max(column-1, 0))
	return protocol.Position{Line: uint32(lineIndex), Character: uint32(character)}
}

// fromLSPPosition converts an LSP position into 1-based line and column (in runes).
func fromLSPPosition(text string, position protocol.Position) (line, column int) {
	column = runeIndex(lineOf(text, int(position.Line)), int(position.Character))
	return int(position.Line) + 1, column + 1
}

// locationIn builds the domain location of an LSP position in the text of file.
func locationIn(text, file string, position protocol.Position) domain.SourceLocation {
	line, column := fromLSPPosition(text, position)
	return domain.SourceLocation{File: file, Line: line, Column: column}
}

// rangeIn builds the domain range of an LSP range in the text of file.
func rangeIn(text, file string, lspRange protocol.Range) domain.SourceRange {
	return domain.SourceRange{
		Start: locationIn(text, file, lspRange.Start),
		End:   locationIn(text, file, lspRange.End),
	}
}
