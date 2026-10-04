// Command vizcacha is the VizcachaIDE desktop app: a Go IDE for beginners.
//
// This file is the composition root: the only place that creates adapters,
// services and the Wails application.
package main

import (
	"context"
	"embed"
	"log"
	"time"

	"github.com/codeplai/VizcachaIDE/wails/internal/adapters/filesystem"
	"github.com/codeplai/VizcachaIDE/wails/internal/adapters/filewatch"
	"github.com/codeplai/VizcachaIDE/wails/internal/adapters/golang/console"
	"github.com/codeplai/VizcachaIDE/wails/internal/adapters/golang/delve"
	golangerrors "github.com/codeplai/VizcachaIDE/wails/internal/adapters/golang/errors"
	"github.com/codeplai/VizcachaIDE/wails/internal/adapters/golang/gopls"
	"github.com/codeplai/VizcachaIDE/wails/internal/adapters/golang/packages"
	"github.com/codeplai/VizcachaIDE/wails/internal/adapters/golang/runner"
	"github.com/codeplai/VizcachaIDE/wails/internal/adapters/settings"
	"github.com/codeplai/VizcachaIDE/wails/internal/adapters/windowstate"
	"github.com/codeplai/VizcachaIDE/wails/internal/bridge"
	"github.com/codeplai/VizcachaIDE/wails/internal/domain"
	"github.com/codeplai/VizcachaIDE/wails/internal/i18n"
	"github.com/codeplai/VizcachaIDE/wails/internal/protocol/lsp"
	"github.com/codeplai/VizcachaIDE/wails/internal/protocol/process"
	"github.com/wailsapp/wails/v2"
	"github.com/wailsapp/wails/v2/pkg/options"
	"github.com/wailsapp/wails/v2/pkg/options/assetserver"
	"github.com/wailsapp/wails/v2/pkg/options/windows"
)

//go:embed all:frontend/dist
var assets embed.FS

func main() {
	if err := run(); err != nil {
		log.Fatalf("vizcacha: %v", err)
	}
}

// backend is everything the Wails application binds and the way to close it.
type backend struct {
	services []any
	shutdown func(ctx context.Context)
}

func run() error {
	sink := bridge.NewWailsEventSink()
	parts, err := newBackend(sink)
	if err != nil {
		return err
	}
	window, err := newWindowKeeper()
	if err != nil {
		return err
	}
	saved := window.State()
	startState := options.Normal
	if saved.Maximised {
		startState = options.Maximised
	}
	return wails.Run(&options.App{
		Title:            "VizcachaIDE",
		Width:            saved.Width,
		Height:           saved.Height,
		MinWidth:         domain.WindowMinWidth,
		MinHeight:        domain.WindowMinHeight,
		WindowStartState: startState,
		AssetServer: &assetserver.Options{
			Assets: assets,
		},
		BackgroundColour: &options.RGBA{R: 245, G: 247, B: 249, A: 1},
		OnStartup:        func(ctx context.Context) { sink.SetContext(ctx) },
		OnDomReady:       window.Restore,
		OnBeforeClose:    window.Save,
		OnShutdown:       parts.shutdown,
		Bind:             parts.services,
		Windows:          &windows.Options{Theme: windows.SystemDefault},
	})
}

// newWindowKeeper remembers the window geometry in window.json (not in settings.json).
func newWindowKeeper() (*bridge.WindowKeeper, error) {
	store, err := windowstate.NewDefaultStore()
	if err != nil {
		return nil, err
	}
	return bridge.NewWindowKeeper(store), nil
}

// newGoTools creates the Delve and gopls adapters. Tool paths are resolved when a debug
// session or gopls starts (configured -> bundled -> PATH), so changing them in Settings
// needs no restart.
func newGoTools(sink *bridge.WailsEventSink, goRunner *runner.Runner, texts *backendTexts) (*delve.Debugger, *lsp.Server) {
	debugger := delve.New(sink, delve.Options{
		DelvePath:   func() string { return goRunner.Locate("dlv").Path },
		Environment: goRunner.Environment,
		Translate:   texts.text,
	})
	languageServer := gopls.New(sink, gopls.Config{
		Executable:  func() string { return goRunner.Locate("gopls").Path },
		Environment: goRunner.Environment,
	}, lsp.Options{IdleTimeout: 5 * time.Minute})
	return debugger, languageServer
}

// newBackend creates the adapters and the services that use them.
func newBackend(sink *bridge.WailsEventSink) (*backend, error) {
	store, err := settings.NewDefaultStore()
	if err != nil {
		return nil, err
	}
	language := bridge.NewLanguageResolver(store, i18n.SystemLocale)
	texts, err := newBackendTexts(language.Current)
	if err != nil {
		return nil, err
	}
	explainer, err := golangerrors.NewExplainer()
	if err != nil {
		return nil, err
	}

	// One supervisor for the whole IDE: a single program runs at a time, whatever its language.
	supervisor := process.New(sink)
	goRunner := runner.New(supervisor, runner.Options{
		Settings:         store,
		FirstBuildNotice: func() string { return texts.text("run.firstBuild") },
	})
	// Transitional (M0): N5 injects the package manager into the bridge; until then the bridge
	// still runs "go mod" through the app.Toolchain view of the runner.
	_ = packages.New(supervisor, goRunner)
	debugger, languageServer := newGoTools(sink, goRunner, texts)

	watcher, err := filewatch.New(sink, filewatch.DefaultDebounce)
	if err != nil {
		return nil, err
	}

	return &backend{
		services: []any{
			bridge.NewRunService(goRunner.Compat()),
			bridge.NewConsoleService(console.New(0)),
			bridge.NewDebugService(debugger),
			bridge.NewLanguageService(languageServer),
			bridge.NewAssistantService(sink, explainer, language),
			bridge.NewFilesServiceWithTexts(sink, store, watcher, texts.withData).UseShell(filesystem.New()),
			bridge.NewSettingsService(sink, store, language).
				UseTools(goRunner.Compat(), bridge.NewExecutableDialog(sink, texts.withData)),
		},
		shutdown: func(ctx context.Context) {
			_ = supervisor.Stop() // never leave the user's program running
			_ = debugger.Stop()
			_ = languageServer.Shutdown(ctx)
			_ = watcher.Close()
		},
	}, nil
}
