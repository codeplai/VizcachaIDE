package main

import (
	"context"
	"time"

	"github.com/codeplai/VizcachaIDE/wails/internal/adapters/golang"
	"github.com/codeplai/VizcachaIDE/wails/internal/adapters/golang/console"
	"github.com/codeplai/VizcachaIDE/wails/internal/adapters/golang/delve"
	golangerrors "github.com/codeplai/VizcachaIDE/wails/internal/adapters/golang/errors"
	"github.com/codeplai/VizcachaIDE/wails/internal/adapters/golang/gopls"
	"github.com/codeplai/VizcachaIDE/wails/internal/adapters/golang/packages"
	"github.com/codeplai/VizcachaIDE/wails/internal/adapters/golang/runner"
	"github.com/codeplai/VizcachaIDE/wails/internal/app"
	"github.com/codeplai/VizcachaIDE/wails/internal/bridge"
	"github.com/codeplai/VizcachaIDE/wails/internal/protocol/lsp"
	"github.com/codeplai/VizcachaIDE/wails/internal/protocol/process"
)

// languageServerIdle is how long a language server lives without open documents.
const languageServerIdle = 5 * time.Minute

// newGoSupport creates the adapters of Go and gathers them in one LanguageSupport. Every runner
// shares the supervisor, so one program runs at a time in the whole IDE. Tool paths are resolved
// when a run, a debug session or gopls starts (configured -> bundled -> PATH), so changing them in
// Settings needs no restart. The returned function releases what the adapters hold beyond what
// shutdown already stops (runner, debugger and language server): Go holds nothing.
func newGoSupport(sink *bridge.WailsEventSink, store app.SettingsStore, texts *backendTexts, supervisor *process.Supervisor) (app.LanguageSupport, func(context.Context), error) {
	explainer, err := golangerrors.NewExplainer()
	if err != nil {
		return app.LanguageSupport{}, nil, err
	}
	goRunner := runner.New(supervisor, runner.Options{
		Settings:         store,
		FirstBuildNotice: func() string { return texts.text("run.firstBuild") },
	})
	support := app.LanguageSupport{
		Profile: golang.Profile,
		Runner:  goRunner,
		Debugger: delve.New(sink, delve.Options{
			DelvePath:   func() string { return goRunner.Locate("dlv").Path },
			Environment: goRunner.Environment,
			Translate:   texts.text,
		}),
		LanguageServer: gopls.New(sink, gopls.Config{
			Executable:  func() string { return goRunner.Locate("gopls").Path },
			Environment: goRunner.Environment,
		}, lsp.Options{IdleTimeout: languageServerIdle}),
		Explainer: explainer,
		Console:   console.New(0),
		Formatter: goRunner,
		Checker:   goRunner,
		Packages:  packages.New(supervisor, goRunner),
		Scaffold:  packages.Scaffold{},
	}
	return support, func(context.Context) {}, nil
}
