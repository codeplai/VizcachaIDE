// Package filewatch notices changes made to open files by other programs.
//
// It watches the folders of the files (editors often save by replacing the file, which
// a per-file watch would lose), filters by file name, merges bursts of events and
// ignores content the IDE itself just saved.
package filewatch

import (
	"crypto/sha256"
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"time"

	"github.com/codeplai/VizcachaIDE/wails/internal/app"
	"github.com/fsnotify/fsnotify"
)

// DefaultDebounce is how long the watcher waits for a burst of events to end.
const DefaultDebounce = 200 * time.Millisecond

// Watcher implements app.FileWatcher on top of fsnotify.
type Watcher struct {
	sink     app.FileChangeSink
	fs       *fsnotify.Watcher
	debounce time.Duration

	mu     sync.Mutex
	known  map[string][sha256.Size]byte // watched file -> last content seen or saved
	dirs   map[string]bool              // folders added to fsnotify
	timers map[string]*time.Timer
	closed bool
}

var _ app.FileWatcher = (*Watcher)(nil)

// New starts a watcher that reports changes to sink after the debounce time.
func New(sink app.FileChangeSink, debounce time.Duration) (*Watcher, error) {
	inner, err := fsnotify.NewWatcher()
	if err != nil {
		return nil, fmt.Errorf("start file watcher: %w", err)
	}
	w := &Watcher{
		sink:     sink,
		fs:       inner,
		debounce: debounce,
		known:    map[string][sha256.Size]byte{},
		dirs:     map[string]bool{},
		timers:   map[string]*time.Timer{},
	}
	go w.loop()
	return w, nil
}

// Watch implements app.FileWatcher.
func (w *Watcher) Watch(paths []string) error {
	wanted := map[string]bool{}
	for _, path := range paths {
		if path != "" {
			wanted[filepath.Clean(path)] = true
		}
	}
	w.mu.Lock()
	defer w.mu.Unlock()
	if w.closed {
		return nil
	}
	w.forgetUnwanted(wanted)
	var firstErr error
	for path := range wanted {
		if err := w.add(path); err != nil && firstErr == nil {
			firstErr = err
		}
	}
	return firstErr
}

// Remember implements app.FileWatcher.
func (w *Watcher) Remember(path, text string) {
	w.mu.Lock()
	defer w.mu.Unlock()
	path = filepath.Clean(path)
	if _, watched := w.known[path]; watched {
		w.known[path] = sha256.Sum256([]byte(text))
	}
}

// Close implements app.FileWatcher.
func (w *Watcher) Close() error {
	w.mu.Lock()
	w.closed = true
	for _, timer := range w.timers {
		timer.Stop()
	}
	w.mu.Unlock()
	return w.fs.Close()
}

// add starts watching a file (caller holds the lock). A file already watched keeps its content.
func (w *Watcher) add(path string) error {
	dir := filepath.Dir(path)
	if !w.dirs[dir] {
		if err := w.fs.Add(dir); err != nil {
			return fmt.Errorf("watch %s: %w", dir, err)
		}
		w.dirs[dir] = true
	}
	if _, watched := w.known[path]; !watched {
		w.known[path] = hashOf(path)
	}
	return nil
}

// forgetUnwanted stops watching what is no longer wanted (caller holds the lock).
func (w *Watcher) forgetUnwanted(wanted map[string]bool) {
	keep := map[string]bool{}
	for path := range w.known {
		if !wanted[path] {
			delete(w.known, path)
		}
	}
	for path := range wanted {
		keep[filepath.Dir(path)] = true
	}
	for dir := range w.dirs {
		if !keep[dir] {
			_ = w.fs.Remove(dir) // the folder may be gone already
			delete(w.dirs, dir)
		}
	}
}

func hashOf(path string) [sha256.Size]byte {
	data, err := os.ReadFile(path)
	if err != nil {
		return [sha256.Size]byte{}
	}
	return sha256.Sum256(data)
}
