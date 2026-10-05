package runner

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

const askName = `#include <iostream>
#include <string>
int main(int argc, char** argv) {
    std::string name;
    std::cout << "Nombre: ";
    std::cin >> name;
    std::cout << "Hola, " << name << "! args=" << (argc - 1) << std::endl;
    return 0;
}
`

func TestRunReadsFromStdinThroughTheTerminal(t *testing.T) {
	eachCompiler(t, func(t *testing.T, r *Runner, sink *testSink) {
		path := writeSource(t, askName)
		if err := r.Run(context.Background(), r.Configure(path, []string{"uno", "dos"})); err != nil {
			t.Fatal(err)
		}
		sink.waitOutput(t, "Nombre:")
		if err := r.WriteInput("Ana"); err != nil {
			t.Fatal(err)
		}
		if code := sink.waitFinished(t); code != 0 {
			t.Fatalf("exit code %d, stderr %q", code, errText(sink))
		}
		out, _ := sink.text()
		if !strings.Contains(out, "Hola, Ana! args=2") {
			t.Fatalf("stdout = %q", out)
		}
		if len(sink.started) != 1 || !sink.started[0].Echo {
			t.Fatalf("started = %+v, want one run with Echo", sink.started)
		}
		if _, err := os.Stat(filepath.Join(filepath.Dir(path), "main.exe")); err == nil {
			t.Fatal("the executable must not be left next to the source")
		}
	})
}

func errText(sink *testSink) string {
	_, text := sink.text()
	return text
}

func TestRunProjectCompilesEverySource(t *testing.T) {
	eachCompiler(t, func(t *testing.T, r *Runner, sink *testSink) {
		path := writeSource(t, "int suma(int a, int b);\n#include <iostream>\nint main() { std::cout << suma(2, 3) << std::endl; }\n")
		folder := filepath.Dir(path)
		if err := os.WriteFile(filepath.Join(folder, "suma.cpp"), []byte("int suma(int a, int b) { return a + b; }\n"), 0o600); err != nil {
			t.Fatal(err)
		}
		config := r.Configure(path, nil)
		if err := r.Run(context.Background(), config); err != nil {
			t.Fatal(err)
		}
		if code := sink.waitFinished(t); code != 0 {
			t.Fatalf("exit code %d, stderr %q", code, errText(sink))
		}
		if out, _ := sink.text(); !strings.Contains(out, "5") {
			t.Fatalf("stdout = %q", out)
		}
	})
}

func TestCompileErrorEndsTheRunWithTheCompilerCode(t *testing.T) {
	eachCompiler(t, func(t *testing.T, r *Runner, sink *testSink) {
		path := writeSource(t, "int main() {\n    return totl;\n}\n")
		if err := r.Run(context.Background(), r.Configure(path, nil)); err != nil {
			t.Fatal(err)
		}
		if code := sink.waitFinished(t); code == 0 {
			t.Fatal("a compile error must not exit with 0")
		}
		if text := errText(sink); !strings.Contains(text, "totl") {
			t.Fatalf("stderr = %q, want the compiler's message", text)
		}
		time.Sleep(200 * time.Millisecond)
		if len(sink.started) != 1 || r.IsRunning() {
			t.Fatal("there must be no second stage")
		}
	})
}

func TestBuildLeavesTheExecutableNextToTheSource(t *testing.T) {
	eachCompiler(t, func(t *testing.T, r *Runner, sink *testSink) {
		path := writeSource(t, "int main() { return 0; }\n")
		if err := r.Build(context.Background(), r.Configure(path, nil)); err != nil {
			t.Fatal(err)
		}
		if code := sink.waitFinished(t); code != 0 {
			t.Fatalf("exit code %d, stderr %q", code, errText(sink))
		}
		exe := filepath.Join(filepath.Dir(path), executableName("mis_programas_nandu"))
		if _, err := os.Stat(exe); err != nil {
			t.Fatalf("no executable: %v", err)
		}
	})
}

func TestRunUntitledCleansUp(t *testing.T) {
	eachCompiler(t, func(t *testing.T, r *Runner, sink *testSink) {
		config, err := r.RunUntitled(context.Background(), "untitled-1.cpp", "#include <iostream>\nint main() { std::cout << \"listo\\n\"; }\n", nil)
		if err != nil {
			t.Fatal(err)
		}
		if filepath.Base(config.Target) != "main.cpp" {
			t.Fatalf("target = %s", config.Target)
		}
		if code := sink.waitFinished(t); code != 0 {
			t.Fatalf("exit code %d, stderr %q", code, errText(sink))
		}
		if out, _ := sink.text(); !strings.Contains(out, "listo") {
			t.Fatalf("stdout = %q", out)
		}
		waitGone(t, filepath.Dir(config.Target))
		waitGone(t, r.buildFolder(filepath.Dir(config.Target)))
	})
}

func waitGone(t *testing.T, folder string) {
	t.Helper()
	deadline := time.Now().Add(5 * time.Second)
	for time.Now().Before(deadline) {
		if _, err := os.Stat(folder); os.IsNotExist(err) {
			return
		}
		time.Sleep(50 * time.Millisecond)
	}
	t.Fatalf("%s was not removed", folder)
}

func TestCompileForDebugReturnsTheExecutable(t *testing.T) {
	eachCompiler(t, func(t *testing.T, r *Runner, _ *testSink) {
		path := writeSource(t, "int main() { int sobra = 1; return 0; }\n")
		exe, output, err := r.CompileForDebug(context.Background(), r.Configure(path, nil))
		if err != nil {
			t.Fatalf("err = %v, output = %q", err, output)
		}
		if _, statErr := os.Stat(exe); statErr != nil || !strings.HasPrefix(exe, filepath.Join(filepath.Dir(path), "build")) {
			t.Fatalf("exe = %q (%v), want a file in the project's build folder", exe, statErr)
		}
		if !strings.Contains(output, "sobra") {
			t.Fatalf("output = %q, want the unused variable warning", output)
		}
		bad := writeSource(t, "int main() { return totl; }\n")
		exe, output, err = r.CompileForDebug(context.Background(), r.Configure(bad, nil))
		if err == nil || exe != "" || !strings.Contains(output, "totl") {
			t.Fatalf("exe=%q output=%q err=%v, want a compile failure", exe, output, err)
		}
	})
}
