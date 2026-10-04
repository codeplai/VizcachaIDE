package main

import (
	"path/filepath"
	"testing"

	"github.com/codeplai/VizcachaIDE/wails/internal/adapters/settings"
	"github.com/codeplai/VizcachaIDE/wails/internal/bridge"
	"github.com/codeplai/VizcachaIDE/wails/internal/domain"
)

// The start-up check: the Go, Python, C++ and Rust supports must form a consistent registry, with every
// extension routed to its language.
func TestRegistryOfTheApplicationIsConsistent(t *testing.T) {
	texts, err := newBackendTexts(func() string { return domain.LanguageEN })
	if err != nil {
		t.Fatal(err)
	}
	registry, _, err := newRegistry(bridge.NewWailsEventSink(), settings.NewStore(filepath.Join(t.TempDir(), "settings.json")), texts)
	if err != nil {
		t.Fatal(err)
	}
	for path, want := range map[string]domain.CodeLanguage{
		"main.go": domain.CodeLanguageGo, "tool.PY": domain.CodeLanguagePython, "a.hpp": domain.CodeLanguageCpp,
		"main.rs": domain.CodeLanguageRust,
	} {
		support, ok := registry.ForPath(path)
		if !ok || support.Profile.ID != want {
			t.Errorf("ForPath(%q) = %q, %v; want %q", path, support.Profile.ID, ok, want)
		}
	}
	if registry.Default().Profile.ID != domain.CodeLanguageGo || len(registry.Profiles()) != 4 {
		t.Errorf("default = %q, profiles = %d", registry.Default().Profile.ID, len(registry.Profiles()))
	}
}
