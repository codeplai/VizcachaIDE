package main

import (
	"context"
	_ "embed"
	"encoding/json"
	"log/slog"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"time"

	"github.com/codeplai/VizcachaIDE/wails/internal/adapters/filesystem"
	"github.com/codeplai/VizcachaIDE/wails/internal/adapters/updates"
	"github.com/codeplai/VizcachaIDE/wails/internal/app"
	"github.com/codeplai/VizcachaIDE/wails/internal/bridge"
)

// wailsConfig holds info.productVersion, the version the About dialog and the installers show.
//
//go:embed wails.json
var wailsConfig []byte

// updateURLVariable points the updater to another release list (QA uses a fake server).
const updateURLVariable = "VIZCACHA_UPDATE_URL"

// appVersion reads info.productVersion from the embedded wails.json.
func appVersion() string {
	var config struct {
		Info struct {
			ProductVersion string `json:"productVersion"`
		} `json:"info"`
	}
	if err := json.Unmarshal(wailsConfig, &config); err != nil {
		return "0.0.0"
	}
	return config.Info.ProductVersion
}

// newUpdater creates the updater of this installation (variant, system, installed or portable)
// and the function that checks once a day at start.
func newUpdater(sink *bridge.WailsEventSink, store app.SettingsStore) (*updates.Updater, func(context.Context)) {
	exeDir := executableDirectory()
	cacheDir, err := os.UserCacheDir()
	if err != nil {
		cacheDir = os.TempDir()
	}
	settings, _ := store.Load()
	updater := updates.New(updates.Options{
		Current:      appVersion(),
		APIURL:       os.Getenv(updateURLVariable),
		CacheDir:     filepath.Join(cacheDir, "VizcachaIDE", "updates"),
		Installation: updates.DetectInstallation(runtime.GOOS, runtime.GOARCH, exeDir),
		Sink:         sink,
		LastCheck:    settings.LastUpdateCheck,
		Launch:       startInstaller,
		Quit:         sink.Quit,
		Reveal:       filesystem.New().Reveal,
	})
	autoUpdate := func(ctx context.Context) {
		if err := app.AutoUpdate(ctx, updater, store, time.Now()); err != nil {
			slog.Warn("update check failed", "error", err) // the state already says it to the user
		}
	}
	return updater, autoUpdate
}

// startInstaller runs the downloaded installer on its own: the IDE closes right after.
func startInstaller(path string) error {
	cmd := exec.Command(path)
	cmd.Dir = filepath.Dir(path)
	if err := cmd.Start(); err != nil {
		return err
	}
	return cmd.Process.Release()
}

func executableDirectory() string {
	executable, err := os.Executable()
	if err != nil {
		return "."
	}
	return filepath.Dir(executable)
}
