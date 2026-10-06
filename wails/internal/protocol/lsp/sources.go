package lsp

import (
	"context"
	"io/fs"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"go.lsp.dev/protocol"
)

// Source files of a project that are not open in the editor can change under the server's feet:
// another program saves them, git checks out a branch, a rename edits them on disk. A server
// only re-reads a closed file when told (workspace/didChangeWatchedFiles). Like the manifests,
// they are checked when a query is made, at most every sourceCheckEvery, and only the files of
// the language (Options.SourceExtensions) under the root of an open document.
const (
	sourceCheckEvery = 3 * time.Second
	maxSourceFiles   = 4000
)

// skippedFolders are never searched for sources (dependencies, build output).
var skippedFolders = map[string]bool{
	"node_modules": true, "vendor": true, "target": true, "build": true, "dist": true, "__pycache__": true,
}

type sourceWatch struct {
	mu      sync.Mutex
	checked time.Time
	seen    map[string]time.Time // path -> modification time
	roots   map[string]bool      // roots already scanned once
}

// scanSources lists the modification time of the source files under root.
func scanSources(root string, extensions []string) map[string]time.Time {
	found := map[string]time.Time{}
	_ = filepath.WalkDir(root, func(path string, entry fs.DirEntry, err error) error {
		if err != nil {
			return nil //nolint:nilerr // an unreadable folder is skipped
		}
		name := entry.Name()
		if entry.IsDir() {
			if path != root && (strings.HasPrefix(name, ".") || skippedFolders[name]) {
				return filepath.SkipDir
			}
			return nil
		}
		if !hasExtension(name, extensions) {
			return nil
		}
		if info, infoErr := entry.Info(); infoErr == nil {
			found[path] = info.ModTime()
		}
		if len(found) >= maxSourceFiles {
			return filepath.SkipAll
		}
		return nil
	})
	return found
}

func hasExtension(name string, extensions []string) bool {
	for _, extension := range extensions {
		if strings.HasSuffix(strings.ToLower(name), extension) {
			return true
		}
	}
	return false
}

// sourceRoots are the project roots of the open documents.
func (s *Server) sourceRoots() []string {
	seen := map[string]bool{}
	var roots []string
	for _, doc := range s.docs.all() {
		root := s.flavor.RootOf(doc.path)
		if !seen[pathKey(root)] {
			seen[pathKey(root)] = true
			roots = append(roots, root)
		}
	}
	return roots
}

// noticeSources tells the server which source files (not open in the editor) were created,
// changed or deleted since it last heard.
func (s *Server) noticeSources(ctx context.Context, conn *connection) {
	extensions := s.opts.SourceExtensions
	w := &s.sources
	w.mu.Lock()
	if len(extensions) == 0 || time.Since(w.checked) < sourceCheckEvery {
		w.mu.Unlock()
		return
	}
	w.checked = time.Now()
	if w.seen == nil {
		w.seen, w.roots = map[string]time.Time{}, map[string]bool{}
	}
	var changes []*protocol.FileEvent
	for _, root := range s.sourceRoots() {
		now := scanSources(root, extensions)
		changes = append(changes, w.diff(root, now, s.docs.get)...)
	}
	w.mu.Unlock()
	if len(changes) > 0 {
		_ = conn.notify(ctx, "workspace/didChangeWatchedFiles", protocol.DidChangeWatchedFilesParams{Changes: changes})
	}
}

// diff compares a scan of root with what was seen (caller holds the lock). The first scan of a
// root only remembers. Open documents are skipped: the editor sends their text itself.
func (w *sourceWatch) diff(root string, now map[string]time.Time, isOpen func(string) (document, bool)) []*protocol.FileEvent {
	first := !w.roots[pathKey(root)]
	w.roots[pathKey(root)] = true
	var changes []*protocol.FileEvent
	prefix := pathKey(root)
	for path := range w.seen {
		if _, there := now[path]; !there && strings.HasPrefix(pathKey(path), prefix) {
			delete(w.seen, path)
			if _, open := isOpen(path); !open {
				changes = append(changes, &protocol.FileEvent{URI: pathToURI(path), Type: protocol.FileChangeTypeDeleted})
			}
		}
	}
	for path, modified := range now {
		before, known := w.seen[path]
		w.seen[path] = modified
		if first || (known && modified.Equal(before)) {
			continue
		}
		if _, open := isOpen(path); open {
			continue
		}
		kind := protocol.FileChangeTypeChanged
		if !known {
			kind = protocol.FileChangeTypeCreated
		}
		changes = append(changes, &protocol.FileEvent{URI: pathToURI(path), Type: kind})
	}
	return changes
}

// rememberSources takes the state of the project the server is about to read when it starts.
func (s *Server) rememberSources(root string) {
	extensions := s.opts.SourceExtensions
	if len(extensions) == 0 {
		return
	}
	now := scanSources(root, extensions)
	w := &s.sources
	w.mu.Lock()
	defer w.mu.Unlock()
	w.seen, w.roots = now, map[string]bool{pathKey(root): true}
	w.checked = time.Now()
}
