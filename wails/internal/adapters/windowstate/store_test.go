package windowstate

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/codeplai/VizcachaIDE/wails/internal/domain"
)

func TestRoundTrip(t *testing.T) {
	store := NewStore(filepath.Join(t.TempDir(), "sub", "window.json"))
	want := domain.WindowState{Width: 1400, Height: 900, X: -20, Y: 30, Maximised: true, HasPosition: true}
	if err := store.Save(want); err != nil {
		t.Fatal(err)
	}
	got, ok := store.Load()
	if !ok || got != want {
		t.Fatalf("got %+v ok=%v, want %+v", got, ok, want)
	}
}

func TestMissingFileGivesDefaults(t *testing.T) {
	got, ok := NewStore(filepath.Join(t.TempDir(), "window.json")).Load()
	if ok || got != domain.DefaultWindowState() {
		t.Fatalf("got %+v ok=%v", got, ok)
	}
}

func TestCorruptFileGivesDefaults(t *testing.T) {
	path := filepath.Join(t.TempDir(), "window.json")
	if err := os.WriteFile(path, []byte("{not json"), 0o600); err != nil {
		t.Fatal(err)
	}
	got, ok := NewStore(path).Load()
	if ok || got != domain.DefaultWindowState() {
		t.Fatalf("got %+v ok=%v", got, ok)
	}
}

func TestSaveLeavesNoTemporaryFiles(t *testing.T) {
	dir := t.TempDir()
	store := NewStore(filepath.Join(dir, "window.json"))
	for range 3 {
		if err := store.Save(domain.DefaultWindowState()); err != nil {
			t.Fatal(err)
		}
	}
	entries, err := os.ReadDir(dir)
	if err != nil || len(entries) != 1 {
		t.Fatalf("entries=%v err=%v", entries, err)
	}
}
