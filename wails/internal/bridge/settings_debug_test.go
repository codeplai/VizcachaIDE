package bridge

import (
	"testing"

	"github.com/codeplai/VizcachaIDE/wails/internal/domain"
)

func newPickService(t *testing.T, store *MemorySettingsStore, pick ExecutablePicker) (*SettingsService, *recordingSink) {
	t.Helper()
	sink := &recordingSink{}
	service := NewSettingsService(sink, store, NewLanguageResolver(store, nil), newTestRegistry(t)).UseTools(pick)
	return service, sink
}

func TestPickExecutableSavesRedetectsAndNotifies(t *testing.T) {
	store := NewMemorySettingsStore()
	service, sink := newPickService(t, store, func(tool string) (string, error) { return "/tools/" + tool, nil })

	statuses, err := service.PickExecutable("dlv")
	if err != nil {
		t.Fatal(err)
	}

	if saved, _ := store.Load(); saved.ToolPaths["dlv"] != "/tools/dlv" {
		t.Errorf("ToolPaths = %v", saved.ToolPaths)
	}
	if len(statuses) != 2 || statuses[0].ID != "go" || statuses[1].ID != "python" {
		t.Errorf("statuses = %+v, want the tools of every language", statuses)
	}
	if len(sink.calls) != 1 || sink.calls[0] != "settings" {
		t.Errorf("calls = %v, want [settings]", sink.calls)
	}
}

func TestPickExecutableCancelledChangesNothing(t *testing.T) {
	store := NewMemorySettingsStore()
	service, sink := newPickService(t, store, func(string) (string, error) { return "", nil })

	if _, err := service.PickExecutable("python"); err != nil {
		t.Fatal(err)
	}
	if len(sink.calls) != 0 {
		t.Errorf("calls = %v, want none", sink.calls)
	}
}

func TestPickExecutableRejectsUnknownAndProvidedTools(t *testing.T) {
	asked := false
	service, _ := newPickService(t, NewMemorySettingsStore(), func(string) (string, error) {
		asked = true
		return "/x", nil
	})
	for _, tool := range []string{"rm", "", "debugpy"} {
		if _, err := service.PickExecutable(tool); err == nil {
			t.Errorf("tool %q must be rejected", tool)
		}
	}
	if asked {
		t.Error("the dialog must not open for a rejected tool")
	}
}

func TestResolvedLanguageFollowsTheSetting(t *testing.T) {
	store := NewMemorySettingsStore()
	service := NewSettingsService(&recordingSink{}, store, NewLanguageResolver(store, func() string { return "es-PE" }), newTestRegistry(t))
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
