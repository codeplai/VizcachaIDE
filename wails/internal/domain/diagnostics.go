// Package domain holds the pure model of VizcachaIDE: plain structs and rules.
// It depends on nothing else in the project, and every other layer depends on it.
package domain

// Severity ranks how serious a Diagnostic is.
type Severity string

// Severity values, in the same order as the 1.0 enum.
const (
	SeverityError   Severity = "error"
	SeverityWarning Severity = "warning"
	SeverityInfo    Severity = "info"
	SeverityHint    Severity = "hint"
)

// SourceLocation is a position in a source file. Line and Column are 1-based.
type SourceLocation struct {
	File   string `json:"file"`
	Line   int    `json:"line"`
	Column int    `json:"column"`
}

// Diagnostic is one problem reported by a Go tool.
//
// Message is the tool's message and RawText the exact text the tool printed.
// Both stay untranslated, so people can search for them on the web.
// End is where the problem ends (exclusive), when the tool reports a range.
type Diagnostic struct {
	Location *SourceLocation `json:"location"`
	Severity Severity        `json:"severity"`
	Message  string          `json:"message"`
	RawText  string          `json:"rawText"`
	Source   string          `json:"source"`
	Code     string          `json:"code"`
	End      *SourceLocation `json:"end"`
}
