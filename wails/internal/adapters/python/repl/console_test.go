package repl

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/codeplai/VizcachaIDE/wails/internal/adapters/python"
	"github.com/codeplai/VizcachaIDE/wails/internal/adapters/python/pythontest"
	"github.com/codeplai/VizcachaIDE/wails/internal/app"
	"github.com/codeplai/VizcachaIDE/wails/internal/domain"
)

type fixedFinder struct {
	path string
	err  error
}

func (f fixedFinder) Find(context.Context, string) (python.Interpreter, error) {
	return python.Interpreter{Path: f.path, Source: domain.ToolConfigured}, f.err
}

func translate(key string) string { return "T:" + key }

// newConsole returns a console on the real interpreter, or skips the test.
func newConsole(t *testing.T, timeout time.Duration) *Console {
	t.Helper()
	c := New(Options{Finder: fixedFinder{path: pythontest.Interpreter(t)}, Translate: translate, Timeout: timeout, CacheDir: t.TempDir()})
	t.Cleanup(c.Reset)
	return c
}

func TestExpressionShowsItsValue(t *testing.T) {
	c := newConsole(t, 20*time.Second)
	if got := c.Eval("2 + 2"); got.Result != "4" || got.Error != "" || got.Output != "" {
		t.Fatalf("got %+v", got)
	}
	if got := c.Eval(`"hola"`); got.Result != `'hola'` {
		t.Fatalf("string should be quoted, got %+v", got)
	}
}

func TestSessionIsPersistentAndStatementsAreSilent(t *testing.T) {
	c := newConsole(t, 20*time.Second)
	if got := c.Eval("x = 5"); got.Result != "" || got.Error != "" {
		t.Fatalf("statement should be silent, got %+v", got)
	}
	if got := c.Eval("def doble(n):\n    return n * 2\n"); got.Error != "" {
		t.Fatalf("got %+v", got)
	}
	if got := c.Eval("doble(x)"); got.Result != "10" {
		t.Fatalf("session lost, got %+v", got)
	}
	c.Reset()
	if got := c.Eval("x"); !strings.Contains(got.Error, "NameError") {
		t.Fatalf("Reset should forget x, got %+v", got)
	}
}

func TestPrintedOutputIsCapturedWithAccents(t *testing.T) {
	c := newConsole(t, 20*time.Second)
	got := c.Eval("print('¡hola, ñandú!')")
	if got.Output != "¡hola, ñandú!\n" || got.Result != "" || got.Error != "" {
		t.Fatalf("got %+v", got)
	}
}

func TestErrorIsExplainedWithoutReplFrames(t *testing.T) {
	c := newConsole(t, 20*time.Second)
	got := c.Eval("def f():\n    return 1 / 0\nf()")
	if !strings.Contains(got.Error, "ZeroDivisionError: division by zero") || !strings.Contains(got.Error, `File "<console>"`) {
		t.Fatalf("got %+v", got)
	}
	if strings.Contains(got.Error, "repl.py") || strings.Contains(got.Error, "compile_snippet") {
		t.Fatalf("frames of repl.py leaked: %s", got.Error)
	}
	syntax := c.Eval("for i in range(3)")
	if !strings.Contains(syntax.Error, "SyntaxError") || strings.Contains(syntax.Error, "repl.py") {
		t.Fatalf("got %+v", syntax)
	}
	if again := c.Eval("1 + 1"); again.Result != "2" {
		t.Fatalf("session should survive errors, got %+v", again)
	}
}

func TestInputGivesTheNoKeyboardText(t *testing.T) {
	c := newConsole(t, 20*time.Second)
	if got := c.Eval(`input("Nombre: ")`); got.Error != "T:errors.consoleNoInput" {
		t.Fatalf("got %+v", got)
	}
	if got := c.Eval("import sys\nsys.stdin.readline()"); got.Error != "T:errors.consoleNoInput" {
		t.Fatalf("got %+v", got)
	}
	if got := c.Eval("2 * 3"); got.Result != "6" {
		t.Fatalf("session should survive, got %+v", got)
	}
}

func TestTimeoutKillsTheProcessAndTheNextEvalWorks(t *testing.T) {
	c := newConsole(t, 1500*time.Millisecond)
	c.Eval("x = 1")
	start := time.Now()
	got := c.Eval("while True: pass")
	if !strings.HasPrefix(got.Error, "timeout") || time.Since(start) > 10*time.Second {
		t.Fatalf("got %+v after %s", got, time.Since(start))
	}
	if next := c.Eval("1 + 1"); next.Result != "2" || next.Error != "" {
		t.Fatalf("next Eval = %+v", next)
	}
	if lost := c.Eval("x"); !strings.Contains(lost.Error, "NameError") {
		t.Fatalf("a killed session is forgotten, got %+v", lost)
	}
}

func TestMissingPythonSaysSo(t *testing.T) {
	c := New(Options{Finder: fixedFinder{err: app.MissingTool(python.ToolPython)}, Translate: translate, CacheDir: t.TempDir()})
	if got := c.Eval("1"); got.Error != "T:errors.pythonNotFound" {
		t.Fatalf("got %+v", got)
	}
	c.Reset() // nothing started: must not panic
}

func TestScriptIsCopiedAndRewrittenWhenItChanges(t *testing.T) {
	cache := t.TempDir()
	path, err := installScript(cache)
	if err != nil || path != filepath.Join(cache, "VizcachaIDE", "repl", "repl.py") {
		t.Fatalf("path = %q, err = %v", path, err)
	}
	if err := os.WriteFile(path, []byte("# old"), 0o644); err != nil {
		t.Fatal(err)
	}
	if _, err := installScript(cache); err != nil {
		t.Fatal(err)
	}
	if content, _ := os.ReadFile(path); string(content) != string(script) {
		t.Fatal("repl.py was not rewritten")
	}
}
