package lsp

import (
	"os"
	"sync"

	"github.com/codeplai/VizcachaIDE/wails/internal/domain"
)

// document is a file the editor has open.
type document struct {
	path        string
	text        string
	version     int32
	diagnostics []domain.Diagnostic // the last ones the server published; nil before that
}

// openDocuments is the registry of open documents, safe for concurrent use.
type openDocuments struct {
	mu    sync.Mutex
	byKey map[string]*document
}

func newOpenDocuments() *openDocuments {
	return &openDocuments{byKey: map[string]*document{}}
}

// open registers a document at version 1. A document that is already open (the window
// reloaded, the tab was opened again) keeps its version and diagnostics: reopened is true, and
// changed says whether the text differs (then it gets the next version).
func (d *openDocuments) open(path, text string) (doc document, reopened, changed bool) {
	d.mu.Lock()
	defer d.mu.Unlock()
	current, found := d.byKey[pathKey(path)]
	if !found {
		stored := &document{path: path, text: text, version: 1}
		d.byKey[pathKey(path)] = stored
		return *stored, false, false
	}
	if current.text != text {
		current.text = text
		current.version++
		return *current, true, true
	}
	return *current, true, false
}

// remember keeps the diagnostics the server published for an open document.
func (d *openDocuments) remember(path string, diagnostics []domain.Diagnostic) {
	d.mu.Lock()
	defer d.mu.Unlock()
	if current, found := d.byKey[pathKey(path)]; found {
		current.diagnostics = diagnostics
	}
}

// change stores the new text and the next version. ok is false when the file is not open.
func (d *openDocuments) change(path, text string) (doc document, ok bool) {
	d.mu.Lock()
	defer d.mu.Unlock()
	current, found := d.byKey[pathKey(path)]
	if !found {
		return document{}, false
	}
	current.text = text
	current.version++
	return *current, true
}

// close forgets a document and reports whether it was open.
func (d *openDocuments) close(path string) bool {
	d.mu.Lock()
	defer d.mu.Unlock()
	_, found := d.byKey[pathKey(path)]
	delete(d.byKey, pathKey(path))
	return found
}

func (d *openDocuments) get(path string) (document, bool) {
	d.mu.Lock()
	defer d.mu.Unlock()
	doc, found := d.byKey[pathKey(path)]
	if !found {
		return document{}, false
	}
	return *doc, true
}

// all returns a copy of every open document.
func (d *openDocuments) all() []document {
	d.mu.Lock()
	defer d.mu.Unlock()
	docs := make([]document, 0, len(d.byKey))
	for _, doc := range d.byKey {
		docs = append(docs, *doc)
	}
	return docs
}

// empty reports whether no document is open.
func (d *openDocuments) empty() bool {
	d.mu.Lock()
	defer d.mu.Unlock()
	return len(d.byKey) == 0
}

// textOf is the text of an open document, or the file on disk, or "".
func (d *openDocuments) textOf(path string) string {
	if doc, ok := d.get(path); ok {
		return doc.text
	}
	data, err := os.ReadFile(path)
	if err != nil {
		return ""
	}
	return string(data)
}

// reset forgets every document.
func (d *openDocuments) reset() {
	d.mu.Lock()
	defer d.mu.Unlock()
	d.byKey = map[string]*document{}
}
