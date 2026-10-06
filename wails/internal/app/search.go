package app

import (
	"strings"

	"github.com/codeplai/VizcachaIDE/wails/internal/domain"
)

const (
	maxSearchMatches = 2000
	maxMatchTextRune = 300
)

// SearchFolder finds the occurrences of query in the text files under root.
func SearchFolder(root, query string, options domain.SearchOptions) (domain.SearchResult, error) {
	result := domain.SearchResult{Files: []domain.FileMatches{}}
	if query == "" {
		return result, nil
	}
	m, err := newMatcher(query, options)
	if err != nil {
		return result, err
	}
	total := 0
	err = walkTextFiles(root, func(file textFile) bool {
		matches := matchesOf(m, file.text)
		if len(matches) == 0 {
			return true
		}
		if total+len(matches) > maxSearchMatches {
			matches = matches[:maxSearchMatches-total]
			result.Truncated = true
		}
		total += len(matches)
		result.Files = append(result.Files, domain.FileMatches{Path: file.path, Matches: matches})
		return total < maxSearchMatches
	})
	return result, err
}

func matchesOf(m *matcher, text string) []domain.SearchMatch {
	var matches []domain.SearchMatch
	for i, line := range strings.Split(text, "\n") {
		body := strings.TrimSuffix(line, "\r")
		for _, index := range m.find(body) {
			matches = append(matches, domain.SearchMatch{
				Line:   i + 1,
				Column: utf16Length(body[:index[0]]) + 1,
				Length: utf16Length(body[index[0]:index[1]]),
				Text:   cutLine(body),
			})
		}
	}
	return matches
}

func cutLine(line string) string {
	runes := []rune(line)
	if len(runes) <= maxMatchTextRune {
		return line
	}
	return string(runes[:maxMatchTextRune])
}
