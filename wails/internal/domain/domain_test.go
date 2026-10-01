package domain

import (
	"encoding/json"
	"strings"
	"testing"
)

func TestNewFileRunConfiguration(t *testing.T) {
	config := NewFileRunConfiguration("proj/main.go", []string{"-v"})
	if config.Mode != RunFile {
		t.Fatalf("mode = %q, want %q", config.Mode, RunFile)
	}
	if got := config.GoTargetArgument(); got != "main.go" {
		t.Errorf("GoTargetArgument() = %q, want main.go", got)
	}
	if got := config.ExecutableName(true); got != "main.exe" {
		t.Errorf("ExecutableName(true) = %q, want main.exe", got)
	}
	if got := config.ExecutableName(false); got != "main" {
		t.Errorf("ExecutableName(false) = %q, want main", got)
	}
}

func TestPackageRunConfiguration(t *testing.T) {
	config := RunConfiguration{Target: "proj/main.go", WorkingDir: "proj", Mode: RunPackage}
	if got := config.GoTargetArgument(); got != "." {
		t.Errorf("GoTargetArgument() = %q, want .", got)
	}
	if got := config.ExecutableName(false); got != "proj" {
		t.Errorf("ExecutableName(false) = %q, want proj", got)
	}
}

func TestSourceRangeIsEmpty(t *testing.T) {
	at := func(line, column int) SourceLocation { return SourceLocation{Line: line, Column: column} }
	cases := []struct {
		name  string
		r     SourceRange
		empty bool
	}{
		{"same point", SourceRange{at(1, 3), at(1, 3)}, true},
		{"one character", SourceRange{at(1, 3), at(1, 4)}, false},
		{"next line", SourceRange{at(1, 3), at(2, 1)}, false},
		{"reversed", SourceRange{at(2, 1), at(1, 9)}, true},
	}
	for _, tc := range cases {
		if got := tc.r.IsEmpty(); got != tc.empty {
			t.Errorf("%s: IsEmpty() = %v, want %v", tc.name, got, tc.empty)
		}
	}
}

func TestVariableAndStateHelpers(t *testing.T) {
	if (Variable{Name: "x"}).HasChildren() {
		t.Error("a plain variable has no children")
	}
	if !(Variable{Reference: 7}).HasChildren() {
		t.Error("a variable with a reference has children")
	}
	loc := SourceLocation{File: "main.go", Line: 6, Column: 1}
	state := DebugState{Frames: []StackFrame{{Function: "sumar", Location: &loc}}}
	if got := state.CurrentLocation(); got == nil || got.Line != 6 {
		t.Errorf("CurrentLocation() = %v, want line 6", got)
	}
	if (DebugState{}).CurrentLocation() != nil {
		t.Error("an empty state has no location")
	}
}

func TestCompletionTextToInsert(t *testing.T) {
	if got := (CompletionItem{Label: "Println"}).TextToInsert(); got != "Println" {
		t.Errorf("got %q, want the label", got)
	}
	if got := (CompletionItem{Label: "Println", InsertText: "Println($0)"}).TextToInsert(); got != "Println($0)" {
		t.Errorf("got %q, want the insert text", got)
	}
}

func TestJSONUsesCamelCase(t *testing.T) {
	data, err := json.Marshal(DebugState{Reason: StopBreakpoint, Variables: []Variable{{TypeName: "int"}}})
	if err != nil {
		t.Fatal(err)
	}
	text := string(data)
	for _, want := range []string{`"reason"`, `"typeName"`, `"currentGoroutine"`} {
		if !strings.Contains(text, want) {
			t.Errorf("json %s lacks %s", text, want)
		}
	}
	if strings.Contains(text, "_") {
		t.Errorf("json %s has snake_case keys", text)
	}
}
