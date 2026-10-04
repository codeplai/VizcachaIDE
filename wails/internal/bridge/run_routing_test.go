package bridge

import (
	"errors"
	"reflect"
	"testing"

	"github.com/codeplai/VizcachaIDE/wails/internal/app"
	"github.com/codeplai/VizcachaIDE/wails/internal/domain"
)

func TestRunRoutesByTheExtensionOfThePath(t *testing.T) {
	l := newTestLanguages(t)
	service := NewRunService(l.registry)

	config, err := service.Run("/w/main.GO", []string{"a"})
	if err != nil || l.goRunner.ran != 1 || config.CodeLanguage != "go" || l.pyRunner.ran != 0 {
		t.Fatalf("go run: config = %+v, err = %v", config, err)
	}
	if config, err = service.Run("/w/tool.py", nil); err != nil || l.pyRunner.ran != 1 || config.CodeLanguage != "python" {
		t.Fatalf("python run: config = %+v, err = %v", config, err)
	}
	if _, err := service.Build("/w/main.go", nil); err != nil || l.goRunner.built != 1 {
		t.Errorf("build: err = %v, built = %d", err, l.goRunner.built)
	}
}

func TestRunRejectsUnknownExtensionsAndPropagatesBusy(t *testing.T) {
	l := newTestLanguages(t)
	service := NewRunService(l.registry)
	if _, err := service.Run("notes.txt", nil); !errors.Is(err, app.ErrUnknownCodeLanguage) {
		t.Errorf("unknown extension error = %v", err)
	}
	l.goRunner.runErr = app.ErrBusy
	if _, err := service.Run("main.go", nil); !errors.Is(err, app.ErrBusy) {
		t.Errorf("busy error = %v", err)
	}
}

func TestRunUntitledUsesTheExtensionOfTheName(t *testing.T) {
	l := newTestLanguages(t)
	service := NewRunService(l.registry)
	if _, err := service.RunUntitled("untitled-1.py", "print()", nil); err != nil || l.pyRunner.untitledPath != "untitled-1.py" {
		t.Errorf("err = %v, path = %q", err, l.pyRunner.untitledPath)
	}
	if _, err := service.RunUntitled("untitled-2", "", nil); !errors.Is(err, app.ErrUnknownCodeLanguage) {
		t.Errorf("a name without extension: error = %v", err)
	}
}

func TestCheckAndFormatNeedTheirPorts(t *testing.T) {
	l := newTestLanguages(t)
	service := NewRunService(l.registry)

	out, err := service.Check(domain.RunConfiguration{CodeLanguage: "go", Target: "main.go"})
	if err != nil || out != "vet output" || len(l.checker.configs) != 1 {
		t.Errorf("Check = %q, %v", out, err)
	}
	if _, err := service.Check(domain.RunConfiguration{CodeLanguage: "python"}); !errors.Is(err, app.ErrUnsupported) {
		t.Errorf("Check without a checker: %v", err)
	}
	if _, err := service.Check(domain.RunConfiguration{CodeLanguage: "cobol"}); !errors.Is(err, app.ErrUnknownCodeLanguage) {
		t.Errorf("Check of an unknown language: %v", err)
	}
	if text, err := service.Format("a.go", "x"); err != nil || text != "a.go|x" {
		t.Errorf("Format = %q, %v", text, err)
	}
	if _, err := service.Format("a.py", "x"); !errors.Is(err, app.ErrUnsupported) {
		t.Errorf("Format without a formatter: %v", err)
	}
}

func TestStopAndInputGoToTheRunnerOfTheLastProgram(t *testing.T) {
	l := newTestLanguages(t)
	service := NewRunService(l.registry)
	l.pyRunner.running = true // before any run, the runner that reports a program serves
	if err := service.Stop(); err != nil || l.pyRunner.stopped != 1 {
		t.Fatalf("Stop before a run: err = %v, python stops = %d", err, l.pyRunner.stopped)
	}

	if _, err := service.Run("main.go", nil); err != nil {
		t.Fatal(err)
	}
	if err := service.WriteInput("hi"); err != nil || !reflect.DeepEqual(l.goRunner.inputs, []string{"hi"}) {
		t.Errorf("WriteInput: err = %v, inputs = %v", err, l.goRunner.inputs)
	}
	if err := service.Stop(); err != nil || l.goRunner.stopped != 1 {
		t.Errorf("Stop: err = %v, go stops = %d", err, l.goRunner.stopped)
	}
}

func TestARefusedRunDoesNotHideTheRunningRunner(t *testing.T) {
	l := newTestLanguages(t)
	l.registry = mustRegistry(t, app.UnavailableSupport(domain.LanguageProfile{ID: "python", Extensions: []string{".py"}}), goSupportOf(l))
	service := NewRunService(l.registry)
	if _, err := service.Run("main.go", nil); err != nil {
		t.Fatal(err)
	}
	if _, err := service.Run("tool.py", nil); !errors.Is(err, app.ErrUnsupported) {
		t.Fatalf("an unavailable language must answer ErrUnsupported, got %v", err)
	}
	if err := service.Stop(); err != nil || l.goRunner.stopped != 1 {
		t.Errorf("Stop must still reach the Go runner: err = %v, stops = %d", err, l.goRunner.stopped)
	}
}

func goSupportOf(l *testLanguages) app.LanguageSupport {
	support, _ := l.registry.ForID("go")
	return support
}

func mustRegistry(t *testing.T, supports ...app.LanguageSupport) *app.LanguageRegistry {
	t.Helper()
	registry, err := app.NewLanguageRegistry("go", supports...)
	if err != nil {
		t.Fatal(err)
	}
	return registry
}
