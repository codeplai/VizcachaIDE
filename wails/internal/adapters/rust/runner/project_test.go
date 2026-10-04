package runner

import (
	"context"
	"errors"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"

	"github.com/codeplai/VizcachaIDE/wails/internal/app"
	"github.com/codeplai/VizcachaIDE/wails/internal/domain"
)

func runAndWait(t *testing.T, r *Runner, sink *testSink, path string) (int, string) {
	t.Helper()
	if err := r.Run(context.Background(), r.Configure(path, nil)); err != nil {
		t.Fatal(err)
	}
	code := sink.waitFinished(t)
	return code, sink.all()
}

func TestConfigureTellsFilesFromCrates(t *testing.T) {
	r := New(nil, Options{AppDir: t.TempDir()})
	loose := looseFile(t, hello)
	if config := r.Configure(loose, nil); config.Mode != domain.RunFile || config.Project != nil {
		t.Fatalf("loose file = %+v", config)
	}
	root := filepath.Join(t.TempDir(), "mi crate")
	crate(t, root, "mi_crate", map[string]string{"src/main.rs": hello})
	config := r.Configure(filepath.Join(root, "src", "main.rs"), nil)
	if config.Mode != domain.RunProject || config.WorkingDir != root || config.Project.Kind != domain.ProjectCargo || config.Project.Name != "mi_crate" {
		t.Fatalf("crate = %+v, %+v", config, config.Project)
	}
	if config.Target != filepath.Join(root, "src", "main.rs") {
		t.Fatalf("target = %s, want the open file (it picks the binary)", config.Target)
	}
}

func TestCrateRunsTheBinaryOfTheOpenFile(t *testing.T) {
	r, sink := newRunner(t)
	root := filepath.Join(t.TempDir(), "mi crate")
	crate(t, root, "mi_crate", map[string]string{
		"src/main.rs":      "fn main() { println!(\"principal\"); }\n",
		"src/bin/extra.rs": "fn main() { println!(\"extra\"); }\n",
	})
	code, _ := runAndWait(t, r, sink, filepath.Join(root, "src", "bin", "extra.rs"))
	if out, _ := sink.text(); code != 0 || !strings.Contains(out, "extra") || strings.Contains(out, "principal") {
		t.Fatalf("code %d, output %q, want the extra binary", code, sink.all())
	}
	if text := sink.all(); strings.Contains(text, `{"`) || !strings.Contains(text, "Finished") {
		t.Fatalf("output = %q, want cargo's text and no JSON", text)
	}
	if !sink.started[0].Echo {
		t.Fatal("the program of a crate runs in the terminal, not under cargo")
	}
}

func TestCrateCompileErrorShowsRenderedText(t *testing.T) {
	r, sink := newRunner(t)
	root := t.TempDir()
	crate(t, root, "roto", map[string]string{"src/main.rs": "fn main() { let x: i32 = \"a\"; }\n"})
	code, text := runAndWait(t, r, sink, filepath.Join(root, "src", "main.rs"))
	if code == 0 || !strings.Contains(text, "error[E0308]: mismatched types") || strings.Contains(text, `{"`) {
		t.Fatalf("code %d, output %q", code, text)
	}
}

func TestWorkspaceMemberRunsWithP(t *testing.T) {
	r, sink := newRunner(t)
	root := t.TempDir()
	write(t, root, "Cargo.toml", "[workspace]\nmembers = [\"crates/*\"]\nresolver = \"2\"\n")
	crate(t, filepath.Join(root, "crates", "uno"), "uno", map[string]string{"src/main.rs": "fn main() { println!(\"soy uno\"); }\n"})
	crate(t, filepath.Join(root, "crates", "dos"), "dos", map[string]string{"src/main.rs": "fn main() { println!(\"soy dos\"); }\n"})
	code, text := runAndWait(t, r, sink, filepath.Join(root, "crates", "dos", "src", "main.rs"))
	if code != 0 || !strings.Contains(text, "soy dos") || strings.Contains(text, "soy uno") {
		t.Fatalf("code %d, output %q", code, text)
	}
	if _, err := os.Stat(filepath.Join(root, "target", "debug")); err != nil {
		t.Fatalf("the workspace shares one target/: %v", err)
	}
}

func TestVirtualRootAsksWhichMemberUnlessThereIsOne(t *testing.T) {
	r, sink := newRunner(t)
	root := t.TempDir()
	write(t, root, "Cargo.toml", "[workspace]\nmembers = [\"app\", \"otra\", \"lib\"]\nresolver = \"2\"\n")
	crate(t, filepath.Join(root, "app"), "app", map[string]string{"src/main.rs": "fn main() { println!(\"app\"); }\n"})
	crate(t, filepath.Join(root, "otra"), "otra", map[string]string{"src/main.rs": "fn main() { println!(\"otra\"); }\n"})
	crate(t, filepath.Join(root, "lib"), "lib", map[string]string{"src/lib.rs": "pub fn f() {}\n"})

	err := r.Run(context.Background(), r.Configure(filepath.Join(root, "Cargo.toml"), nil))
	var choice *ChooseMemberError
	if !errors.As(err, &choice) || !errors.Is(err, ErrChooseMember) || !errors.Is(err, app.ErrUnsupported) {
		t.Fatalf("err = %v, want ErrChooseMember", err)
	}
	if !reflect.DeepEqual(choice.Members, []string{"app", "otra"}) || !strings.Contains(err.Error(), "run.chooseMember") {
		t.Fatalf("members = %v, err = %v", choice.Members, err)
	}

	// Without the second program, the only runnable member is picked.
	if err := os.RemoveAll(filepath.Join(root, "otra", "src")); err != nil {
		t.Fatal(err)
	}
	write(t, filepath.Join(root, "otra"), "src/lib.rs", "pub fn g() {}\n")
	code, text := runAndWait(t, r, sink, filepath.Join(root, "Cargo.toml"))
	if code != 0 || !strings.Contains(text, "app") {
		t.Fatalf("code %d, output %q", code, text)
	}
}

func TestLibraryCrateHasNothingToRun(t *testing.T) {
	r, _ := newRunner(t)
	root := t.TempDir()
	crate(t, root, "biblio", map[string]string{"src/lib.rs": "pub fn f() {}\n"})
	path := filepath.Join(root, "src", "lib.rs")
	for name, call := range map[string]func() error{
		"run":   func() error { return r.Run(context.Background(), r.Configure(path, nil)) },
		"build": func() error { return r.Build(context.Background(), r.Configure(path, nil)) },
	} {
		err := call()
		if !errors.Is(err, ErrNoBinary) || !strings.Contains(err.Error(), "errors.rustNoBinary") {
			t.Fatalf("%s: err = %v, want ErrNoBinary", name, err)
		}
	}
}

func TestCrateBuildAndCompileForDebugUseTheTargetFolder(t *testing.T) {
	r, sink := newRunner(t)
	root := t.TempDir()
	crate(t, root, "mi_crate", map[string]string{"src/main.rs": hello})
	config := r.Configure(filepath.Join(root, "src", "main.rs"), nil)
	if err := r.Build(context.Background(), config); err != nil {
		t.Fatal(err)
	}
	if code := sink.waitFinished(t); code != 0 {
		t.Fatalf("exit code %d: %s", code, sink.all())
	}
	want := filepath.Join(root, "target", "debug", executableName("mi_crate"))
	exe, output, err := r.CompileForDebug(context.Background(), config)
	if err != nil || !strings.EqualFold(filepath.Clean(exe), want) {
		t.Fatalf("exe = %q, err = %v, output = %q, want %s", exe, err, output, want)
	}
}
