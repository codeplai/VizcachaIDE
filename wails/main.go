// Command vizcacha is the VizcachaIDE desktop app: a Go IDE for beginners.
//
// This file is the composition root: the only place that creates adapters,
// services and the Wails application.
package main

import (
	"context"
	"embed"
	"log"

	"github.com/codeplai/VizcachaIDE/wails/internal/adapters/gopls"
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

	// W1 G3: pass gopls.Config{Environment: toolchain.Environment} once the toolchain adapter exists.
	languageServer := gopls.New(sink, gopls.Config{})

	services := []any{
		bridge.NewRunService(sink),
		bridge.NewDebugService(sink),
		bridge.NewLanguageService(languageServer),
		bridge.NewAssistantService(sink, settingsStore),
		bridge.NewFilesService(),
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
		OnShutdown:       func(ctx context.Context) { _ = languageServer.Shutdown(ctx) },
		Bind:             services,
		Windows:          &windows.Options{Theme: windows.SystemDefault},
	})
}
