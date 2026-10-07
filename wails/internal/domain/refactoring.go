package domain

// RenameRefusal says why a symbol cannot be renamed. The frontend translates it.
type RenameRefusal string

const (
	// RenameAllowed means nothing stops the rename.
	RenameAllowed RenameRefusal = ""
	// RenameNotRenameable means the position holds no symbol that can be renamed (a keyword,
	// a library symbol, a comment).
	RenameNotRenameable RenameRefusal = "notRenameable"
	// RenameUnsupported means the language server cannot rename (no rename support, or the
	// optional refactoring library of the Python server is missing).
	RenameUnsupported RenameRefusal = "unsupported"
	// RenameFailed means the server tried and failed; Detail has its message.
	RenameFailed RenameRefusal = "failed"
)

// RenameTarget is the answer of textDocument/prepareRename.
type RenameTarget struct {
	Refusal RenameRefusal `json:"refusal"`
	// Range is the symbol to rename; nil when the server did not say (the editor then uses the
	// word under the cursor).
	Range *SourceRange `json:"range,omitempty"`
	// Placeholder is the current name; empty when the server did not say.
	Placeholder string `json:"placeholder"`
	Detail      string `json:"detail,omitempty"`
}

// TextEdit replaces a range (columns in runes, End exclusive) with NewText.
type TextEdit struct {
	Range   SourceRange `json:"range"`
	NewText string      `json:"newText"`
}

// FileEdit is every edit of one file.
type FileEdit struct {
	File  string     `json:"file"`
	Edits []TextEdit `json:"edits"`
}

// RenameResult is the answer of textDocument/rename: the edits to apply to every file, or why
// there are none.
type RenameResult struct {
	Refusal RenameRefusal `json:"refusal"`
	Detail  string        `json:"detail,omitempty"`
	Files   []FileEdit    `json:"files"`
}

// Reference is one use of a symbol, with the text of its line.
type Reference struct {
	Range   SourceRange `json:"range"`
	Preview string      `json:"preview"`
}

// EditSummary tells what applying edits to files on disk did.
type EditSummary struct {
	Files int `json:"files"`
	Edits int `json:"edits"`
}
