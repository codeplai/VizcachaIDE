package domain

// SearchOptions are the switches of the Search panel.
type SearchOptions struct {
	CaseSensitive bool `json:"caseSensitive"`
	WholeWord     bool `json:"wholeWord"`
	Regex         bool `json:"regex"`
}

// SearchMatch is one occurrence found in a line. Line and Column are 1-based and Column and
// Length count UTF-16 units, like the editor does. Text is the line (cut when very long).
type SearchMatch struct {
	Line   int    `json:"line"`
	Column int    `json:"column"`
	Length int    `json:"length"`
	Text   string `json:"text"`
}

// FileMatches are the occurrences found in one file.
type FileMatches struct {
	Path    string        `json:"path"`
	Matches []SearchMatch `json:"matches"`
}

// SearchResult is what a search in the open folder found.
type SearchResult struct {
	Files []FileMatches `json:"files"`
	// Truncated is true when the search stopped at its limit of matches.
	Truncated bool `json:"truncated"`
}

// ReplaceResult says how much a replacement in files changed.
type ReplaceResult struct {
	Files   int `json:"files"`
	Matches int `json:"matches"`
}
