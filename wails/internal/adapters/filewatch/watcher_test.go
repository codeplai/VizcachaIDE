package filewatch

import (
	"os"
	"path/filepath"
	"sync"
	"testing"
	"time"
)

type recorder struct {
	mu    sync.Mutex
	paths []string
}

func (r *recorder) FileChanged(path string) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.paths = append(r.paths, path)
}

func (r *recorder) count() int {
	r.mu.Lock()
	defer r.mu.Unlock()
	return len(r.paths)
}

const testDebounce = 50 * time.Millisecond

func setup(t *testing.T, content string) (*Watcher, *recorder, string) {
	t.Helper()
	path := filepath.Join(t.TempDir(), "main.go")
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}
	sink := &recorder{}
	w, err := New(sink, testDebounce)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = w.Close() })
	if err := w.Watch([]string{path}); err != nil {
		t.Fatal(err)
	}
	return w, sink, path
}

func waitFor(t *testing.T, sink *recorder, want int) {
	t.Helper()
	deadline := time.Now().Add(3 * time.Second)
	for time.Now().Before(deadline) {
		if sink.count() >= want {
			return
		}
		time.Sleep(10 * time.Millisecond)
	}
	t.Fatalf("got %d events, want %d", sink.count(), want)
}

func TestExternalWriteIsReported(t *testing.T) {
	_, sink, path := setup(t, "one")
	if err := os.WriteFile(path, []byte("two"), 0o644); err != nil {
		t.Fatal(err)
	}
	waitFor(t, sink, 1)
	if sink.paths[0] != path {
		t.Errorf("path = %q, want %q", sink.paths[0], path)
	}
}

func TestBurstIsMergedIntoOneEvent(t *testing.T) {
	_, sink, path := setup(t, "one")
	for _, text := range []string{"a", "ab", "abc"} {
		if err := os.WriteFile(path, []byte(text), 0o644); err != nil {
			t.Fatal(err)
		}
	}
	waitFor(t, sink, 1)
	time.Sleep(4 * testDebounce)
	if got := sink.count(); got != 1 {
		t.Errorf("events = %d, want 1", got)
	}
}

func TestOwnSaveIsIgnored(t *testing.T) {
	w, sink, path := setup(t, "one")
	if err := os.WriteFile(path, []byte("mine"), 0o644); err != nil {
		t.Fatal(err)
	}
	w.Remember(path, "mine")
	time.Sleep(6 * testDebounce)
	if got := sink.count(); got != 0 {
		t.Fatalf("own save reported %d times", got)
	}
	if err := os.WriteFile(path, []byte("theirs"), 0o644); err != nil {
		t.Fatal(err)
	}
	waitFor(t, sink, 1)
}

func TestReplaceByRenameIsReported(t *testing.T) {
	_, sink, path := setup(t, "one")
	temp := path + ".tmp"
	if err := os.WriteFile(temp, []byte("replaced"), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.Rename(temp, path); err != nil {
		t.Fatal(err)
	}
	waitFor(t, sink, 1)
}

func TestOtherFilesAndUnwatchedFilesAreIgnored(t *testing.T) {
	w, sink, path := setup(t, "one")
	other := filepath.Join(filepath.Dir(path), "other.go")
	if err := os.WriteFile(other, []byte("x"), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := w.Watch(nil); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path, []byte("two"), 0o644); err != nil {
		t.Fatal(err)
	}
	time.Sleep(6 * testDebounce)
	if got := sink.count(); got != 0 {
		t.Errorf("events = %d, want 0", got)
	}
}
