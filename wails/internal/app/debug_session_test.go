package app

import (
	"path/filepath"
	"testing"

	"github.com/codeplai/VizcachaIDE/wails/internal/domain"
)

func frame(function string) domain.StackFrame { return domain.StackFrame{Function: function} }

func values(pairs ...string) []domain.Variable {
	variables := make([]domain.Variable, 0, len(pairs)/2)
	for index := 0; index < len(pairs); index += 2 {
		variables = append(variables, domain.Variable{Name: pairs[index], Value: pairs[index+1]})
	}
	return variables
}

func changedNames(variables []domain.Variable) []string {
	names := []string{}
	for _, variable := range variables {
		if variable.Changed {
			names = append(names, variable.Name)
		}
	}
	return names
}

func TestChangeTrackerMarksOnlyValuesThatChanged(t *testing.T) {
	tracker := NewChangeTracker()
	stack := []domain.StackFrame{frame("main.multiply"), frame("main.main")}

	first := tracker.Mark(stack, values("a", "5", "result", "0"))
	second := tracker.Mark(stack, values("a", "5", "result", "35"))
	third := tracker.Mark(stack, values("a", "5", "result", "35"))

	if got := changedNames(first); len(got) != 0 {
		t.Errorf("first stop changed %v, want none", got)
	}
	if got := changedNames(second); len(got) != 1 || got[0] != "result" {
		t.Errorf("second stop changed %v, want [result]", got)
	}
	if got := changedNames(third); len(got) != 0 {
		t.Errorf("third stop changed %v, want none", got)
	}
}

// A line such as `product := multiply(5, 7)` declares a variable: when the frame is visited again
// the new variable is shown as just changed (found by the QA with functions.go).
func TestChangeTrackerMarksAVariableDeclaredAfterTheFirstStop(t *testing.T) {
	tracker := NewChangeTracker()
	stack := []domain.StackFrame{frame("main.main")}

	first := tracker.Mark(stack, values("sum", "30"))
	second := tracker.Mark(stack, values("sum", "30", "product", "35"))

	if got := changedNames(first); len(got) != 0 {
		t.Errorf("first stop of the frame changed %v, want none", got)
	}
	if got := changedNames(second); len(got) != 1 || got[0] != "product" {
		t.Errorf("second stop changed %v, want [product]", got)
	}
}

func TestChangeTrackerComparesRecursiveCallsWithTheirOwnDepth(t *testing.T) {
	tracker := NewChangeTracker()
	outer := []domain.StackFrame{frame("main.factorial"), frame("main.main")}
	inner := []domain.StackFrame{frame("main.factorial"), frame("main.factorial"), frame("main.main")}

	tracker.Mark(outer, values("n", "5"))
	entered := tracker.Mark(inner, values("n", "4"))

	if got := changedNames(entered); len(got) != 0 {
		t.Errorf("a new recursive frame is not a change, got %v", got)
	}
}

func TestChangeTrackerResetForgetsHistory(t *testing.T) {
	tracker := NewChangeTracker()
	stack := []domain.StackFrame{frame("main.main")}
	tracker.Mark(stack, values("x", "1"))
	tracker.Reset()

	if got := changedNames(tracker.Mark(stack, values("x", "2"))); len(got) != 0 {
		t.Errorf("after Reset nothing is a change, got %v", got)
	}
}

func TestBreakpointBookKeepsConditionsAndTheTemporaryBreakpoint(t *testing.T) {
	file := filepath.Join(t.TempDir(), "main.go")
	book := NewBreakpointBook()
	book.Reset([]domain.Breakpoint{{Location: domain.SourceLocation{File: file, Line: 13}, Condition: "a > 3"}})

	book.Replace(file, []int{13, 20})
	book.SetTemporary(domain.SourceLocation{File: file, Line: 30})
	points := book.For(file)

	if len(points) != 3 || points[0].Condition != "a > 3" || points[1].Location.Line != 20 || points[2].Location.Line != 30 {
		t.Fatalf("points = %+v", points)
	}
	if cleared := book.ClearTemporary(); cleared != BookKey(file) {
		t.Errorf("cleared = %q, want %q", cleared, BookKey(file))
	}
	if len(book.For(file)) != 2 || book.ClearTemporary() != "" {
		t.Error("the temporary breakpoint must be gone")
	}
}
