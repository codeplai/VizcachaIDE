package bridge

import (
	"errors"
	"reflect"
	"testing"

	"github.com/codeplai/VizcachaIDE/wails/internal/app"
	"github.com/codeplai/VizcachaIDE/wails/internal/domain"
)

func TestLanguageServiceRoutesByPath(t *testing.T) {
	l := newTestLanguages(t)
	service := NewLanguageService(l.registry)
	if err := service.OpenDocument("a.go", ""); err != nil {
		t.Fatal(err)
	}
	if err := service.OpenDocument("b.py", ""); err != nil {
		t.Fatal(err)
	}
	if text, err := service.Hover(domain.SourceLocation{File: "c.py"}); err != nil || text != "hover" {
		t.Errorf("Hover = %q, %v", text, err)
	}
	if !reflect.DeepEqual(l.goServer.opened, []string{"a.go"}) || !reflect.DeepEqual(l.pyServer.opened, []string{"b.py"}) ||
		!reflect.DeepEqual(l.pyServer.hovers, []string{"c.py"}) || len(l.goServer.hovers) != 0 {
		t.Errorf("go = %+v, python = %+v", l.goServer, l.pyServer)
	}
	// A file no language claims has no intelligence, but it is not an error.
	if text, err := service.Hover(domain.SourceLocation{File: "README.txt"}); err != nil || text != "" {
		t.Errorf("unknown file Hover = %q, %v", text, err)
	}
	if symbols, err := service.DocumentSymbols("README.txt"); err != nil || symbols == nil || len(symbols) != 0 {
		t.Errorf("unknown file symbols = %v, %v", symbols, err)
	}
}

func TestAssistantRoutesDiagnosticsByFile(t *testing.T) {
	l := newTestLanguages(t)
	store := NewMemorySettingsStore()
	sink := &explainedSink{}
	service := NewAssistantService(sink, l.registry, NewLanguageResolver(store, nil))
	at := func(file string) *domain.SourceLocation { return &domain.SourceLocation{File: file} }

	items, err := service.ExplainDiagnostics("", []domain.Diagnostic{
		{Severity: domain.SeverityError, Message: "a", Location: at("x.py")},
		{Severity: domain.SeverityError, Message: "b", Location: at("x.go")},
		{Severity: domain.SeverityError, Message: "c"}, // no file: the default language
	})
	if err != nil {
		t.Fatal(err)
	}
	var titles []string
	for _, item := range items {
		titles = append(titles, item.Explanation.Title)
	}
	if want := []string{"py:en:a", "go:en:b", "go:en:c"}; !reflect.DeepEqual(titles, want) {
		t.Errorf("titles = %v, want %v", titles, want)
	}
	// Without a file, the language of the last run wins over the default one.
	crash, err := service.ExplainDiagnostics(domain.CodeLanguagePython, []domain.Diagnostic{{Severity: domain.SeverityError, Message: "d"}})
	if err != nil || len(crash) != 1 || crash[0].Explanation.Title != "py:en:d" {
		t.Errorf("fallback = %+v, %v", crash, err)
	}
	if _, err := service.Explain("cobol", "x", ""); !errors.Is(err, app.ErrUnknownCodeLanguage) {
		t.Errorf("Explain with an unknown language: %v", err)
	}
}

func TestConsoleNeedsAConsole(t *testing.T) {
	l := newTestLanguages(t)
	service := NewConsoleService(l.registry)
	if result, err := service.Eval("go", "1+1"); err != nil || result.Result != "1+1" {
		t.Errorf("Eval = %+v, %v", result, err)
	}
	if err := service.Reset("go"); err != nil || l.console.resets != 1 {
		t.Errorf("Reset: %v", err)
	}
	if _, err := service.Eval("python", "1"); !errors.Is(err, app.ErrUnsupported) {
		t.Errorf("Eval without a console: %v", err)
	}
	if err := service.Reset("python"); !errors.Is(err, app.ErrUnsupported) {
		t.Errorf("Reset without a console: %v", err)
	}
	if err := service.Reset("cobol"); !errors.Is(err, app.ErrUnknownCodeLanguage) {
		t.Errorf("Reset of an unknown language: %v", err)
	}
}

func TestPackagesRouteByLanguage(t *testing.T) {
	l := newTestLanguages(t)
	service := NewPackagesService(l.registry)
	if err := service.Init("go", "/w", "demo"); err != nil {
		t.Fatal(err)
	}
	if err := service.Add("go", "/w", "x/y"); err != nil {
		t.Fatal(err)
	}
	if err := service.Tidy("go", "/w"); err != nil {
		t.Fatal(err)
	}
	if want := []string{"init /w demo", "add /w x/y", "tidy /w"}; !reflect.DeepEqual(l.packages.calls, want) {
		t.Errorf("calls = %v, want %v", l.packages.calls, want)
	}
	if err := service.Remove("go", "/w", "x"); !errors.Is(err, app.ErrUnsupported) {
		t.Errorf("a verb the manager lacks: %v", err)
	}
	if err := service.List("python", "/w"); !errors.Is(err, app.ErrUnsupported) {
		t.Errorf("a language without a package manager: %v", err)
	}
	if err := service.Tidy("cobol", "/w"); !errors.Is(err, app.ErrUnknownCodeLanguage) {
		t.Errorf("an unknown language: %v", err)
	}
}

func TestCodeLanguagesServiceListsProfilesAndAllTools(t *testing.T) {
	service := NewCodeLanguagesService(newTestLanguages(t).registry)
	profiles := service.Profiles()
	if len(profiles) != 2 || profiles[0].ID != "go" || profiles[1].ID != "python" {
		t.Errorf("profiles = %+v", profiles)
	}
	tools := service.Tools()
	if len(tools) != 2 || tools[0].ID != "go" || tools[1].ID != "python" {
		t.Errorf("tools = %+v", tools)
	}
}
