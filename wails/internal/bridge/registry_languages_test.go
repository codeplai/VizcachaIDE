package bridge

import (
	"testing"

	"github.com/codeplai/VizcachaIDE/wails/internal/app"
	"github.com/codeplai/VizcachaIDE/wails/internal/domain"
)

// fakes of the two languages every bridge test uses: "go" has every capability, "python" has none
// of the optional ones (no console, formatter, checker or packages).
type testLanguages struct {
	registry               *app.LanguageRegistry
	goRunner, pyRunner     *fakeRunner
	goDebugger, pyDebugger *fakeDebugger
	goServer, pyServer     *fakeServer
	console                *fakeConsole
	checker                *fakeChecker
	packages               *fakePackages
}

func newTestLanguages(t *testing.T) *testLanguages {
	t.Helper()
	l := &testLanguages{
		goRunner:   &fakeRunner{language: "go", tools: []domain.ToolStatus{{ID: "go", CodeLanguage: "go"}}},
		pyRunner:   &fakeRunner{language: "python", tools: []domain.ToolStatus{{ID: "python", CodeLanguage: "python"}}},
		goDebugger: &fakeDebugger{}, pyDebugger: &fakeDebugger{},
		goServer: &fakeServer{}, pyServer: &fakeServer{},
		console: &fakeConsole{}, checker: &fakeChecker{}, packages: &fakePackages{},
	}
	goSupport := app.LanguageSupport{
		Profile: domain.LanguageProfile{
			ID: "go", Extensions: []string{".go"},
			Capabilities: domain.Capabilities{Build: true, Console: true, Format: true, Check: true,
				PackageActions: []domain.PackageAction{domain.PackageInit}},
			Tools: []domain.ToolSpec{{ID: "go", Role: domain.RoleRuntime}, {ID: "dlv", Role: domain.RoleDebugAdapter}},
		},
		Runner: l.goRunner, Debugger: l.goDebugger, LanguageServer: l.goServer, Explainer: fakeExplainer{tag: "go"},
		Console: l.console, Formatter: fakeFormatter{}, Checker: l.checker, Packages: l.packages,
		Search: fakeSearch{},
	}
	pySupport := app.LanguageSupport{
		Profile: domain.LanguageProfile{
			ID: "python", Extensions: []string{".py", ".pyw"},
			Tools: []domain.ToolSpec{
				{ID: "python", Role: domain.RoleRuntime},
				{ID: "debugpy", Role: domain.RoleDebugAdapter, ProvidedBy: "python"},
			},
		},
		Runner: l.pyRunner, Debugger: l.pyDebugger, LanguageServer: l.pyServer, Explainer: fakeExplainer{tag: "py"},
	}
	registry, err := app.NewLanguageRegistry("go", goSupport, pySupport)
	if err != nil {
		t.Fatal(err)
	}
	l.registry = registry
	return l
}

func newTestRegistry(t *testing.T) *app.LanguageRegistry {
	t.Helper()
	return newTestLanguages(t).registry
}
