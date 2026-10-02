package domain

// Breakpoint is a line where the program must pause, with an optional condition.
type Breakpoint struct {
	Location  SourceLocation `json:"location"`
	Condition string         `json:"condition"`
}

// Variable is a variable value.
//
// Children are either already loaded (Children) or available lazily through
// Reference (non-zero means "ask the debugger for the children").
// Changed marks a variable whose value differs from the previous stop.
type Variable struct {
	Name      string     `json:"name"`
	TypeName  string     `json:"typeName"`
	Value     string     `json:"value"`
	Reference int        `json:"reference"`
	Changed   bool       `json:"changed"`
	Children  []Variable `json:"children"`
}

// HasChildren reports whether the variable can be expanded.
func (v Variable) HasChildren() bool {
	return v.Reference != 0 || len(v.Children) > 0
}

// StackFrame is one entry of the call stack.
type StackFrame struct {
	FrameID  int             `json:"frameId"`
	Function string          `json:"function"`
	Location *SourceLocation `json:"location"`
}

// FrameVariables is what one stack frame holds: the arguments it received and its
// local variables. The Calls view asks for it frame by frame.
type FrameVariables struct {
	Arguments []Variable `json:"arguments"`
	Locals    []Variable `json:"locals"`
}

// Goroutine is one goroutine of the debugged program.
type Goroutine struct {
	GoroutineID int             `json:"goroutineId"`
	Name        string          `json:"name"`
	Location    *SourceLocation `json:"location"`
}

// StopReason says why the program paused.
type StopReason string

// StopReason values.
const (
	StopEntry      StopReason = "entry"
	StopBreakpoint StopReason = "breakpoint"
	StopStep       StopReason = "step"
	StopPause      StopReason = "pause"
	StopPanic      StopReason = "panic"
)

// TerminatedByUser is the exit code of debug:terminated when the user stopped
// the session. Real exit codes are never negative.
const TerminatedByUser = -1

// DebugState is the snapshot taken every time the program stops.
type DebugState struct {
	Reason           StopReason   `json:"reason"`
	Frames           []StackFrame `json:"frames"`
	Variables        []Variable   `json:"variables"`
	Goroutines       []Goroutine  `json:"goroutines"`
	CurrentGoroutine *int         `json:"currentGoroutine"`
	Description      string       `json:"description"`
}

// CurrentLocation returns where the top frame is, or nil when unknown.
func (s DebugState) CurrentLocation() *SourceLocation {
	if len(s.Frames) == 0 {
		return nil
	}
	return s.Frames[0].Location
}
