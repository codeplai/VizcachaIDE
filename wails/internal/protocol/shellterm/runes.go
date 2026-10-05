package shellterm

import "unicode/utf8"

// completeLength is how many leading bytes of data end on a character boundary; the rest is the
// start of a rune that the next read finishes. Bytes that are not UTF-8 at all count as complete.
func completeLength(data []byte) int {
	for back := 1; back <= utf8.UTFMax-1 && back <= len(data); back++ {
		start := len(data) - back
		if !utf8.RuneStart(data[start]) {
			continue
		}
		if !utf8.FullRune(data[start:]) {
			return start
		}
		return len(data)
	}
	return len(data)
}
