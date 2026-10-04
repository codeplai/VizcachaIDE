// Command vizcacha is the VizcachaIDE desktop app: a Go IDE for beginners.
//
// This file is the composition root: the only place that creates adapters,
// services and the Wails application.
package main

import (
	"context"
	"embed"
	"log"

	"github.com/codeplai/VizcachaIDE/wails/internal/adapters/filesystem"
	"github.com/codeplai/VizcachaIDE/wails/internal/adapters/filewatch"
	"github.com/codeplai/VizcachaIDE/wails/internal/adapters/settings"
	"github.com/codeplai/VizcachaIDE/wails/internal/adapters/windowstate"
	"github.com/codeplai/VizcachaIDE/wails/internal/app"
	"github.com/codeplai/VizcachaIDE/wails/internal/bridge"
	"github.com/codeplai/VizcachaIDE/wails/internal/domain"
	"github.com/codeplai/VizcachaIDE/wails/internal/i18n"
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

// newRegistry creates the language supports and the registry that maps files to them. It
// fails at start-up when a profile and its ports disagree, never in the middle of a lesson.
// The returned function releases what the supports hold beyond what shutdownLanguages stops.
func newRegistry(sink *bridge.WailsEventSink, store app.SettingsStore, texts *backendTexts) (*app.LanguageRegistry, func(context.Context), error) {
	// One supervisor for the whole IDE: a single program runs at a time, whatever its language.
	supervisor := process.New(sink)
	goSupport, closeGo, err := newGoSupport(sink, store, texts, supervisor)
	if err != nil {
		return nil, nil, err
	}
	pythonSupport, closePython, err := newPythonSupport(sink, store, texts, supervisor)
	if err != nil {
		return nil, nil, err
	}
	supports := append([]app.LanguageSupport{goSupport, pythonSupport}, newUnavailableSupports()...)
	registry, err := app.NewLanguageRegistry(domain.CodeLanguageGo, supports...)
	if err != nil {
		return nil, nil, err
	}
	closeAll := func(ctx context.Context) {
		closeGo(ctx)
		closePython(ctx)
	}
	return registry, closeAll, nil
}

// shutdownLanguages stops what every language may have running: the user's program, a debug
// session and the language server.
func shutdownLanguages(ctx context.Context, registry *app.LanguageRegistry) {
	for _, support := range registry.All() {
		_ = support.Runner.Stop() // never leave the user's program running
		_ = support.Debugger.Stop()
		_ = support.LanguageServer.Shutdown(ctx)
	}
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
	registry, closeSupports, err := newRegistry(sink, store, texts)
	if err != nil {
		return nil, err
	}
	watcher, err := filewatch.New(sink, filewatch.DefaultDebounce)
	if err != nil {
		return nil, err
	}

	return &backend{
		services: []any{
			bridge.NewRunService(registry),
			bridge.NewPackagesService(registry),
			bridge.NewCodeLanguagesService(registry),
			bridge.NewConsoleService(registry),
			bridge.NewDebugService(registry),
			bridge.NewLanguageService(registry),
			bridge.NewAssistantService(sink, registry, language),
			bridge.NewFilesServiceWithTexts(sink, store, watcher, texts.withData).
				UseShell(filesystem.New()).UseLanguages(registry),
			bridge.NewSettingsService(sink, store, language, registry).
				UseTools(bridge.NewExecutableDialog(sink, texts.withData)),
		},
		shutdown: func(ctx context.Context) {
			shutdownLanguages(ctx, registry)
			closeSupports(ctx)
			_ = watcher.Close()
		},
	}, nil
}
