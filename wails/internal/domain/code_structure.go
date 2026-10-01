package domain

// SourceRange goes from Start (inclusive) to End (exclusive), both 1-based.
type SourceRange struct {
	Start SourceLocation `json:"start"`
	End   SourceLocation `json:"end"`
}

// IsEmpty reports whether the range covers no text.
func (r SourceRange) IsEmpty() bool {
	if r.Start.Line != r.End.Line {
		return r.Start.Line > r.End.Line
	}
	return r.Start.Column >= r.End.Column
}

// SymbolKind classifies a DocumentSymbol.
type SymbolKind string

// SymbolKind values.
const (
	SymbolFunction  SymbolKind = "function"
	SymbolMethod    SymbolKind = "method"
	SymbolStruct    SymbolKind = "struct"
	SymbolInterface SymbolKind = "interface"
	SymbolType      SymbolKind = "type"
	SymbolVariable  SymbolKind = "variable"
	SymbolConstant  SymbolKind = "constant"
	SymbolField     SymbolKind = "field"
	SymbolPackage   SymbolKind = "package"
	SymbolOther     SymbolKind = "other"
)

// DocumentSymbol is one declaration of a file (for the Outline) with its nested declarations.
//
// Location is where the symbol's name is (the navigation target) and Range the
// whole declaration, when the language server reports it.
type DocumentSymbol struct {
	Name     string           `json:"name"`
	Kind     SymbolKind       `json:"kind"`
	Location SourceLocation   `json:"location"`
	Range    *SourceRange     `json:"range"`
	Detail   string           `json:"detail"`
	Children []DocumentSymbol `json:"children"`
}
