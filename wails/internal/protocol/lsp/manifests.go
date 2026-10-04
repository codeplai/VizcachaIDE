package lsp

import (
	"context"
	"os"
	"path/filepath"
	"sync"
	"time"

	"go.lsp.dev/protocol"
)

// ManifestFiles is implemented by a Flavor whose server must hear when project files that are
// not open in the editor change: go.mod and go.sum after "go get", Cargo.toml after "cargo add".
// Without that gopls kept saying "could not import github.com/google/uuid" after the package was
// added (found in the multi-file experiment after M3).
type ManifestFiles interface {
	// Manifests are the file names to watch in the root of each open project.
	Manifests() []string
}

// manifestCheckEvery throttles the checks: they ride on the queries the editor makes anyway.
const manifestCheckEvery = 2 * time.Second

// manifestWatch remembers the modification time of the manifests of the open projects.
type manifestWatch struct {
	mu      sync.Mutex
	checked time.Time
	seen    map[string]time.Time // path -> modification time (zero: the file does not exist)
}

func manifestsOf(flavor Flavor) []string {
	if files, ok := flavor.(ManifestFiles); ok {
		return files.Manifests()
	}
	return nil
}

// manifestPaths are the manifests of the roots of the open documents and of root.
func (s *Server) manifestPaths(root string) []string {
	names := manifestsOf(s.flavor)
	if len(names) == 0 {
		return nil
	}
	roots := map[string]bool{}
	if root != "" {
		roots[root] = true
	}
	for _, doc := range s.docs.all() {
		roots[s.flavor.RootOf(doc.path)] = true
	}
	var paths []string
	for folder := range roots {
		for _, name := range names {
			paths = append(paths, filepath.Join(folder, name))
		}
	}
	return paths
}

// rememberManifests takes the state the server is about to read when it starts.
func (s *Server) rememberManifests(root string) {
	w := &s.manifests
	w.mu.Lock()
	defer w.mu.Unlock()
	w.seen = map[string]time.Time{}
	for _, path := range s.manifestPaths(root) {
		w.seen[path] = modified(path)
	}
	w.checked = time.Now()
}

// noticeManifests tells the server which manifests changed since it last heard, at most every
// manifestCheckEvery. A manifest of a project opened later is only remembered the first time.
func (s *Server) noticeManifests(ctx context.Context, conn *connection) {
	w := &s.manifests
	w.mu.Lock()
	if time.Since(w.checked) < manifestCheckEvery || w.seen == nil {
		w.mu.Unlock()
		return
	}
	w.checked = time.Now()
	var changes []*protocol.FileEvent
	for _, path := range s.manifestPaths("") {
		now, before := modified(path), w.seen[path]
		_, known := w.seen[path]
		w.seen[path] = now
		if known && !now.Equal(before) {
			changes = append(changes, &protocol.FileEvent{URI: pathToURI(path), Type: changeType(before, now)})
		}
	}
	w.mu.Unlock()
	if len(changes) > 0 {
		_ = conn.notify(ctx, "workspace/didChangeWatchedFiles", protocol.DidChangeWatchedFilesParams{Changes: changes})
	}
}

func changeType(before, now time.Time) protocol.FileChangeType {
	switch {
	case before.IsZero():
		return protocol.FileChangeTypeCreated
	case now.IsZero():
		return protocol.FileChangeTypeDeleted
	}
	return protocol.FileChangeTypeChanged
}

func modified(path string) time.Time {
	info, err := os.Stat(path)
	if err != nil {
		return time.Time{}
	}
	return info.ModTime()
}
