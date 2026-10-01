package bridge

import (
	"context"
	"os"
	"path/filepath"
	"reflect"
	"testing"

	"github.com/codeplai/VizcachaIDE/wails/internal/app"
	"github.com/codeplai/VizcachaIDE/wails/internal/domain"
)

type startRecordingDebugger struct {
	app.Debugger
	config      domain.RunConfiguration
	breakpoints []domain.Breakpoint
}

func (d *startRecordingDebugger) Start(_ context.Context, c domain.RunConfiguration, b []domain.Breakpoint) error {
	d.config, d.breakpoints = c, b
	return nil
}

func TestDebugStartUsesTheRunConfigurationOfRun(t *testing.T) {
	dir := t.TempDir()
	if err := os.WriteFile(filepath.Join(dir, "go.mod"), []byte("module demo\n"), 0o600); err != nil {
		t.Fatal(err)
	}
	debugger := &startRecordingDebugger{}
	breakpoints := []domain.Breakpoint{{Condition: "i > 2"}}

	err := NewDebugService(debugger).Start(filepath.Join(dir, "main.go"), breakpoints, `-n 3 "two words"`)
	if err != nil {
		t.Fatal(err)
	}

	if debugger.config.Mode != domain.RunPackage || debugger.config.Module == nil {
		t.Errorf("config = %+v, want a package run inside the module", debugger.config)
	}
	if want := []string{"-n", "3", "two words"}; !reflect.DeepEqual(debugger.config.ProgramArgs, want) {
		t.Errorf("args = %v, want %v", debugger.config.ProgramArgs, want)
	}
	if len(debugger.breakpoints) != 1 {
		t.Errorf("breakpoints = %v", debugger.breakpoints)
	}
}

func TestDebugStartRejectsUnclosedQuotes(t *testing.T) {
	debugger := &startRecordingDebugger{}
	if err := NewDebugService(debugger).Start("main.go", nil, `"oops`); err == nil {
		t.Error("an unclosed quote must be an error")
	}
}

type infoToolchain struct {
	app.Toolchain
	store *MemorySettingsStore
}

func (f infoToolchain) Info(context.Context) domain.ToolchainInfo {
	settings, _ := f.store.Load()
	return domain.ToolchainInfo{DelveVersion: settings.DelvePath}
}

func TestPickExecutableSavesRedetectsAndNotifies(t *testing.T) {
	store := NewMemorySettingsStore()
	sink := &recordingSink{}
	service := NewSettingsService(sink, store, NewLanguageResolver(store, nil)).UseTools(infoToolchain{store: store},
		func(tool string) (string, error) { return "/tools/" + tool, nil })

	info, err := service.PickExecutable("dlv")
	if err != nil {
		t.Fatal(err)
	}

	if saved, _ := store.Load(); saved.DelvePath != "/tools/dlv" {
		t.Errorf("DelvePath = %q", saved.DelvePath)
	}
	if info.DelveVersion != "/tools/dlv" {
		t.Errorf("info = %+v, want the tools detected after the change", info)
	}
	if len(sink.calls) != 1 || sink.calls[0] != "settings" {
		t.Errorf("calls = %v, want [settings]", sink.calls)
	}
}

func TestPickExecutableCancelledChangesNothing(t *testing.T) {
	store := NewMemorySettingsStore()
	sink := &recordingSink{}
	service := NewSettingsService(sink, store, NewLanguageResolver(store, nil)).UseTools(infoToolchain{store: store},
		func(string) (string, error) { return "", nil })

	if _, err := service.PickExecutable("gopls"); err != nil {
		t.Fatal(err)
	}
	if len(sink.calls) != 0 {
		t.Errorf("calls = %v, want none", sink.calls)
	}
	if _, err := service.PickExecutable("rm"); err == nil {
		t.Error("an unknown tool must be rejected")
	}
}

func TestResolvedLanguageFollowsTheSetting(t *testing.T) {
	store := NewMemorySettingsStore()
	service := NewSettingsService(&recordingSink{}, store, NewLanguageResolver(store, func() string { return "es-PE" }))
	for setting, want := range map[string]string{"es": "es", "en": "en"} {
		if err := store.Save(domain.Settings{Language: setting}); err != nil {
			t.Fatal(err)
		}
		if got := service.ResolvedLanguage(); got != want {
			t.Errorf("ResolvedLanguage with %q = %q, want %q", setting, got, want)
		}
	}
	if err := store.Save(domain.Settings{Language: domain.LanguageAuto}); err != nil {
		t.Fatal(err)
	}
	if got := service.ResolvedLanguage(); got != "es" {
		t.Errorf("auto with a Spanish system resolved to %q", got)
	}
}
