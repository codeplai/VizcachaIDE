package app

import (
	"context"
	"errors"
	"testing"

	"github.com/codeplai/VizcachaIDE/wails/internal/domain"
)

// Fakes that only have to exist: the registry never calls them.
type (
	portlessRunner    struct{ ProgramRunner }
	portlessDebugger  struct{ Debugger }
	portlessServer    struct{ LanguageServer }
	portlessExplainer struct{ ErrorExplainer }
	portlessConsole   struct{ Console }
	portlessFormatter struct{ CodeFormatter }
	portlessChecker   struct{ CodeChecker }
	portlessPackages  struct{ PackageManager }
)

func supportOf(id domain.CodeLanguage, extensions ...string) LanguageSupport {
	return LanguageSupport{
		Profile:        domain.LanguageProfile{ID: id, Extensions: extensions},
		Runner:         portlessRunner{},
		Debugger:       portlessDebugger{},
		LanguageServer: portlessServer{},
		Explainer:      portlessExplainer{},
	}
}

func fullSupport(id domain.CodeLanguage, extensions ...string) LanguageSupport {
	support := supportOf(id, extensions...)
	support.Profile.Capabilities = domain.Capabilities{
		Console: true, Format: true, Check: true, PackageActions: []domain.PackageAction{domain.PackageAdd},
	}
	support.Console, support.Formatter, support.Checker, support.Packages =
		portlessConsole{}, portlessFormatter{}, portlessChecker{}, portlessPackages{}
	return support
}

func TestRegistryFindsLanguagesByPathAndID(t *testing.T) {
	registry, err := NewLanguageRegistry("go", fullSupport("go", ".go"), supportOf("python", ".py", ".PYW"))
	if err != nil {
		t.Fatal(err)
	}
	cases := map[string]domain.CodeLanguage{"main.go": "go", "/x/Main.GO": "go", "a.py": "python", "a.pyw": "python"}
	for path, want := range cases {
		if got, ok := registry.ForPath(path); !ok || got.Profile.ID != want {
			t.Errorf("ForPath(%q) = %q, %v; want %q", path, got.Profile.ID, ok, want)
		}
	}
	for _, path := range []string{"Makefile", "notes.txt", ""} {
		if _, ok := registry.ForPath(path); ok {
			t.Errorf("ForPath(%q) must not match", path)
		}
	}
	if got, ok := registry.ForID("python"); !ok || got.Profile.ID != "python" {
		t.Errorf("ForID(python) = %v, %v", got.Profile.ID, ok)
	}
	if _, ok := registry.ForID("cpp"); ok {
		t.Error("ForID(cpp) must not match")
	}
	if registry.Default().Profile.ID != "go" || len(registry.All()) != 2 || len(registry.Profiles()) != 2 {
		t.Errorf("default = %q, all = %d", registry.Default().Profile.ID, len(registry.All()))
	}
}

func TestRegistryRejectsInconsistentSupports(t *testing.T) {
	consoleMissing := supportOf("go", ".go")
	consoleMissing.Profile.Capabilities.Console = true
	consoleExtra := supportOf("go", ".go")
	consoleExtra.Console = portlessConsole{}
	formatMissing := supportOf("go", ".go")
	formatMissing.Profile.Capabilities.Format = true
	checkExtra := supportOf("go", ".go")
	checkExtra.Checker = portlessChecker{}
	packagesMissing := supportOf("go", ".go")
	packagesMissing.Profile.Capabilities.PackageActions = []domain.PackageAction{domain.PackageInit}
	noRunner := supportOf("go", ".go")
	noRunner.Runner = nil

	cases := map[string][]LanguageSupport{
		"console without port":    {consoleMissing},
		"console port unlisted":   {consoleExtra},
		"format without port":     {formatMissing},
		"checker unlisted":        {checkExtra},
		"packages without port":   {packagesMissing},
		"missing runner":          {noRunner},
		"same extension":          {supportOf("go", ".go"), supportOf("python", ".GO")},
		"same id":                 {supportOf("go", ".go"), supportOf("go", ".golang")},
		"no id":                   {supportOf("", ".x")},
		"default is not in":       {supportOf("python", ".py")},
		"default compared by id":  {},
		"capability on a go port": {consoleMissing, supportOf("python", ".py")},
	}
	for name, supports := range cases {
		if _, err := NewLanguageRegistry("go", supports...); !errors.Is(err, ErrInconsistentProfile) {
			t.Errorf("%s: error = %v, want ErrInconsistentProfile", name, err)
		}
	}
}

func TestUnavailableSupportAnswersUnsupported(t *testing.T) {
	profile := domain.LanguageProfile{
		ID: "python", Extensions: []string{".py"},
		Capabilities: domain.Capabilities{Console: true, Format: true, Check: true,
			PackageActions: []domain.PackageAction{domain.PackageAdd}},
	}
	support := UnavailableSupport(profile)
	registry, err := NewLanguageRegistry("python", support)
	if err != nil {
		t.Fatalf("the capability check must be skipped: %v", err)
	}
	if support.Console != nil || support.Formatter != nil || support.Checker != nil || support.Packages != nil {
		t.Error("optional ports must be nil")
	}
	ctx := context.Background()
	runner := registry.Default().Runner
	config := runner.Configure("prog.py", nil)
	if config.CodeLanguage != "python" || config.Mode != domain.RunFile {
		t.Errorf("config = %+v", config)
	}
	for name, err := range map[string]error{
		"run":      runner.Run(ctx, config),
		"build":    runner.Build(ctx, config),
		"debugger": support.Debugger.Start(ctx, config, nil),
		"step":     support.Debugger.StepOver(),
	} {
		if !errors.Is(err, ErrUnsupported) {
			t.Errorf("%s error = %v, want ErrUnsupported", name, err)
		}
	}
	if _, err := runner.RunUntitled(ctx, "untitled-1.py", "", nil); !errors.Is(err, ErrUnsupported) {
		t.Errorf("RunUntitled error = %v", err)
	}
	items, err := support.LanguageServer.Completion(ctx, domain.SourceLocation{})
	if err != nil || len(items) != 0 || runner.IsRunning() || support.Debugger.IsActive() {
		t.Errorf("language server = %v, %v", items, err)
	}
}
