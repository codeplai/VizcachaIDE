package lldbdap

import (
	"context"
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/codeplai/VizcachaIDE/wails/internal/adapters/rust"
	"github.com/codeplai/VizcachaIDE/wails/internal/adapters/rust/rusttest"
	"github.com/codeplai/VizcachaIDE/wails/internal/app"
	"github.com/codeplai/VizcachaIDE/wails/internal/domain"
	"github.com/codeplai/VizcachaIDE/wails/internal/protocol/dap/lldb"
	"github.com/codeplai/VizcachaIDE/wails/internal/protocol/process"
	"github.com/google/go-dap"
)

const fakeVerbose = "rustc 1.99.0 (b940084d7 2026-09-28)\nhost: x86_64-pc-windows-gnu\nrelease: 1.99.0\n"

// fakeCompiler answers without running anything.
type fakeCompiler struct {
	output string
	err    error
}

func (f fakeCompiler) CompileForDebug(context.Context, domain.RunConfiguration) (string, string, error) {
	return "", f.output, f.err
}

// fakeTools is a locator with fake files: lldb-dap and rustc (answering through the probe) are
// there unless left out.
func fakeTools(t *testing.T, withLldb, withRustc bool) *rust.Locator {
	t.Helper()
	bin := t.TempDir()
	for name, wanted := range map[string]bool{"lldb-dap": withLldb, "rustc": withRustc} {
		if wanted {
			if err := os.WriteFile(filepath.Join(bin, rusttest.Exe(name)), nil, 0o755); err != nil { //nolint:gosec // a fake tool
				t.Fatal(err)
			}
		}
	}
	return rust.NewLocator(rust.Options{
		AppDir: t.TempDir(), BaseEnvironment: []string{"PATH=" + bin},
		Probe: func(_ context.Context, _ string, args ...string) (string, error) {
			if args[0] == "--print" {
				return "C:/rust/sysroot\n", nil
			}
			return fakeVerbose, nil
		},
	})
}

func fakeDebugger(t *testing.T, locator *rust.Locator) (*Debugger, *eventSink) {
	t.Helper()
	sink := newEventSink()
	return New(sink, process.New(sink), locator, translate), sink
}

