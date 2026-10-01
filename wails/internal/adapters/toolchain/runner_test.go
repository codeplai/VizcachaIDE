package toolchain

import (
	"context"
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/codeplai/VizcachaIDE/wails/internal/app"
	"github.com/codeplai/VizcachaIDE/wails/internal/domain"
)

const helloProgram = "package main\n\nimport \"fmt\"\n\nfunc main() { fmt.Println(\"hola desde go\") }\n"

const echoProgram = `package main

import (
	"bufio"
	"fmt"
	"os"
	"strings"
)

func main() {
	if len(os.Args) > 1 {
		fmt.Println("args:" + strings.Join(os.Args[1:], "|"))
		return
	}
	line, _ := bufio.NewReader(os.Stdin).ReadString('\n')
	fmt.Println("got:" + strings.TrimSpace(line))
}
`

func TestRunHello(t *testing.T) {
	tc, sink := newTestToolchain(t)
	dir := writeFiles(t, map[string]string{"main.go": helloProgram})
	config := app.ConfigurationForFile(filepath.Join(dir, "main.go"), nil)

	event := runAndWait(t, tc, sink, config)

	if event.ExitCode != 0 || !strings.Contains(sink.Stdout(), "hola desde go") {
		t.Errorf("exit %d, stdout %q, stderr %q", event.ExitCode, sink.Stdout(), sink.Stderr())
	}
	if len(sink.started) != 1 || tc.IsRunning() {
		t.Errorf("started=%d running=%v", len(sink.started), tc.IsRunning())
	}
}

func TestRunSendsStdin(t *testing.T) {
	tc, sink := newTestToolchain(t)
	dir := writeFiles(t, map[string]string{"main.go": echoProgram})
	config := app.ConfigurationForFile(filepath.Join(dir, "main.go"), nil)
	if err := tc.Run(context.Background(), config); err != nil {
		t.Fatal(err)
	}
	if err := tc.WriteInput("Ana\n"); err != nil {
		t.Fatal(err)
	}
	sink.waitFinished(t)
	if !strings.Contains(sink.Stdout(), "got:Ana") {
		t.Errorf("stdout = %q", sink.Stdout())
	}
}

func TestRunReportsCompileErrors(t *testing.T) {
	tc, sink := newTestToolchain(t)
	source := "package main\n\nfunc main() { x := 1 }\n"
	dir := writeFiles(t, map[string]string{"main.go": source})

	event := runAndWait(t, tc, sink, app.ConfigurationForFile(filepath.Join(dir, "main.go"), nil))

	if event.ExitCode == 0 || !strings.Contains(sink.Stderr(), "declared and not used") {
		t.Errorf("exit %d, stderr %q", event.ExitCode, sink.Stderr())
	}
}

func TestRunProgramArgumentsWithQuotes(t *testing.T) {
	tc, sink := newTestToolchain(t)
	dir := writeFiles(t, map[string]string{"main.go": echoProgram})
	args, err := app.SplitProgramArguments(`uno "dos tres" 'cuatro'`)
	if err != nil {
		t.Fatal(err)
	}

	runAndWait(t, tc, sink, app.ConfigurationForFile(filepath.Join(dir, "main.go"), args))

	if !strings.Contains(sink.Stdout(), "args:uno|dos tres|cuatro") {
		t.Errorf("stdout = %q, stderr = %q", sink.Stdout(), sink.Stderr())
	}
}

func TestRunModuleOfTwoFiles(t *testing.T) {
	tc, sink := newTestToolchain(t)
	dir := writeFiles(t, map[string]string{
		"go.mod":  "module example.com/dos\n\ngo 1.21\n",
		"main.go": "package main\n\nimport \"fmt\"\n\nfunc main() { fmt.Println(saludo()) }\n",
		"util.go": "package main\n\nfunc saludo() string { return \"desde util\" }\n",
	})
	config := app.ConfigurationForFile(filepath.Join(dir, "main.go"), nil)
	if config.Mode != domain.RunPackage || config.Module == nil {
		t.Fatalf("config = %+v, want package mode", config)
	}

	event := runAndWait(t, tc, sink, config)

	if event.ExitCode != 0 || !strings.Contains(sink.Stdout(), "desde util") {
		t.Errorf("exit %d, stdout %q, stderr %q", event.ExitCode, sink.Stdout(), sink.Stderr())
	}
}

