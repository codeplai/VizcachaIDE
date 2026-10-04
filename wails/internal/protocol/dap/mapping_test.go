package dap

import (
	"path/filepath"
	"strings"
	"testing"

	"github.com/codeplai/VizcachaIDE/wails/internal/domain"
	"github.com/google/go-dap"
)

func TestFramesFromRecordedStackTrace(t *testing.T) {
	dir := t.TempDir()
	response := responseFor[*dap.StackTraceResponse](t, functionsSession(t, dir))

	frames := mapFrames(response.Body.StackFrames, testFlavor{}, false)

	// runtime.main (a "subtle" frame below the last kept one) is dropped.
	names := make([]string, 0, len(frames))
	for _, frame := range frames {
		names = append(names, frame.Function)
	}
	if strings.Join(names, ",") != "main.multiply,main.main" {
		t.Fatalf("functions = %v", names)
	}
	if frames[0].FrameID != 1000 {
		t.Errorf("frame id = %d, want 1000", frames[0].FrameID)
	}
	want := domain.SourceLocation{File: filepath.Join(dir, "functions.go"), Line: 13, Column: 1}
	if *frames[0].Location != want {
		t.Errorf("location = %+v, want %+v", *frames[0].Location, want)
	}
}

func TestPanicSkipsLeadingRuntimeFrames(t *testing.T) {
	dir := t.TempDir()
	response := panicPart(t, dir, "stackTrace").(*dap.StackTraceResponse)

	frames := mapFrames(response.Body.StackFrames, testFlavor{}, true)

	// The leading panic frames and the trailing runtime.main are both hidden.
	if len(frames) != 1 || frames[0].Function != "main.main" {
		t.Fatalf("frames = %+v", frames)
	}
	if frames[0].Location.File != filepath.Join(dir, "panic.go") {
		t.Errorf("file = %s", frames[0].Location.File)
	}
	if kept := mapFrames(response.Body.StackFrames, testFlavor{}, false); len(kept) != 4 {
		t.Errorf("without skipping the leading ones there must be 4 frames, got %d", len(kept))
	}
}

func TestVariablesKeepReferenceForLazyExpansion(t *testing.T) {
	dir := t.TempDir()
	simple := mapVariables(responseFor[*dap.VariablesResponse](t, functionsSession(t, dir)).Body.Variables, testFlavor{})
	nested := mapVariables(panicPart(t, dir, "long_variable").(*dap.VariablesResponse).Body.Variables, testFlavor{})

	names := []string{simple[0].Name, simple[1].Name, simple[2].Name, simple[3].Name}
	if strings.Join(names, ",") != "a,b,~r0,result" {
		t.Fatalf("names = %v", names)
	}
	if simple[0].Value != "5" || simple[0].TypeName != "int" || simple[0].HasChildren() {
		t.Errorf("a = %+v", simple[0])
	}
	if nested[0].Reference != 1001 || !nested[0].HasChildren() {
		t.Errorf("msgs = %+v", nested[0])
	}
}

func TestLongValuesAreTruncated(t *testing.T) {
	raw := panicPart(t, t.TempDir(), "long_variable").(*dap.VariablesResponse).Body.Variables
	if len([]rune(raw[0].Value)) <= maxValueLength {
		t.Fatal("the fixture value must be longer than the limit")
	}

	value := mapVariables(raw, testFlavor{})[0].Value

	if len([]rune(value)) != maxValueLength || !strings.HasSuffix(value, ellipsis) {
		t.Errorf("value has %d runes and ends with %q", len([]rune(value)), value[len(value)-3:])
	}
}

func TestLocalsScopeEvenForOptimizedFunctions(t *testing.T) {
	dir := t.TempDir()
	normal := responseFor[*dap.ScopesResponse](t, functionsSession(t, dir))
	optimized := panicPart(t, dir, "optimized_scopes").(*dap.ScopesResponse)

	if localsReference(normal.Body.Scopes, testFlavor{}) != 1000 || localsReference(optimized.Body.Scopes, testFlavor{}) != 1000 {
		t.Error("the Locals scope must be found in both responses")
	}
	if localsReference(nil, testFlavor{}) != 0 {
		t.Error("no scopes means reference 0")
	}
}

func TestGoroutinesDropTheCurrentMark(t *testing.T) {
	response := responseFor[*dap.ThreadsResponse](t, functionsSession(t, t.TempDir()))

	goroutines := mapThreads(response.Body.Threads)

	if len(goroutines) != 6 || goroutines[0].Name != "[Go 1] main.multiply (Thread 41936)" {
		t.Errorf("goroutines = %+v", goroutines)
	}
}

func TestStandardOutputCategories(t *testing.T) {
	event := func(category, text string) *dap.OutputEvent {
		return &dap.OutputEvent{Body: dap.OutputEventBody{Category: category, Output: text}}
	}
	if _, category, ok := StandardOutput(event("stdout", "hi\n")); !ok || category != "stdout" {
		t.Errorf("stdout: ok=%v category=%q", ok, category)
	}
	if _, category, _ := StandardOutput(event("important", "x")); category != "console" {
		t.Errorf("unknown categories are console, got %q", category)
	}
	for _, noise := range []*dap.OutputEvent{event("telemetry", "x"), event("stdout", "")} {
		if _, _, ok := StandardOutput(noise); ok {
			t.Errorf("%+v must be dropped", noise.Body)
		}
	}
}

func TestStandardStopReasons(t *testing.T) {
	cases := map[string]domain.StopReason{
		"breakpoint": domain.StopBreakpoint, "step": domain.StopStep, "entry": domain.StopEntry,
		"exception": domain.StopException, "pause": domain.StopPause, "something new": domain.StopPause,
	}
	for reason, want := range cases {
		if got := StandardStopReason(reason); got != want {
			t.Errorf("StandardStopReason(%q) = %q, want %q", reason, got, want)
		}
	}
}
