package runner

import (
	"context"
	"errors"
	"path/filepath"
	"strings"
	"testing"

	"github.com/codeplai/VizcachaIDE/wails/internal/app"
)

const unitTests = "pub fn double(n: i32) -> i32 { n * 2 }\n" +
	"#[cfg(test)]\nmod tests {\n    use super::*;\n    #[test]\n    fn doubles() { assert_eq!(double(2), 4); }\n}\n"

func TestTestRunsCargoTestOfALibraryCrate(t *testing.T) {
	r, sink := newRunner(t)
	root := t.TempDir()
	crate(t, root, "doblar", map[string]string{"src/lib.rs": unitTests})
	config, err := r.Test(context.Background(), filepath.Join(root, "src", "lib.rs"))
	if err != nil {
		t.Fatal(err)
	}
	if config.Target != "cargo test" {
		t.Fatalf("target %q", config.Target)
	}
	if code := sink.waitFinished(t); code != 0 {
		t.Fatalf("exit code %d: %s", code, sink.all())
	}
	if !strings.Contains(sink.all(), "test result: ok. 1 passed") {
		t.Fatalf("no test result: %s", sink.all())
	}
}

func TestTestReportsAFailingTest(t *testing.T) {
	r, sink := newRunner(t)
	root := t.TempDir()
	failing := strings.Replace(unitTests, "double(2), 4", "double(2), 5", 1)
	crate(t, root, "doblar", map[string]string{"src/lib.rs": failing})
	if _, err := r.Test(context.Background(), filepath.Join(root, "src", "lib.rs")); err != nil {
		t.Fatal(err)
	}
	if code := sink.waitFinished(t); code == 0 {
		t.Fatalf("a failing test must not exit with 0: %s", sink.all())
	}
	if !strings.Contains(sink.all(), "doubles") {
		t.Fatalf("the failing test is not named: %s", sink.all())
	}
}

func TestTestNeedsACargoProject(t *testing.T) {
	r, _ := newRunner(t)
	_, err := r.Test(context.Background(), looseFile(t, hello))
	if !errors.Is(err, ErrNoCargoProject) || !errors.Is(err, app.ErrUnsupported) {
		t.Fatalf("got %v", err)
	}
}
