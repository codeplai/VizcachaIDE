package main

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/codeplai/VizcachaIDE/wails/internal/adapters/settings"
	"github.com/codeplai/VizcachaIDE/wails/internal/bridge"
	"github.com/codeplai/VizcachaIDE/wails/internal/domain"
)

// Every language can create its starting project, and the file it says to open exists.
func TestEveryLanguageCreatesItsStartingProject(t *testing.T) {
	texts, err := newBackendTexts(func() string { return domain.LanguageEN })
	if err != nil {
		t.Fatal(err)
	}
	registry, _, err := newRegistry(bridge.NewWailsEventSink(), settings.NewStore(filepath.Join(t.TempDir(), "settings.json")), texts)
	if err != nil {
		t.Fatal(err)
	}
	service := bridge.NewProjectsService(registry)
	location := t.TempDir()
	for _, profile := range registry.Profiles() {
		project, err := service.Create(profile.ID, location, "Hola "+string(profile.ID))
		if err != nil {
			t.Fatalf("%s: %v", profile.ID, err)
		}
		if _, err := os.Stat(project.MainFile); err != nil {
			t.Errorf("%s: main file: %v", profile.ID, err)
		}
	}
}
