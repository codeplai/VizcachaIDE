package domain

// ErrorExplanation is a beginner-friendly explanation of a recognised Go error.
//
// ExplanationID is stable across versions and languages (for example
// "E-UNUSED-VAR"), so it can be used in tests, help links and bug reports.
// Title, Body and FixHint are already translated and have their placeholders
// filled in when they reach the frontend.
type ErrorExplanation struct {
	ExplanationID string            `json:"explanationId"`
	Title         string            `json:"title"`
	Body          string            `json:"body"`
	FixHint       string            `json:"fixHint"`
	Placeholders  map[string]string `json:"placeholders"`
}

// ExplainedDiagnostic pairs a Diagnostic with its explanation, when there is one.
type ExplainedDiagnostic struct {
	Diagnostic  Diagnostic        `json:"diagnostic"`
	Explanation *ErrorExplanation `json:"explanation"`
}
