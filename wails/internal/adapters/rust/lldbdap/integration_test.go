package lldbdap

import (
	"strings"
	"testing"

	"github.com/codeplai/VizcachaIDE/wails/internal/domain"
)

// Line numbers matter: the tests put breakpoints on them.
const factorialSource = `fn factorial(n: u64) -> u64 {
    let mut result: u64 = 1;
    for i in 2..=n {
        result *= i;
    }
    result
}

fn main() {
    let count: i32 = 3;
    let name = String::from("Ana");
    let numbers: Vec<i32> = vec![1, 2, 3];
    let maybe: Option<i32> = Some(7);
    let none: Option<i32> = None;
    let total = factorial(5);
    println!("{} {} {} {:?} {:?} {}", name, total, numbers.len(), maybe, none, count);
}
`

func TestRealDebugFactorial(t *testing.T) {
	debugger, sink, _ := realDebugger(t)
	path := writeProgram(t, "factorial.rs", factorialSource)
	start(t, debugger, path, 4)

	first := receive(t, sink.stops)
	if first.Reason != domain.StopBreakpoint || first.Frames[0].Location.Line != 4 {
		t.Fatalf("first stop = %+v", first)
	}
	values := variablesOf(first)
	if values["n"].Value != "5" || values["result"].Value != "1" || values["i"].Value != "2" {
		t.Errorf("variables = %+v", first.Variables)
	}
	if err := debugger.StepOver(); err != nil {
		t.Fatal(err)
	}
	second := receive(t, sink.stops)
	if second.Reason != domain.StopStep || second.Frames[0].Location.Line != 3 {
		t.Fatalf("after the step: %+v", second)
	}
	if got := variablesOf(second)["result"].Value; got != "2" {
		t.Errorf("result after the step = %q, want 2", got)
	}
	if sink.runEvents() != 0 {
		t.Error("debugging must not emit run:* events")
	}
}

func TestRealStdTypesWithFormatters(t *testing.T) {
	debugger, sink, _ := realDebugger(t)
	path := writeProgram(t, "types.rs", factorialSource)
	start(t, debugger, path, 16)

	stop := receive(t, sink.stops)
	locals := variablesOf(stop)
	for _, name := range []string{"count", "name", "numbers", "maybe", "none", "total"} {
		v := locals[name]
		t.Logf("%s = %q (%s, ref %d)", name, v.Value, v.TypeName, v.Reference)
	}
	if locals["count"].Value != "3" || locals["total"].Value != "120" {
		t.Errorf("scalars = %+v", stop.Variables)
	}
	if !strings.Contains(locals["name"].Value, "Ana") {
		t.Errorf("String shows %q, want its text", locals["name"].Value)
	}
	if err := debugger.RequestVariables(locals["numbers"].Reference); err != nil {
		t.Fatal(err)
	}
	children := receive(t, sink.children)
	t.Logf("Vec children: %+v", children)
	if !strings.Contains(locals["numbers"].Value, "3") || len(children) < 3 {
		t.Errorf("Vec<i32> shows %q with %d children", locals["numbers"].Value, len(children))
	}
	// LLDB 23 read Some(7) as garbage on *-gnu until data/vizcacha_rust_enums.py.
	if locals["maybe"].Value != "Some(7)" || locals["none"].Value != "None" {
		t.Errorf("Option shows %q and %q, want Some(7) and None", locals["maybe"].Value, locals["none"].Value)
	}
}

func TestRealStdinThroughSupervisor(t *testing.T) {
	debugger, sink, supervisor := realDebugger(t)
	source := "use std::io::{self, Write};\nfn main() {\n    print!(\"Name: \");\n    io::stdout().flush().unwrap();\n    let mut name = String::new();\n    io::stdin().read_line(&mut name).unwrap();\n    println!(\"Hello {}\", name.trim());\n}\n"
	start(t, debugger, writeProgram(t, "hello.rs", source))

	waitFor(t, sink, "Name:")
	if err := supervisor.WriteInput("Ana"); err != nil {
		t.Fatal(err)
	}
	waitFor(t, sink, "Hello Ana")
	if code := receive(t, sink.terminated); code != 0 {
		t.Errorf("exit code = %d", code)
	}
	if sink.runEvents() != 0 {
		t.Error("the program output must not become run:* events")
	}
}

func TestRealPanicStopsInStudentCode(t *testing.T) {
	debugger, sink, _ := realDebugger(t)
	source := "fn main() {\n    let numbers = vec![1, 2, 3];\n    let index = numbers.len() + 7;\n    println!(\"before\");\n    let value = numbers[index];\n    println!(\"{}\", value);\n}\n"
	start(t, debugger, writeProgram(t, "panic.rs", source))

	stop := receive(t, sink.stops)
	t.Logf("panic stop: %+v", stop)
	if stop.Reason != domain.StopException || !strings.Contains(stop.Description, "index out of bounds") {
		t.Fatalf("stop = %+v", stop)
	}
	if len(stop.Frames) == 0 || stop.Frames[0].Location == nil || stop.Frames[0].Location.Line != 5 {
		t.Errorf("stopped at %+v, want line 5", stop.Frames)
	}
}

func TestRealCompileError(t *testing.T) {
	debugger, sink, _ := realDebugger(t)
	start(t, debugger, writeProgram(t, "bad.rs", "fn main() {\n    let x: i32 = \"a\";\n}\n"))

	if code := receive(t, sink.terminated); code == 0 {
		t.Error("a compile error must not end with exit code 0")
	}
	var shown string
	for _, event := range sink.debugEvents() {
		if strings.HasPrefix(event, "stderr:") && strings.Contains(event, "bad.rs:2") {
			shown = event
		}
	}
	if shown == "" {
		t.Errorf("the compiler output did not reach debug:output: %q", sink.debugEvents())
	}
	if len(sink.stops) != 0 {
		t.Error("nothing must stop after a compile error")
	}
}
