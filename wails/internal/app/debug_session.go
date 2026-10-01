package app

import (
	"path/filepath"
	"slices"
	"sort"
	"strconv"
	"sync"

	"github.com/codeplai/VizcachaIDE/wails/internal/domain"
)

// BreakpointBook remembers the user's breakpoints per file, plus the temporary
// one used by "Run to here". It knows nothing about the debugger protocol.
type BreakpointBook struct {
	mu         sync.Mutex
	conditions map[string]map[int]string // file -> line -> condition
	temporary  *domain.SourceLocation
}

// NewBreakpointBook creates an empty book.
func NewBreakpointBook() *BreakpointBook {
	return &BreakpointBook{conditions: map[string]map[int]string{}}
}

// BookKey is the canonical spelling of a file path inside the book.
func BookKey(file string) string {
	if absolute, err := filepath.Abs(file); err == nil {
		file = absolute
	}
	return filepath.Clean(file)
}

// Reset replaces everything with the breakpoints of a new session.
func (b *BreakpointBook) Reset(breakpoints []domain.Breakpoint) {
	b.mu.Lock()
	defer b.mu.Unlock()
	b.conditions = map[string]map[int]string{}
	b.temporary = nil
	for _, point := range breakpoints {
		key := BookKey(point.Location.File)
		if b.conditions[key] == nil {
			b.conditions[key] = map[int]string{}
		}
		b.conditions[key][point.Location.Line] = point.Condition
	}
}

// Replace sets the lines of one file (keeping the conditions of lines that stay)
// and returns the file key.
func (b *BreakpointBook) Replace(file string, lines []int) string {
	b.mu.Lock()
	defer b.mu.Unlock()
	key := BookKey(file)
	updated := map[int]string{}
	for _, line := range lines {
		updated[line] = b.conditions[key][line]
	}
	b.conditions[key] = updated
	return key
}

// SetTemporary adds the "Run to here" breakpoint and returns its file key.
func (b *BreakpointBook) SetTemporary(location domain.SourceLocation) string {
	b.mu.Lock()
	defer b.mu.Unlock()
	b.temporary = &domain.SourceLocation{File: BookKey(location.File), Line: location.Line, Column: 1}
	return b.temporary.File
}

// ClearTemporary forgets the temporary breakpoint and returns the file whose
// breakpoints must be sent again, or "" when there was none.
func (b *BreakpointBook) ClearTemporary() string {
	b.mu.Lock()
	defer b.mu.Unlock()
	if b.temporary == nil {
		return ""
	}
	file := b.temporary.File
	b.temporary = nil
	return file
}

// Files lists every file that has breakpoints.
func (b *BreakpointBook) Files() []string {
	b.mu.Lock()
	defer b.mu.Unlock()
	files := make([]string, 0, len(b.conditions))
	for file := range b.conditions {
		files = append(files, file)
	}
	sort.Strings(files)
	return files
}

// For returns the breakpoints of one file ordered by line, temporary one included.
func (b *BreakpointBook) For(file string) []domain.Breakpoint {
	b.mu.Lock()
	defer b.mu.Unlock()
	key := BookKey(file)
	points := map[int]string{}
	for line, condition := range b.conditions[key] {
		points[line] = condition
	}
	if b.temporary != nil && b.temporary.File == key {
		points[b.temporary.Line] = ""
	}
	result := make([]domain.Breakpoint, 0, len(points))
	for line, condition := range points {
		location := domain.SourceLocation{File: key, Line: line, Column: 1}
		result = append(result, domain.Breakpoint{Location: location, Condition: condition})
	}
	slices.SortFunc(result, func(a, c domain.Breakpoint) int { return a.Location.Line - c.Location.Line })
	return result
}

// ChangeTracker marks the variables whose value changed since the previous stop.
//
// A variable is "the same" when it has the same name in the same frame; a frame is
// identified by the function and the depth of the stack, so a recursive call is
// compared with its own earlier visit and not with its caller.
type ChangeTracker struct {
	mu       sync.Mutex
	previous map[string]string
	frames   map[string]bool
}

// NewChangeTracker creates a tracker with no history.
func NewChangeTracker() *ChangeTracker { return &ChangeTracker{} }

// Reset forgets the history (a new session starts).
func (t *ChangeTracker) Reset() {
	t.mu.Lock()
	defer t.mu.Unlock()
	t.previous = nil
	t.frames = nil
}

// Mark sets Changed on the variables of the top frame and remembers their values.
// The first stop of a frame never marks anything. After that, a variable that was not there
// before (the one a line has just declared) counts as changed, like a new value does.
func (t *ChangeTracker) Mark(frames []domain.StackFrame, variables []domain.Variable) []domain.Variable {
	if len(frames) == 0 {
		return variables
	}
	t.mu.Lock()
	defer t.mu.Unlock()
	frameKey := frames[0].Function + "#" + strconv.Itoa(len(frames))
	frameKnown := t.frames[frameKey]
	current := make(map[string]string, len(variables))
	for index, variable := range variables {
		key := frameKey + "/" + variable.Name
		old, seen := t.previous[key]
		variables[index].Changed = (seen && old != variable.Value) || (!seen && frameKnown)
		current[key] = variable.Value
	}
	t.previous = mergeValues(t.previous, current)
	if t.frames == nil {
		t.frames = map[string]bool{}
	}
	t.frames[frameKey] = true
	return variables
}

// mergeValues keeps the history of other frames so returning to a caller still compares.
func mergeValues(history, current map[string]string) map[string]string {
	if history == nil {
		return current
	}
	for key, value := range current {
		history[key] = value
	}
	return history
}
