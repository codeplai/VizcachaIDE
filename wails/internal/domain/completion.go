package domain

// CompletionKind classifies a CompletionItem.
type CompletionKind string

// CompletionKind values.
const (
	CompletionKeyword  CompletionKind = "keyword"
	CompletionFunction CompletionKind = "function"
	CompletionMethod   CompletionKind = "method"
	CompletionVariable CompletionKind = "variable"
	CompletionConstant CompletionKind = "constant"
	CompletionField    CompletionKind = "field"
	CompletionType     CompletionKind = "type"
	CompletionPackage  CompletionKind = "package"
	CompletionOther    CompletionKind = "other"
)

// CompletionItem is one code suggestion.
type CompletionItem struct {
	Label         string         `json:"label"`
	Kind          CompletionKind `json:"kind"`
	Detail        string         `json:"detail"`
	Documentation string         `json:"documentation"`
	InsertText    string         `json:"insertText"`
}

// TextToInsert returns InsertText, or Label when there is no specific text.
func (c CompletionItem) TextToInsert() string {
	if c.InsertText != "" {
		return c.InsertText
	}
	return c.Label
}

// SignatureHelp is the call tip of a function being called.
type SignatureHelp struct {
	Label           string   `json:"label"`
	Documentation   string   `json:"documentation"`
	Parameters      []string `json:"parameters"`
	ActiveParameter int      `json:"activeParameter"`
}