func TestInitCommandsLoadTheFormattersAndStopAtPanic(t *testing.T) {
	toolchain := rust.Toolchain{Sysroot: `C:\Users\Ana María\rust`}
	commands := InitCommands(toolchain, true)
	if len(commands) != 2 || commands[0] != "breakpoint set --name rust_panic" {
		t.Fatalf("commands = %q", commands)
	}
	// Forward slashes and quotes: LLDB would eat backslashes and break at the space.
	if !strings.HasPrefix(commands[1], `command script import "`) || !strings.HasSuffix(commands[1], `lldb_lookup.py"`) ||
		strings.Contains(commands[1], `\`) || !strings.Contains(commands[1], "Ana María") {
		t.Errorf("import command = %q", commands[1])
	}
}

func TestInitCommandsWithoutPythonOnlyStopAtPanic(t *testing.T) {
	commands := InitCommands(rust.Toolchain{Sysroot: "/sysroot"}, false)
	if len(commands) != 1 || !strings.Contains(commands[0], "rust_panic") {
		t.Errorf("commands = %q", commands)
	}
}

func TestSourceRootIsTheFolderOfAFileOrTheCrate(t *testing.T) {
	loose := t.TempDir()
	config := domain.NewFileRunConfiguration(domain.CodeLanguageRust, filepath.Join(loose, "main.rs"), nil)
	if got := sourceRoot(config); got != loose {
		t.Errorf("loose file root = %q, want %q", got, loose)
	}
	crate := t.TempDir()
	if err := os.MkdirAll(filepath.Join(crate, "src"), 0o750); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(crate, "Cargo.toml"), []byte("[package]\nname = \"demo\"\nversion = \"0.1.0\"\nedition = \"2021\"\n"), 0o600); err != nil {
		t.Fatal(err)
	}
	config = domain.NewFileRunConfiguration(domain.CodeLanguageRust, filepath.Join(crate, "src", "main.rs"), nil)
	if got := sourceRoot(config); got != crate {
		t.Errorf("crate root = %q, want %q", got, crate)
	}
}

func TestStartNeedsLldbDapAndRustc(t *testing.T) {
	config := domain.NewFileRunConfiguration(domain.CodeLanguageRust, "main.rs", nil)
	debugger, _ := fakeDebugger(t, fakeTools(t, false, true))
	if err := debugger.Start(t.Context(), config, nil); !errors.Is(err, app.ErrToolNotFound) || !strings.Contains(err.Error(), "lldb-dap") {
		t.Errorf("without lldb-dap: %v", err)
	}
	debugger, _ = fakeDebugger(t, fakeTools(t, true, false))
	if err := debugger.Start(t.Context(), config, nil); !errors.Is(err, app.ErrToolNotFound) || !strings.Contains(err.Error(), "rustc") {
		t.Errorf("without rustc: %v", err)
	}
}

func TestCompileErrorEndsTheSessionWithTheCompilersText(t *testing.T) {
	debugger, sink := fakeDebugger(t, fakeTools(t, true, true))
	debugger.UseCompiler(fakeCompiler{output: "error[E0308]: mismatched types", err: ErrBuildFailed})
	config := domain.NewFileRunConfiguration(domain.CodeLanguageRust, "bad.rs", nil)
	if err := debugger.Start(t.Context(), config, nil); err != nil {
		t.Fatal(err)
	}
	if code := receive(t, sink.terminated); code != 1 {
		t.Errorf("exit code = %d", code)
	}
	events := sink.debugEvents()
	if len(events) != 1 || events[0] != "stderr:error[E0308]: mismatched types\n" {
		t.Errorf("events = %q", events)
	}
	if debugger.IsActive() {
		t.Error("the session must be over")
	}
}

func TestPanicWatchKeepsTheMessageAcrossChunks(t *testing.T) {
	sink := newEventSink()
	watch := &panicWatch{EventSink: sink}
	watch.DebugOutput("before\r\nthread 'main' (35236) panicked at main.rs:4:21:\r\n", "stderr")
	watch.DebugOutput("index out of", "stderr")
	watch.DebugOutput(" bounds: the len is 3 but the index is 10\r\nnote: run with", "stderr")
	if got := watch.description(); got != "panicked: index out of bounds: the len is 3 but the index is 10" {
		t.Errorf("description = %q", got)
	}
	if !strings.Contains(sink.text(), "before") {
		t.Error("the output must still reach the sink")
	}
}

func TestStopReasonOfThePanicBreakpointIsAnException(t *testing.T) {
	watch := &panicWatch{EventSink: newEventSink(), message: "boom"}
	f := newFlavor(lldb.Options{Roots: []string{"/code"}}, watch)
	panicStop := &dap.StoppedEvent{Body: dap.StoppedEventBody{Reason: "breakpoint", HitBreakpointIds: []int{1}}}
	if reason, description := f.StopReason(panicStop); reason != domain.StopException || description != "panicked: boom" {
		t.Errorf("panic stop = %v %q", reason, description)
	}
	student := &dap.StoppedEvent{Body: dap.StoppedEventBody{Reason: "breakpoint", HitBreakpointIds: []int{2}}}
	if reason, _ := f.StopReason(student); reason != domain.StopBreakpoint {
		t.Errorf("breakpoint of the student = %v", reason)
	}
}

func TestLibraryFramesAreHidden(t *testing.T) {
	f := newFlavor(lldb.Options{Roots: []string{"/work/crate"}}, &panicWatch{EventSink: newEventSink()})
	frame := func(path string) dap.StackFrame { return dap.StackFrame{Source: &dap.Source{Path: path}} }
	for path, keep := range map[string]bool{
		"/work/crate/src/main.rs":                         true,
		"/rustc/b940084d7/library/core/src/panicking.rs":  false,
		"/home/u/.cargo/registry/src/x/rand/src/lib.rs":   false,
		"C:/rust/lib/rustlib/src/rust/library/alloc/x.rs": false,
	} {
		if got := f.KeepFrame(frame(path)); got != keep {
			t.Errorf("KeepFrame(%s) = %v, want %v", path, got, keep)
		}
	}
}

func TestExecutableIsReadFromCargoMessages(t *testing.T) {
	messages := `{"reason":"compiler-artifact","target":{"name":"dep"},"executable":null}
{"reason":"compiler-artifact","target":{"name":"demo"},"executable":"C:\\t\\debug\\demo.exe"}
{"reason":"build-finished","success":true}
`
	if got := executableFrom(messages, "demo"); got != `C:\t\debug\demo.exe` {
		t.Errorf("executable = %q", got)
	}
}