func TestRunUntitledUsesATemporaryFolder(t *testing.T) {
	tc, sink := newTestToolchain(t)

	config, err := tc.RunUntitled(context.Background(), helloProgram, nil)
	if err != nil {
		t.Fatal(err)
	}
	event := sink.waitFinished(t)

	if event.ExitCode != 0 || !strings.Contains(sink.Stdout(), "hola desde go") {
		t.Errorf("exit %d, stdout %q, stderr %q", event.ExitCode, sink.Stdout(), sink.Stderr())
	}
	if _, err := os.Stat(config.WorkingDir); !errors.Is(err, os.ErrNotExist) {
		t.Errorf("the temporary folder %s must be deleted, stat error: %v", config.WorkingDir, err)
	}
}

func TestBuildProducesAnExecutable(t *testing.T) {
	tc, sink := newTestToolchain(t)
	dir := writeFiles(t, map[string]string{"main.go": helloProgram})
	config := app.ConfigurationForFile(filepath.Join(dir, "main.go"), nil)

	if err := tc.Build(context.Background(), config); err != nil {
		t.Fatal(err)
	}
	event := sink.waitFinished(t)

	built := filepath.Join(dir, config.ExecutableName(isWindows()))
	if _, err := os.Stat(built); event.ExitCode != 0 || err != nil {
		t.Errorf("exit %d, executable %s: %v, stderr %q", event.ExitCode, built, err, sink.Stderr())
	}
}

func TestGoModInitAndTidy(t *testing.T) {
	tc, sink := newTestToolchain(t)
	dir := writeFiles(t, map[string]string{"main.go": helloProgram})
	initArgs, err := app.ModInitArguments("example.com/hola")
	if err != nil {
		t.Fatal(err)
	}

	for _, args := range [][]string{initArgs, app.ModTidyArguments()} {
		if err := tc.RunGoCommand(context.Background(), dir, args); err != nil {
			t.Fatal(err)
		}
		if event := sink.waitFinished(t); event.ExitCode != 0 {
			t.Fatalf("go %v exit %d: %s", args, event.ExitCode, sink.Stderr())
		}
	}
	text, err := app.ReadSourceFile(filepath.Join(dir, "go.mod"))
	if err != nil || app.ParseModulePath(text) != "example.com/hola" {
		t.Errorf("go.mod = %q, %v", text, err)
	}
}

func TestRunWithoutGoReportsToolNotFound(t *testing.T) {
	sink := newTestSink()
	tc := New(Options{Sink: sink, AppDir: t.TempDir(), BaseEnvironment: []string{"PATH="}})

	err := tc.Run(context.Background(), domain.NewFileRunConfiguration("main.go", nil))

	if !errors.Is(err, app.ErrToolNotFound) {
		t.Errorf("error = %v, want ErrToolNotFound", err)
	}
}

func TestFirstBuildNoticeAppearsWhenSilent(t *testing.T) {
	sink := newTestSink()
	slow := writeFiles(t, map[string]string{"main.go": "package main\n\nimport \"time\"\n\nfunc main() { time.Sleep(time.Second) }\n"})
	requireGo(t)
	tc := New(Options{
		Sink: sink, AppDir: t.TempDir(),
		FirstBuildNotice: func() string { return "preparando" },
		FirstBuildDelay:  10 * time.Millisecond,
	})

	runAndWait(t, tc, sink, app.ConfigurationForFile(filepath.Join(slow, "main.go"), nil))

	if !strings.Contains(sink.Stdout(), "preparando") {
		t.Errorf("stdout = %q, want the first build notice", sink.Stdout())
	}
}
