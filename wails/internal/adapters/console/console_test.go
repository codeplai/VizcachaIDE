package console

import (
	"strings"
	"testing"
	"time"
)

func TestExpressionShowsItsValue(t *testing.T) {
	c := New(0)
	if got := c.Eval("2 + 3"); got.Result != "5" || got.Error != "" {
		t.Fatalf("got %+v", got)
	}
	if got := c.Eval(`"hi"`); got.Result != `"hi"` {
		t.Fatalf("string should be quoted, got %+v", got)
	}
}

func TestStatementShowsNothingAndKeepsVariables(t *testing.T) {
	c := New(0)
	if got := c.Eval("x := 5"); got.Result != "" || got.Error != "" {
		t.Fatalf("statement should be silent, got %+v", got)
	}
	if got := c.Eval("x * 2"); got.Result != "10" {
		t.Fatalf("variable lost, got %+v", got)
	}
}

func TestPrintedOutputIsCaptured(t *testing.T) {
	c := New(0)
	c.Eval("import \"fmt\"")
	got := c.Eval("fmt.Println(\"hola\")")
	if got.Output != "hola\n" || got.Error != "" {
		t.Fatalf("got %+v", got)
	}
}

func TestCompileErrorIsReported(t *testing.T) {
	c := New(0)
	got := c.Eval("y := undefinedName + 1")
	if got.Error == "" {
		t.Fatalf("expected an error, got %+v", got)
	}
	if got := c.Eval("1 + 1"); got.Result != "2" {
		t.Fatalf("session should survive an error, got %+v", got)
	}
}

func TestTimeoutResetsTheSession(t *testing.T) {
	c := New(200 * time.Millisecond)
	c.Eval("z := 1")
	got := c.Eval("for { }")
	if !strings.Contains(got.Error, "timeout") {
		t.Fatalf("expected timeout, got %+v", got)
	}
	if got := c.Eval("z"); got.Error == "" {
		t.Fatalf("variables should be gone after the timeout, got %+v", got)
	}
	if got := c.Eval("3 + 4"); got.Result != "7" {
		t.Fatalf("console should work again, got %+v", got)
	}
}

func TestResetClearsVariables(t *testing.T) {
	c := New(0)
	c.Eval("a := 1")
	c.Reset()
	if got := c.Eval("a"); got.Error == "" {
		t.Fatalf("a should be undefined after Reset, got %+v", got)
	}
}

func TestStandardPackagesNeedNoImport(t *testing.T) {
	c := New(0)
	if got := c.Eval(`strings.ToUpper("go")`); got.Result != `"GO"` || got.Error != "" {
		t.Fatalf("got %+v", got)
	}
	if got := c.Eval(`strings.Repeat("a", 2)`); got.Result != `"aa"` {
		t.Fatalf("second use, got %+v", got)
	}
	got := c.Eval(`fmt.Println("hi")`)
	if got.Output != "hi\n" || got.Result != "" {
		t.Fatalf("Println should only print, got %+v", got)
	}
}

func TestExplicitImportThenUse(t *testing.T) {
	c := New(0)
	c.Eval("import \"math\"")
	if got := c.Eval("math.Sqrt(16)"); got.Result != "4" || got.Error != "" {
		t.Fatalf("got %+v", got)
	}
}

func TestFunctionsPersist(t *testing.T) {
	c := New(0)
	c.Eval("func double(n int) int { return n * 2 }")
	if got := c.Eval("double(4)"); got.Result != "8" {
		t.Fatalf("got %+v", got)
	}
}
