package settings

import (
	"os"
	"path/filepath"
	"strings"
	"sync"
	"testing"

	"github.com/codeplai/VizcachaIDE/wails/internal/domain"
)

func newTestStore(t *testing.T) (*Store, string) {
	t.Helper()
	path := filepath.Join(t.TempDir(), "VizcachaIDE", "settings.json")
	return NewStore(path), path
}

func writeRaw(t *testing.T, path, text string) {
	t.Helper()
	if err := os.MkdirAll(filepath.Dir(path), 0o700); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path, []byte(text), 0o600); err != nil {
		t.Fatal(err)
	}
}

func TestMissingFileGivesDefaults(t *testing.T) {
	store, _ := newTestStore(t)
	got, err := store.Load()
	if err != nil || got != domain.DefaultSettings() {
		t.Fatalf("Load = %+v, %v", got, err)
	}
}

func TestSaveThenLoadRoundTrips(t *testing.T) {
	store, _ := newTestStore(t)
	want := domain.Settings{Language: "es", Theme: "dark", FontSize: 18, GoPath: "C:/go/bin/go.exe", LastFolder: "C:/x"}
	if err := store.Save(want); err != nil {
		t.Fatal(err)
	}
	got, err := store.Load()
	if err != nil || got != want {
		t.Fatalf("Load = %+v, %v; want %+v", got, err, want)
	}
}

func TestOldFilesTakeTheDefaultOfNewFields(t *testing.T) {
	store, path := newTestStore(t)
	writeRaw(t, path, `{"language":"es","fontSize":20}`)
	got, err := store.Load()
	if err != nil {
		t.Fatal(err)
	}
	if got.Language != "es" || got.FontSize != 20 || !got.FormatOnSave || got.Theme != domain.ThemeSystem || !got.FirstRun {
		t.Errorf("Load = %+v", got)
	}
}

func TestCorruptFileIsMovedToBackup(t *testing.T) {
	store, path := newTestStore(t)
	writeRaw(t, path, `{not json`)
	got, err := store.Load()
	if err != nil || got != domain.DefaultSettings() {
		t.Fatalf("Load = %+v, %v", got, err)
	}
	backup, err := os.ReadFile(path + ".bak")
	if err != nil || string(backup) != `{not json` {
		t.Errorf("backup = %q, %v", backup, err)
	}
	if _, err := os.Stat(path); !os.IsNotExist(err) {
		t.Errorf("the corrupt file must be gone, Stat err = %v", err)
	}
	if err := store.Save(got); err != nil {
		t.Errorf("Save after corruption: %v", err)
	}
}

func TestSaveLeavesNoTemporaryFiles(t *testing.T) {
	store, path := newTestStore(t)
	for range 3 {
		if err := store.Save(domain.DefaultSettings()); err != nil {
			t.Fatal(err)
		}
	}
	entries, err := os.ReadDir(filepath.Dir(path))
	if err != nil {
		t.Fatal(err)
	}
	for _, entry := range entries {
		if strings.HasSuffix(entry.Name(), ".tmp") {
			t.Errorf("leftover temporary file %s", entry.Name())
		}
	}
}

func TestFailedSaveKeepsTheOldFile(t *testing.T) {
	store, path := newTestStore(t)
	first := domain.DefaultSettings()
	first.FontSize = 16
	if err := store.Save(first); err != nil {
		t.Fatal(err)
	}
	// A regular file where the folder should be makes the save impossible.
	blocker := filepath.Join(t.TempDir(), "blocker")
	writeRaw(t, blocker, "file")
	if err := NewStore(filepath.Join(blocker, "settings.json")).Save(first); err == nil {
		t.Error("Save into an impossible folder must fail")
	}
	got, err := store.Load()
	if err != nil || got != first {
		t.Errorf("Load = %+v, %v (file %s)", got, err, path)
	}
}

func TestConcurrentSavesAndLoads(t *testing.T) {
	store, _ := newTestStore(t)
	var wg sync.WaitGroup
	for i := range 20 {
		wg.Add(2)
		go func() {
			defer wg.Done()
			settings := domain.DefaultSettings()
			settings.FontSize = 10 + i
			if err := store.Save(settings); err != nil {
				t.Error(err)
			}
		}()
		go func() {
			defer wg.Done()
			if _, err := store.Load(); err != nil {
				t.Error(err)
			}
		}()
	}
	wg.Wait()
	if got, err := store.Load(); err != nil || got.FontSize < 10 || got.FontSize > 29 {
		t.Errorf("Load = %+v, %v", got, err)
	}
}
