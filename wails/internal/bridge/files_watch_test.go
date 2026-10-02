package bridge

import (
	"path/filepath"
	"reflect"
	"testing"
)

type recordingWatcher struct {
	watched  []string
	remember [][2]string
}

func (r *recordingWatcher) Watch(paths []string) error { r.watched = paths; return nil }
func (r *recordingWatcher) Remember(path, text string) {
	r.remember = append(r.remember, [2]string{path, text})
}
func (r *recordingWatcher) Close() error { return nil }

func TestSaveFileTellsTheWatcherAndWatchFilesForwards(t *testing.T) {
	watcher := &recordingWatcher{}
	service := NewFilesService(noContext{}, nil, watcher)
	path := filepath.Join(t.TempDir(), "main.go")
	if err := service.SaveFile(path, "package main\n"); err != nil {
		t.Fatal(err)
	}
	if want := [][2]string{{path, "package main\n"}}; !reflect.DeepEqual(watcher.remember, want) {
		t.Errorf("remembered %v, want %v", watcher.remember, want)
	}
	if err := service.WatchFiles([]string{path}); err != nil || len(watcher.watched) != 1 {
		t.Errorf("WatchFiles = %v, %v", watcher.watched, err)
	}
}
