package bridge

import (
	"os"
	"path/filepath"
	"reflect"
	"testing"

	"github.com/codeplai/VizcachaIDE/wails/internal/domain"
)

func TestApplyTextEditsWritesFilesAndTellsTheWatcher(t *testing.T) {
	watcher := &recordingWatcher{}
	service := NewFilesService(noContext{}, nil, watcher)
	path := filepath.Join(t.TempDir(), "util.go")
	if err := os.WriteFile(path, []byte("func twice() {}\n"), 0o600); err != nil {
		t.Fatal(err)
	}
	edit := domain.FileEdit{File: path, Edits: []domain.TextEdit{{
		Range: domain.SourceRange{
			Start: domain.SourceLocation{Line: 1, Column: 6}, End: domain.SourceLocation{Line: 1, Column: 11},
		},
		NewText: "double",
	}}}
	summary, err := service.ApplyTextEdits([]domain.FileEdit{edit})
	if err != nil || summary.Files != 1 || summary.Edits != 1 {
		t.Fatalf("summary %+v, %v", summary, err)
	}
	if onDisk, _ := os.ReadFile(path); string(onDisk) != "func double() {}\n" {
		t.Errorf("file = %q", onDisk)
	}
	if want := [][2]string{{path, "func double() {}\n"}}; !reflect.DeepEqual(watcher.remember, want) {
		t.Errorf("remembered %v, want %v", watcher.remember, want)
	}
}
