package filewatch

import (
	"crypto/sha256"
	"os"
	"path/filepath"
	"time"

	"github.com/fsnotify/fsnotify"
)

const reported = fsnotify.Write | fsnotify.Create | fsnotify.Rename

func (w *Watcher) loop() {
	for {
		select {
		case event, ok := <-w.fs.Events:
			if !ok {
				return
			}
			if event.Op&reported != 0 {
				w.schedule(filepath.Clean(event.Name))
			}
		case _, ok := <-w.fs.Errors:
			if !ok {
				return
			}
		}
	}
}

// schedule (re)starts the debounce timer of a watched file.
func (w *Watcher) schedule(path string) {
	w.mu.Lock()
	defer w.mu.Unlock()
	if _, watched := w.known[path]; !watched || w.closed {
		return
	}
	if timer, pending := w.timers[path]; pending {
		timer.Stop()
	}
	w.timers[path] = time.AfterFunc(w.debounce, func() { w.settle(path) })
}

// settle runs when a burst ended: it reports the file only if its content really differs
// from what the IDE saved or last reported.
func (w *Watcher) settle(path string) {
	data, err := os.ReadFile(path)
	if err != nil {
		return // being replaced or deleted: the next event tells
	}
	sum := sha256.Sum256(data)
	w.mu.Lock()
	delete(w.timers, path)
	previous, watched := w.known[path]
	if !watched || previous == sum || w.closed {
		w.mu.Unlock()
		return
	}
	w.known[path] = sum
	w.mu.Unlock()
	w.sink.FileChanged(path)
}
