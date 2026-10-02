// Command vizcacha is the VizcachaIDE desktop app: a Go IDE for beginners.
//
// This file is the composition root: the only place that creates adapters,
// services and the Wails application.
package main

import (
	"context"
	"embed"
	"log"

	"github.com/codeplai/VizcachaIDE/wails/internal/adapters/delve"
	"github.com/codeplai/VizcachaIDE/wails/internal/adapters/errorcatalog"
	"github.com/codeplai/VizcachaIDE/wails/internal/adapters/filewatch"
	"github.com/codeplai/VizcachaIDE/wails/internal/adapters/gopls"
	"github.com/codeplai/VizcachaIDE/wails/internal/adapters/settings"
	"github.com/codeplai/VizcachaIDE/wails/internal/adapters/toolchain"
	"github.com/codeplai/VizcachaIDE/wails/internal/adapters/windowstate"
	"github.com/codeplai/VizcachaIDE/wails/internal/bridge"
	"github.com/codeplai/VizcachaIDE/wails/internal/domain"
	"github.com/codeplai/VizcachaIDE/wails/internal/i18n"
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
	explainer, err := errorcatalog.NewExplainer()
	if err != nil {
		return nil, err
	}

	goToolchain := toolchain.New(toolchain.Options{
		Sink:             sink,
		Settings:         store,
		FirstBuildNotice: func() string { return texts.text("run.firstBuild") },
	})
	// Tool paths are resolved when a debug session or gopls starts (configured ->
	// bundled -> PATH), so changing them in Settings needs no restart.
	debugger := delve.New(sink, delve.Options{
		DelvePath:   func() string { return goToolchain.Locate(toolchain.ToolDelve).Path },
		Environment: goToolchain.Environment,
		Translate:   texts.text,
	})
	languageServer := gopls.New(sink, gopls.Config{
		Executable:  func() string { return goToolchain.Locate(toolchain.ToolGopls).Path },
		Environment: goToolchain.Environment,
	})

	watcher, err := filewatch.New(sink, filewatch.DefaultDebounce)
	if err != nil {
		return nil, err
	}

	return &backend{
		services: []any{
			bridge.NewRunService(goToolchain),
			bridge.NewDebugService(debugger),
			bridge.NewLanguageService(languageServer),
			bridge.NewAssistantService(sink, explainer, language),
			bridge.NewFilesService(sink, store, watcher),
			bridge.NewSettingsService(sink, store, language).
				UseTools(goToolchain, bridge.NewExecutableDialog(sink, texts.withData)),
		},
		shutdown: func(ctx context.Context) {
			_ = goToolchain.Stop() // never leave the user's program running
			_ = debugger.Stop()
			_ = languageServer.Shutdown(ctx)
			_ = watcher.Close()
		},
	}, nil
}
