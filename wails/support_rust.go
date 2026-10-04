package main

import (
	"context"
	"errors"
	"fmt"
	"os"

	"github.com/codeplai/VizcachaIDE/wails/internal/adapters/rust"
	"github.com/codeplai/VizcachaIDE/wails/internal/adapters/rust/analyzer"
	"github.com/codeplai/VizcachaIDE/wails/internal/adapters/rust/clippy"
	rusterrors "github.com/codeplai/VizcachaIDE/wails/internal/adapters/rust/errors"
	"github.com/codeplai/VizcachaIDE/wails/internal/adapters/rust/lldbdap"
	"github.com/codeplai/VizcachaIDE/wails/internal/adapters/rust/packages"
	"github.com/codeplai/VizcachaIDE/wails/internal/adapters/rust/runner"
	"github.com/codeplai/VizcachaIDE/wails/internal/adapters/rust/rustfmt"
	"github.com/codeplai/VizcachaIDE/wails/internal/app"
	"github.com/codeplai/VizcachaIDE/wails/internal/bridge"
	"github.com/codeplai/VizcachaIDE/wails/internal/domain"
	"github.com/codeplai/VizcachaIDE/wails/internal/protocol/lsp"
	"github.com/codeplai/VizcachaIDE/wails/internal/protocol/process"
)

// newRustSupport creates the adapters of Rust (docs/PLAN_RUST.md section 4.8) around one locator,
// so they agree on the toolchain and read it once, and the supervisor every language shares. The
// debugger compiles with the runner's stage 1, so running and debugging build the same way
// (including the pinned MinGW linker). Rust has no console; the returned function has nothing
// more to release.
func newRustSupport(sink *bridge.WailsEventSink, store app.SettingsStore, texts *backendTexts, supervisor *process.Supervisor) (app.LanguageSupport, func(context.Context), error) {
	explainer, err := rusterrors.NewExplainer()
	if err != nil {
		return app.LanguageSupport{}, nil, err
	}
	base := os.Environ()
	locator := rust.NewLocator(rust.Options{Settings: store, BaseEnvironment: base})
	rustRunner := runner.New(supervisor, runner.Options{
		Settings: store, BaseEnvironment: base, Locator: locator,
		CompilingNotice: func() string { return texts.text("run.compiling") },
	})
	debugger := lldbdap.New(sink, supervisor, locator, texts.text)
	debugger.UseCompiler(rustDebugBuild{runner: rustRunner})
	support := app.LanguageSupport{
		Profile:  rust.Profile,
		Runner:   rustRunner,
		Debugger: debugger,
		LanguageServer: analyzer.NewRouter(sink, analyzer.Config{Locator: locator},
			lsp.Options{IdleTimeout: languageServerIdle}),
		Explainer: explainer,
		Formatter: rustfmt.New(locator),
		Checker:   clippy.New(locator),
		Packages:  packages.New(supervisor, rustRunner),
	}
	return support, func(context.Context) {}, nil
}

// rustDebugBuild gives the debugger the runner's compile step, translating its "rejected by the
// compiler" error into the one the debugger knows.
type rustDebugBuild struct{ runner *runner.Runner }

func (b rustDebugBuild) CompileForDebug(ctx context.Context, config domain.RunConfiguration) (string, string, error) {
	exe, output, err := b.runner.CompileForDebug(ctx, config)
	if errors.Is(err, runner.ErrCompileFailed) {
		return "", output, fmt.Errorf("%w: %w", lldbdap.ErrBuildFailed, err)
	}
	return exe, output, err
}
