package main

import (
	"context"
	"os"

	"github.com/codeplai/VizcachaIDE/wails/internal/adapters/python"
	"github.com/codeplai/VizcachaIDE/wails/internal/adapters/python/debugpy"
	pythonerrors "github.com/codeplai/VizcachaIDE/wails/internal/adapters/python/errors"
	"github.com/codeplai/VizcachaIDE/wails/internal/adapters/python/packages"
	"github.com/codeplai/VizcachaIDE/wails/internal/adapters/python/pylsp"
	"github.com/codeplai/VizcachaIDE/wails/internal/adapters/python/repl"
	"github.com/codeplai/VizcachaIDE/wails/internal/adapters/python/ruff"
	"github.com/codeplai/VizcachaIDE/wails/internal/adapters/python/runner"
	"github.com/codeplai/VizcachaIDE/wails/internal/app"
	"github.com/codeplai/VizcachaIDE/wails/internal/bridge"
	"github.com/codeplai/VizcachaIDE/wails/internal/protocol/lsp"
	"github.com/codeplai/VizcachaIDE/wails/internal/protocol/process"
)

// newPythonSupport creates the adapters of Python (docs/PLAN_PYTHON.md section 4.10) around one
// interpreter locator, so they agree on the interpreter and probe it once, and the supervisor every
// language shares. The returned function releases what the adapters hold beyond what shutdown
// already stops (runner, debugger and language server): the console's interpreter process.
func newPythonSupport(sink *bridge.WailsEventSink, store app.SettingsStore, texts *backendTexts, supervisor *process.Supervisor) (app.LanguageSupport, func(context.Context), error) {
	explainer, err := pythonerrors.NewExplainer()
	if err != nil {
		return app.LanguageSupport{}, nil, err
	}
	base := os.Environ()
	locator := python.NewLocator(python.Options{Settings: store, BaseEnvironment: base})
	pythonRunner := runner.New(supervisor, runner.Options{Settings: store, BaseEnvironment: base, Locator: locator})
	formatter := ruff.New(ruff.Config{Locator: locator, BaseEnvironment: base})
	console := repl.New(repl.Options{Finder: locator, Translate: texts.text, BaseEnvironment: base})
	support := app.LanguageSupport{
		Profile:  python.Profile,
		Runner:   pythonRunner,
		Debugger: debugpy.New(sink, supervisor, locator, base, texts.text),
		LanguageServer: pylsp.New(sink, pylsp.Config{Locator: locator, BaseEnvironment: base},
			lsp.Options{IdleTimeout: languageServerIdle}),
		Explainer: explainer,
		Console:   console,
		Formatter: formatter,
		Checker:   formatter,
		Packages:  packages.New(supervisor, pythonRunner),
	}
	return support, func(context.Context) { console.Reset() }, nil
}
