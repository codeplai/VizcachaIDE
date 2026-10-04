package domain

// InlayHintKind says what an inlay hint shows.
type InlayHintKind string

// InlayHintKind values.
const (
	InlayHintType      InlayHintKind = "type"      // the inferred type of a variable: `let x: Vec<i32>`
	InlayHintParameter InlayHintKind = "parameter" // the name of a parameter before an argument
	InlayHintOther     InlayHintKind = "other"
)

// InlayHint is a label the editor draws inside the code without changing it (docs/PLAN_RUST.md
// sections 3.2 and 9.1). Line and Column (1-based) are where it goes; Padding asks for a space
// before or after the label.
type InlayHint struct {
	Line         int           `json:"line"`
	Column       int           `json:"column"`
	Label        string        `json:"label"`
	Kind         InlayHintKind `json:"kind"`
	PaddingLeft  bool          `json:"paddingLeft"`
	PaddingRight bool          `json:"paddingRight"`
}
