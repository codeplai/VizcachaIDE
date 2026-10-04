package lldb

// Options is what a language adapter decides for one debugging session.
type Options struct {
	// Program is the executable to debug, already compiled with debug information.
	Program string
	// Dir is the working directory of the program; "" means the folder of Program.
	Dir string
	// Roots are the folders with the user's code. Frames whose source is outside them (the
	// standard library, the C runtime, ntdll) are hidden at the top of a crash stack.
	// Without roots every frame with a source file is kept.
	Roots []string
	// RunInTerminal asks lldb-dap to start the program through the IDE (runInTerminal request),
	// which gives it a keyboard. False: lldb-dap runs it itself and the output arrives as events.
	RunInTerminal bool
	// InitCommands are LLDB commands run before the program starts.
	InitCommands []string
	// CatchThrow also stops where a C++ exception is thrown (the cpp_throw filter). Off by
	// default: an exception nobody catches ends in abort(), which already stops the program.
	CatchThrow bool
}
