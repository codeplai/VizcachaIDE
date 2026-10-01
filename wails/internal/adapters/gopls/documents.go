package gopls

import (
	"os"
	"sync"
)

// document is a file the editor has open.
type document struct {
	path    string
	text    string
	version int32
}

// openDocuments is the registry of open documents, safe for concurrent use.
type openDocuments struct {
	mu    sync.Mutex
	byKey map[string]*document
}

func newOpenDocuments() *openDocuments {
	return &openDocuments{byKey: map[string]*document{}}
}

// open registers (or replaces) a document at version 1.
func (d *openDocuments) open(path, text string) document {
	d.mu.Lock()
	defer d.mu.Unlock()
	doc := &document{path: path, text: text, version: 1}
	d.byKey[pathKey(path)] = doc
	return *doc
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
