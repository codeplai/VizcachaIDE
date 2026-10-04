package runner

import (
	"context"
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

const askName = `use std::io::{self, Write};

fn main() {
    print!("Nombre: ");
    io::stdout().flush().unwrap();
    let mut name = String::new();
    io::stdin().read_line(&mut name).unwrap();
    let args: Vec<String> = std::env::args().skip(1).collect();
    println!("Hola, {}! args={}", name.trim(), args.len());
}
`

func TestRunReadsFromStdinThroughTheTerminal(t *testing.T) {
	r, sink := newRunner(t)
	path := looseFile(t, askName)
	if err := r.Run(context.Background(), r.Configure(path, []string{"uno", "dos"})); err != nil {
		t.Fatal(err)
	}
	sink.waitOutput(t, "Nombre:")
	if err := r.WriteInput("Ana"); err != nil {
		t.Fatal(err)
	}
	if code := sink.waitFinished(t); code != 0 {
		t.Fatalf("exit code %d: %s", code, sink.all())
	}
	if out, _ := sink.text(); !strings.Contains(out, "Hola, Ana! args=2") {
		t.Fatalf("stdout = %q", out)
	}
	if len(sink.started) != 1 || !sink.started[0].Echo {
		t.Fatalf("started = %+v, want one run with Echo", sink.started)
	}
	if _, err := os.Stat(filepath.Join(filepath.Dir(path), executableName("main"))); err == nil {
		t.Fatal("the executable must not be left next to the source")
	}
}

func TestCompileErrorShowsTheRenderedTextAndSkipsTheProgram(t *testing.T) {
	r, sink := newRunner(t)
	path := looseFile(t, "fn main() {\n    let v = vec![1];\n    let w = v;\n    println!(\"{:?}\", v);\n}\n")
	if err := r.Run(context.Background(), r.Configure(path, nil)); err != nil {
		t.Fatal(err)
	}
	if code := sink.waitFinished(t); code == 0 {
		t.Fatal("a compile error must not exit with 0")
	}
	text := sink.all()
	if !strings.Contains(text, "error[E0382]: borrow of moved value") || !strings.Contains(text, "-->") {
		t.Fatalf("output = %q, want rustc's rendered text", text)
	}
	if strings.Contains(text, `{"`) || strings.Contains(text, "$message_type") {
		t.Fatalf("output = %q, the JSON must not reach the user", text)
	}
	time.Sleep(200 * time.Millisecond)
	if len(sink.started) != 1 || r.IsRunning() {
		t.Fatal("there must be no second stage")
	}
}

func TestPanicExitsWith101AndAddsNoLine(t *testing.T) {
	r, sink := newRunner(t)
	path := looseFile(t, "fn main() {\n    let v = vec![1, 2, 3];\n    let i = v.len() + 7;\n    println!(\"{}\", v[i]);\n}\n")
	if err := r.Run(context.Background(), r.Configure(path, nil)); err != nil {
		t.Fatal(err)
	}
	if code := sink.waitFinished(t); code != 101 {
		t.Fatalf("exit code %d, want 101: %s", code, sink.all())
	}
	text := sink.all()
	if !strings.Contains(text, "panicked at") || !strings.Contains(text, "index out of bounds") {
		t.Fatalf("output = %q, want the panic", text)
	}
	if strings.Contains(text, lineSegfault) || strings.Contains(text, lineAborted) {
		t.Fatalf("output = %q, a panic gets no extra line", text)
	}
}

func TestBuildLeavesTheExecutableNextToTheSource(t *testing.T) {
	r, sink := newRunner(t)
	path := looseFile(t, hello)
	if err := r.Build(context.Background(), r.Configure(path, nil)); err != nil {
		t.Fatal(err)
	}
	if code := sink.waitFinished(t); code != 0 {
		t.Fatalf("exit code %d: %s", code, sink.all())
	}
	if _, err := os.Stat(filepath.Join(filepath.Dir(path), executableName("main"))); err != nil {
		t.Fatalf("no executable: %v", err)
	}
	if strings.Contains(sink.all(), "hola") {
		t.Fatal("Build must not run the program")
	}
}

func TestRunUntitledCleansUp(t *testing.T) {
	r, sink := newRunner(t)
	config, err := r.RunUntitled(context.Background(), "untitled-1.rs", hello, nil)
	if err != nil {
		t.Fatal(err)
	}
	if filepath.Base(config.Target) != "main.rs" {
		t.Fatalf("target = %s", config.Target)
	}
	if code := sink.waitFinished(t); code != 0 {
		t.Fatalf("exit code %d: %s", code, sink.all())
	}
	if out, _ := sink.text(); !strings.Contains(out, "hola") {
		t.Fatalf("stdout = %q", out)
	}
	waitGone(t, filepath.Dir(config.Target))
	waitGone(t, r.buildFolder(filepath.Dir(config.Target)))
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
	r, _ := newRunner(t)
	path := looseFile(t, "fn main() { let sobra = 1; }\n")
	exe, output, err := r.CompileForDebug(context.Background(), r.Configure(path, nil))
	if err != nil {
		t.Fatalf("err = %v, output = %q", err, output)
	}
	if _, statErr := os.Stat(exe); statErr != nil || !strings.HasPrefix(exe, r.options.CacheDir) {
		t.Fatalf("exe = %q (%v), want a file in the cache", exe, statErr)
	}
	if !strings.Contains(output, "unused variable: `sobra`") || strings.Contains(output, `{"`) {
		t.Fatalf("output = %q, want the rendered warning", output)
	}
	bad := looseFile(t, "fn main() { let x: i32 = \"a\"; }\n")
	exe, output, err = r.CompileForDebug(context.Background(), r.Configure(bad, nil))
	if !errors.Is(err, ErrCompileFailed) || exe != "" || !strings.Contains(output, "E0308") {
		t.Fatalf("exe=%q output=%q err=%v, want a compile failure", exe, output, err)
	}
}
