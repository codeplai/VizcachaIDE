package domain

// ConsoleResult is what the interactive console answers to one snippet.
//
// Result is the value of an expression formatted like a REPL (strings quoted), or empty
// for statements. Output is what the snippet printed. Error is a plain message, empty
// when the snippet ran fine.
type ConsoleResult struct {
	Result string `json:"result"`
	Output string `json:"output"`
	Error  string `json:"error"`
}
