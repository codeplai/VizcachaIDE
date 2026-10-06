package domain

import (
	"errors"
	"sort"
	"strings"
)

// ErrOverlappingEdits means two edits touch the same text, so they cannot be applied safely.
var ErrOverlappingEdits = errors.New("overlapping text edits")

// ErrEditOutOfRange means an edit points outside the text.
var ErrEditOutOfRange = errors.New("text edit outside the text")

// lineStarts returns the rune offset where each line of the text starts.
func lineStarts(runes []rune) []int {
	starts := []int{0}
	for i, r := range runes {
		if r == '\n' {
			starts = append(starts, i+1)
		}
	}
	return starts
}

// offsetOf is the rune offset of a 1-based line and rune column. A column may point at the end
// of its line (before the line break) but not beyond it.
func offsetOf(runes []rune, starts []int, line, column int) (int, error) {
	if line < 1 || line > len(starts) || column < 1 {
		return 0, ErrEditOutOfRange
	}
	offset := starts[line-1] + column - 1
	end := len(runes)
	if line < len(starts) {
		end = starts[line] - 1 // the "\n"
	}
	if offset > end {
		return 0, ErrEditOutOfRange
	}
	return offset, nil
}

type spanEdit struct {
	from, to int
	text     string
}

func toSpans(runes []rune, edits []TextEdit) ([]spanEdit, error) {
	starts := lineStarts(runes)
	spans := make([]spanEdit, 0, len(edits))
	for _, edit := range edits {
		from, err := offsetOf(runes, starts, edit.Range.Start.Line, edit.Range.Start.Column)
		if err != nil {
			return nil, err
		}
		to, err := offsetOf(runes, starts, edit.Range.End.Line, edit.Range.End.Column)
		if err != nil || to < from {
			return nil, ErrEditOutOfRange
		}
		spans = append(spans, spanEdit{from, to, edit.NewText})
	}
	sort.SliceStable(spans, func(i, j int) bool { return spans[i].from < spans[j].from })
	return spans, nil
}

// ApplyTextEdits applies edits (positions in runes, as the language service reports them) to
// text. Edits must not overlap; several inserts at the same point keep their order.
func ApplyTextEdits(text string, edits []TextEdit) (string, error) {
	runes := []rune(text)
	spans, err := toSpans(runes, edits)
	if err != nil {
		return "", err
	}
	var out strings.Builder
	cursor := 0
	for _, span := range spans {
		if span.from < cursor {
			return "", ErrOverlappingEdits
		}
		out.WriteString(string(runes[cursor:span.from]))
		out.WriteString(span.text)
		cursor = span.to
	}
	out.WriteString(string(runes[cursor:]))
	return out.String(), nil
}
