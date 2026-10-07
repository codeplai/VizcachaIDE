package lsp

import (
	"encoding/json"
	"errors"
	"sort"

	"go.lsp.dev/protocol"

	"github.com/codeplai/VizcachaIDE/wails/internal/domain"
)

// errResourceOperation means the server asked to create, rename or delete a file, which the
// editor did not announce it supports.
var errResourceOperation = errors.New("the rename needs file operations, which are not supported")

type editWire struct {
	Range   protocol.Range `json:"range"`
	NewText string         `json:"newText"`
}

type documentChangeWire struct {
	Kind         string `json:"kind"`
	TextDocument struct {
		URI protocol.DocumentURI `json:"uri"`
	} `json:"textDocument"`
	Edits []editWire `json:"edits"`
}

type workspaceEditWire struct {
	Changes         map[protocol.DocumentURI][]editWire `json:"changes"`
	DocumentChanges []json.RawMessage                   `json:"documentChanges"`
}

// editCollector gathers the edits of every file, merging the several entries of one file.
type editCollector struct {
	textOf  func(path string) string
	files   []domain.FileEdit
	indexOf map[string]int
}

func (c *editCollector) add(uri protocol.DocumentURI, edits []editWire) {
	path := uriToPath(uri)
	index, seen := c.indexOf[pathKey(path)]
	if !seen {
		index = len(c.files)
		c.indexOf[pathKey(path)] = index
		c.files = append(c.files, domain.FileEdit{File: path, Edits: []domain.TextEdit{}})
	}
	text := c.textOf(path)
	for _, edit := range edits {
		c.files[index].Edits = append(c.files[index].Edits, domain.TextEdit{
			Range: rangeIn(text, path, edit.Range), NewText: edit.NewText,
		})
	}
}

// toWorkspaceEdit maps a WorkspaceEdit result in both of its forms: documentChanges (preferred
// when present) and changes. Columns become runes using the text of each file. Files with no
// edits are left out.
func toWorkspaceEdit(raw json.RawMessage, textOf func(path string) string) ([]domain.FileEdit, error) {
	var wire workspaceEditWire
	if err := json.Unmarshal(raw, &wire); err != nil {
		return nil, err
	}
	collector := &editCollector{textOf: textOf, files: []domain.FileEdit{}, indexOf: map[string]int{}}
	if len(wire.DocumentChanges) > 0 {
		for _, item := range wire.DocumentChanges {
			var change documentChangeWire
			if err := json.Unmarshal(item, &change); err != nil {
				return nil, err
			}
			if change.Kind != "" {
				return nil, errResourceOperation
			}
			collector.add(change.TextDocument.URI, change.Edits)
		}
		return nonEmpty(collector.files), nil
	}
	uris := make([]string, 0, len(wire.Changes))
	for uri := range wire.Changes {
		uris = append(uris, string(uri))
	}
	sort.Strings(uris)
	for _, uri := range uris {
		collector.add(protocol.DocumentURI(uri), wire.Changes[protocol.DocumentURI(uri)])
	}
	return nonEmpty(collector.files), nil
}

func nonEmpty(files []domain.FileEdit) []domain.FileEdit {
	kept := files[:0]
	for _, file := range files {
		if len(file.Edits) > 0 {
			kept = append(kept, file)
		}
	}
	return kept
}
