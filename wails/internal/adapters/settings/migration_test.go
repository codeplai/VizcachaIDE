package settings

import (
	"os"
	"reflect"
	"strings"
	"testing"
)

const settings20 = `{
  "language": "es",
  "theme": "dark",
  "fontSize": 16,
  "goPath": "C:/Go/bin/go.exe",
  "delvePath": "C:/tools/dlv.exe",
  "goplsPath": "",
  "firstRun": false,
  "lastFolder": "C:/work",
  "formatOnSave": false,
  "recentFiles": ["C:/work/main.go"]
}`

func TestSettingsOf20MigrateToolPaths(t *testing.T) {
	store, path := newTestStore(t)
	writeRaw(t, path, settings20)

	got, err := store.Load()
	if err != nil {
		t.Fatal(err)
	}

	if want := map[string]string{"go": "C:/Go/bin/go.exe", "dlv": "C:/tools/dlv.exe"}; !reflect.DeepEqual(got.ToolPaths, want) {
		t.Errorf("ToolPaths = %v, want %v", got.ToolPaths, want)
	}
	if got.Language != "es" || got.FontSize != 16 || got.LastFolder != "C:/work" || got.FormatOnSave {
		t.Errorf("the rest of the settings changed: %+v", got)
	}
	if got.DefaultCodeLanguage != "go" {
		t.Errorf("DefaultCodeLanguage = %q, want the default", got.DefaultCodeLanguage)
	}

	saved, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	if text := string(saved); strings.Contains(text, "goPath") || strings.Contains(text, "delvePath") || !strings.Contains(text, "toolPaths") {
		t.Errorf("the new file must hold toolPaths and none of the old keys:\n%s", text)
	}
	again, err := store.Load()
	if err != nil || !reflect.DeepEqual(again, got) {
		t.Errorf("loading the migrated file = %+v, %v; want %+v", again, err, got)
	}
}

func TestMigrationKeepsPathsAlreadyInToolPaths(t *testing.T) {
	store, path := newTestStore(t)
	writeRaw(t, path, `{"goPath": "old/go", "delvePath": "old/dlv", "toolPaths": {"go": "new/go"}}`)

	got, err := store.Load()
	if err != nil {
		t.Fatal(err)
	}
	if want := map[string]string{"go": "new/go", "dlv": "old/dlv"}; !reflect.DeepEqual(got.ToolPaths, want) {
		t.Errorf("ToolPaths = %v, want %v", got.ToolPaths, want)
	}
}

func TestFilesWithoutLegacyKeysAreNotRewritten(t *testing.T) {
	store, path := newTestStore(t)
	const text = `{"language": "en"}`
	writeRaw(t, path, text)
	if _, err := store.Load(); err != nil {
		t.Fatal(err)
	}
	if saved, _ := os.ReadFile(path); string(saved) != text {
		t.Errorf("the file changed: %s", saved)
	}
}
