package lldbdap

import (
	"strings"
	"testing"

	"github.com/codeplai/VizcachaIDE/wails/internal/adapters/cpp/cpptest"
	"github.com/codeplai/VizcachaIDE/wails/internal/domain"
)

// Line numbers matter: the tests put breakpoints on them.
const factorialSource = `#include <iostream>
#include <string>
#include <vector>

int factorial(int n) {
    int result = 1;
    for (int i = 2; i <= n; i++) {
        result *= i;
    }
    return result;
}

int main() {
    std::string name = "Ana";
    std::vector<int> numbers = {1, 2, 3};
    int total = factorial(5);
    std::cout << name << " " << total << " " << numbers.size() << std::endl;
    return 0;
}
`

// debugFactorial walks breakpoint, variables, step and a second breakpoint with the tools of bins.
func debugFactorial(t *testing.T, bins ...string) {
	debugger, sink, _ := debuggerWith(t, bins...)
	compiler, _ := debugger.locator.Compiler(t.Context())
	t.Logf("compiler: %s", compiler.Version)
	path := writeProgram(t, "factorial.cpp", factorialSource)
	start(t, debugger, path, 8)

	first := receive(t, sink.stops)
	if first.Reason != domain.StopBreakpoint || first.Frames[0].Location.Line != 8 {
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
	if second.Reason != domain.StopStep || second.Frames[0].Location.Line != 7 {
		t.Fatalf("after the step: %+v", second)
	}
	if got := variablesOf(second)["result"].Value; got != "2" {
		t.Errorf("result after the step = %q, want 2", got)
	}
	if err := debugger.SetBreakpoints(path, []int{17}); err != nil {
		t.Fatal(err)
	}
	if err := debugger.Resume(); err != nil {
		t.Fatal(err)
	}
	last := receive(t, sink.stops)
	if last.Frames[0].Location.Line != 17 {
		t.Fatalf("last stop = %+v", last)
	}
	locals := variablesOf(last)
	t.Logf("locals: name=%q (%s) numbers=%q (%s, ref %d) total=%q",
		locals["name"].Value, locals["name"].TypeName, locals["numbers"].Value, locals["numbers"].TypeName, locals["numbers"].Reference, locals["total"].Value)
	if locals["total"].Value != "120" || !strings.Contains(locals["name"].Value, "Ana") {
		t.Errorf("locals = %+v", last.Variables)
	}
	if err := debugger.RequestVariables(locals["numbers"].Reference); err != nil {
		t.Fatal(err)
	}
	children := receive(t, sink.children)
	t.Logf("vector children: %+v", children)
	if locals["numbers"].Value != "size=3" || len(children) < 3 {
		t.Errorf("std::vector shows %q with %d children, want size=3 and 3 elements", locals["numbers"].Value, len(children))
	}
	if sink.runEvents() != 0 {
		t.Error("debugging must not emit run:* events")
	}
}

func TestRealLLDBWithClang(t *testing.T) { debugFactorial(t, cpptest.LLVMBin(t)) }

// TestRealLLDBWithGCCBinary builds with g++ and debugs with the lldb-dap of LLVM.
func TestRealLLDBWithGCCBinary(t *testing.T) {
	debugFactorial(t, cpptest.GCCBin(t), cpptest.LLVMBin(t))
}

func TestRealLLDBInputThroughSupervisor(t *testing.T) {
	debugger, sink, supervisor := debuggerWith(t, cpptest.LLVMBin(t))
	source := "#include <iostream>\n#include <string>\nint main() {\n    std::string name;\n    std::cout << \"Name: \" << std::flush;\n    std::cin >> name;\n    std::cout << \"Hello \" << name << std::endl;\n}\n"
	start(t, debugger, writeProgram(t, "hello.cpp", source))

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

func TestRealLLDBWithoutTerminalFallsBack(t *testing.T) {
	debugger, sink, _ := debuggerWith(t, cpptest.LLVMBin(t))
	debugger.hasTerminal = func(string) bool { return false }
	path := writeProgram(t, "factorial.cpp", factorialSource)
	start(t, debugger, path, 8)

	first := receive(t, sink.stops)
	if first.Reason != domain.StopBreakpoint || variablesOf(first)["n"].Value != "5" {
		t.Fatalf("first stop = %+v", first)
	}
	if !strings.Contains(sink.text(), "run.debugStdin") {
		t.Errorf("the no-keyboard notice is missing: %q", sink.debugEvents())
	}
	_ = debugger.SetBreakpoints(path, nil)
	if err := debugger.Resume(); err != nil {
		t.Fatal(err)
	}
	waitFor(t, sink, "Ana 120 3")
}

func TestRealLLDBNullPointerCrash(t *testing.T) {
	debugger, sink, _ := debuggerWith(t, cpptest.LLVMBin(t))
	source := "#include <iostream>\nint main() {\n    int* p = nullptr;\n    std::cout << \"before\" << std::endl;\n    *p = 7;\n    return 0;\n}\n"
	start(t, debugger, writeProgram(t, "crash.cpp", source))

	stop := receive(t, sink.stops)
	t.Logf("crash stop: %+v", stop)
	if stop.Reason != domain.StopException || !strings.HasPrefix(stop.Description, "Segmentation fault") {
		t.Fatalf("stop = %+v", stop)
	}
	if stop.Frames[0].Location == nil || stop.Frames[0].Location.Line != 5 {
		t.Errorf("stopped at %+v, want line 5", stop.Frames[0])
	}
}

func TestRealLLDBCompileError(t *testing.T) {
	debugger, sink, _ := debuggerWith(t, cpptest.LLVMBin(t))
	start(t, debugger, writeProgram(t, "bad.cpp", "int main() {\n    int x = ;\n}\n"))

	if code := receive(t, sink.terminated); code == 0 {
		t.Error("a compile error must not end with exit code 0")
	}
	var shown string
	for _, event := range sink.debugEvents() {
		if strings.HasPrefix(event, "stderr:") && strings.Contains(event, "bad.cpp:2") {
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
