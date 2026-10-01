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
	"github.com/codeplai/VizcachaIDE/wails/internal/adapters/gopls"
	"github.com/codeplai/VizcachaIDE/wails/internal/adapters/toolchain"
	"github.com/codeplai/VizcachaIDE/wails/internal/bridge"
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

func run() error {
	sink := bridge.NewWailsEventSink()
	settingsStore := bridge.NewMemorySettingsStore() // W1: adapters/settings

	goToolchain := toolchain.New(toolchain.Options{
		Sink:             sink,
		Settings:         settingsStore,
		FirstBuildNotice: firstBuildNotice(settingsStore),
	})
	explainer, err := errorcatalog.NewExplainer()
	if err != nil {
		return err
	}
	// Tool paths are resolved at start-up (configured -> bundled -> PATH); changing them in
	// Settings takes effect after a restart.
	debugger := delve.New(sink, delve.Options{
		DelvePath:   goToolchain.Locate(toolchain.ToolDelve).Path,
		Environment: goToolchain.Environment,
		Translate:   settingsTranslator(settingsStore),
	})
	languageServer := gopls.New(sink, gopls.Config{
		Executable:  goToolchain.Locate(toolchain.ToolGopls).Path,
		Environment: goToolchain.Environment,
	})

	services := []any{
		bridge.NewRunService(goToolchain),
		bridge.NewDebugService(debugger),
		bridge.NewLanguageService(languageServer),
		bridge.NewAssistantService(sink, explainer, settingsStore),
		bridge.NewFilesService(sink, settingsStore),
		bridge.NewSettingsService(sink, settingsStore),
	}

	return wails.Run(&options.App{
		Title:     "VizcachaIDE",
		Width:     1280,
		Height:    800,
		MinWidth:  1024,
		MinHeight: 640,
		AssetServer: &assetserver.Options{
			Assets: assets,
		},
		BackgroundColour: &options.RGBA{R: 245, G: 247, B: 249, A: 1},
		OnStartup:        func(ctx context.Context) { sink.SetContext(ctx) },
		OnShutdown: func(ctx context.Context) {
			_ = goToolchain.Stop() // never leave the user's program running
			_ = debugger.Stop()
			_ = languageServer.Shutdown(ctx)
		},
		Bind:    services,
		Windows: &windows.Options{Theme: windows.SystemDefault},
	})
}
