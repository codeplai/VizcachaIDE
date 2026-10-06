package main

import (
	"context"
	"errors"
	"fmt"
	"os"

	"github.com/codeplai/VizcachaIDE/wails/internal/adapters/cpp"
	"github.com/codeplai/VizcachaIDE/wails/internal/adapters/cpp/clangd"
	"github.com/codeplai/VizcachaIDE/wails/internal/adapters/cpp/clangformat"
	cpperrors "github.com/codeplai/VizcachaIDE/wails/internal/adapters/cpp/errors"
	"github.com/codeplai/VizcachaIDE/wails/internal/adapters/cpp/lldbdap"
	"github.com/codeplai/VizcachaIDE/wails/internal/adapters/cpp/runner"
	"github.com/codeplai/VizcachaIDE/wails/internal/adapters/cpp/vcpkg"
	"github.com/codeplai/VizcachaIDE/wails/internal/app"
	"github.com/codeplai/VizcachaIDE/wails/internal/bridge"
	"github.com/codeplai/VizcachaIDE/wails/internal/domain"
	"github.com/codeplai/VizcachaIDE/wails/internal/protocol/lsp"
	"github.com/codeplai/VizcachaIDE/wails/internal/protocol/process"
)

// newCppSupport creates the adapters of C++ (docs/PLAN_CPP.md section 4) around one locator, so
// they agree on the compiler and probe it once, and the supervisor every language shares. The
// debugger compiles with the runner's stage 1, so running and debugging build the same way. The
// libraries come from vcpkg (docs/PLAN_CPP_CMAKE.md section 4.3). C++ has no console; the returned
// function has nothing more to release.
func newCppSupport(sink *bridge.WailsEventSink, store app.SettingsStore, texts *backendTexts, supervisor *process.Supervisor) (app.LanguageSupport, func(context.Context), error) {
	explainer, err := cpperrors.NewExplainer()
	if err != nil {
		return app.LanguageSupport{}, nil, err
	}
	base := os.Environ()
	locator := cpp.NewLocator(cpp.Options{Settings: store, BaseEnvironment: base})
	libraries := newVcpkgLocator(store, base, locator)
	cppRunner := runner.New(supervisor, runner.Options{
		Settings: store, BaseEnvironment: base, Locator: locator,
		CompilingNotice: func() string { return texts.text("run.compiling") },
		Dependencies:    vcpkg.NewSetup(libraries), Language: texts.language,
	})
	manager := vcpkg.NewManager(libraries, vcpkg.ManagerOptions{
		Configurer: lazyConfigurer{runner: cppRunner}, Events: sink,
		SlowNotice: func() string { return texts.text("packages.cppSlowInstall") },
	})
	search := vcpkg.NewSearch(libraries)
	search.Warm()
	debugger := lldbdap.New(sink, supervisor, locator, base, texts.text)
	debugger.UseCompiler(debugBuild{runner: cppRunner, manager: manager})
	support := app.LanguageSupport{
		Profile:  cpp.Profile,
		Runner:   cppSlot{Runner: cppRunner, manager: manager},
		Debugger: debugger,
		LanguageServer: clangd.New(sink, clangd.Config{Locator: locator, BaseEnvironment: base},
			lsp.Options{IdleTimeout: languageServerIdle}),
		Explainer: explainer,
		Formatter: clangformat.New(locator),
		Checker:   cppRunner,
		Packages:  cppPackages{Manager: manager, supervisor: supervisor},
		Search:    search,
		Scaffold:  cpp.Scaffold{Language: texts.language},
		Shell:     cppShell{Locator: locator, libraries: libraries},
	}
	return support, func(context.Context) {}, nil
}

// debugBuild gives the debugger the runner's compile step, translating its "rejected by the
// compiler" error into the one the debugger knows.
type debugBuild struct {
	runner  *runner.Runner
	manager *vcpkg.Manager
}

func (b debugBuild) CompileForDebug(ctx context.Context, config domain.RunConfiguration) (string, string, error) {
	if b.manager.Busy() {
		return "", "", app.ErrBusy
	}
	exe, output, err := b.runner.CompileForDebug(ctx, config)
	if errors.Is(err, runner.ErrCompileFailed) {
		return "", output, fmt.Errorf("%w: %w", lldbdap.ErrBuildFailed, err)
	}
	return exe, output, err
}
