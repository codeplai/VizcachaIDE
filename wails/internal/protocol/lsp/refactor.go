package lsp

import (
	"context"
	"encoding/json"
	"errors"
	"sort"
	"strings"
	"time"

	"go.lsp.dev/jsonrpc2"
	"go.lsp.dev/protocol"

	"github.com/codeplai/VizcachaIDE/wails/internal/domain"
)

const (
	// refactorTimeout is longer than a normal query: renaming and finding references may make
	// the server load and type-check the whole project.
	refactorTimeout = 20 * time.Second
	maxReferences   = 2000
	maxPreview      = 240
)

// refusal builds the rename answer of a server that said no.
func refusal(err error) (domain.RenameRefusal, string) {
	var rpc *jsonrpc2.Error
	switch {
	case errors.Is(err, errNotAnnounced):
		return domain.RenameUnsupported, ""
	case errors.As(err, &rpc) && rpc.Code == jsonrpc2.MethodNotFound:
		return domain.RenameUnsupported, rpc.Message
	case errors.As(err, &rpc):
		return domain.RenameFailed, rpc.Message
	}
	return domain.RenameFailed, err.Error()
}

// PrepareRename asks whether the symbol at a position can be renamed and what its range is. A
// server without prepare support answers "allowed" with no range: the editor takes the word.
func (s *Server) PrepareRename(ctx context.Context, at domain.SourceLocation) (domain.RenameTarget, error) {
	doc, open := s.docs.get(at.File)
	if !open {
		return domain.RenameTarget{Refusal: domain.RenameNotRenameable}, nil
	}
	if !s.awaitReady(ctx, refactorTimeout) {
		return domain.RenameTarget{Refusal: domain.RenameFailed, Detail: errNotAnswered.Error()}, nil
	}
	if !s.refactor.rename.Load() {
		return domain.RenameTarget{Refusal: domain.RenameUnsupported}, nil
	}
	if !s.refactor.prepare.Load() {
		return domain.RenameTarget{}, nil
	}
	raw, err := s.exchange(ctx, refactorTimeout, "textDocument/prepareRename", positionParams(doc, at.Line, at.Column), nil)
	if err != nil {
		reason, detail := refusal(err)
		if reason == domain.RenameFailed { // prepareRename errors say "cannot rename this"
			reason = domain.RenameNotRenameable
		}
		return domain.RenameTarget{Refusal: reason, Detail: detail}, nil
	}
	return toRenameTarget(raw, doc), nil
}

// toRenameTarget maps a prepareRename result: Range, {range, placeholder},
// {defaultBehavior: true} or null (nothing to rename here).
func toRenameTarget(raw json.RawMessage, doc document) domain.RenameTarget {
	var wire struct {
		Start           *protocol.Position `json:"start"`
		End             *protocol.Position `json:"end"`
		Range           *protocol.Range    `json:"range"`
		Placeholder     string             `json:"placeholder"`
		DefaultBehavior bool               `json:"defaultBehavior"`
	}
	if string(raw) == "null" || json.Unmarshal(raw, &wire) != nil {
		return domain.RenameTarget{Refusal: domain.RenameNotRenameable}
	}
	if wire.DefaultBehavior {
		return domain.RenameTarget{}
	}
	lspRange := wire.Range
	if lspRange == nil && wire.Start != nil && wire.End != nil {
		lspRange = &protocol.Range{Start: *wire.Start, End: *wire.End}
	}
	if lspRange == nil {
		return domain.RenameTarget{Refusal: domain.RenameNotRenameable}
	}
	symbol := rangeIn(doc.text, doc.path, *lspRange)
	return domain.RenameTarget{Range: &symbol, Placeholder: wire.Placeholder}
}

// Rename asks the server for the edits that rename the symbol at a position. Nothing is
// applied: the editor edits the open files and the backend edits the others.
func (s *Server) Rename(ctx context.Context, at domain.SourceLocation, newName string) (domain.RenameResult, error) {
	none := domain.RenameResult{Files: []domain.FileEdit{}}
	doc, open := s.docs.get(at.File)
	if !open {
		none.Refusal = domain.RenameNotRenameable
		return none, nil
	}
	params := map[string]any{
		"textDocument": protocol.TextDocumentIdentifier{URI: pathToURI(doc.path)},
		"position":     toLSPPosition(doc.text, at.Line, at.Column),
		"newName":      newName,
	}
	raw, err := s.exchange(ctx, refactorTimeout, "textDocument/rename", params, &s.refactor.rename)
	if err != nil {
		none.Refusal, none.Detail = refusal(err)
		return none, nil
	}
	files, err := toWorkspaceEdit(raw, s.docs.textOf)
	switch {
	case err != nil:
		none.Refusal, none.Detail = domain.RenameFailed, err.Error()
	case len(files) == 0:
		none.Refusal = domain.RenameUnsupported
	default:
		none.Files = files
	}
	return none, nil
}

// References returns every use of the symbol at a position, declaration included.
func (s *Server) References(ctx context.Context, at domain.SourceLocation) ([]domain.Reference, error) {
	doc, open := s.docs.get(at.File)
	if !open {
		return []domain.Reference{}, nil
	}
	params := map[string]any{
		"textDocument": protocol.TextDocumentIdentifier{URI: pathToURI(doc.path)},
		"position":     toLSPPosition(doc.text, at.Line, at.Column),
		"context":      map[string]any{"includeDeclaration": true},
	}
	raw, err := s.exchange(ctx, refactorTimeout, "textDocument/references", params, &s.refactor.references)
	if err != nil {
		return []domain.Reference{}, nil
	}
	return toReferences(raw, s.docs.textOf), nil
}

// toReferences maps a references result (Location[] or null), ordered by file and position.
func toReferences(raw json.RawMessage, textOf func(path string) string) []domain.Reference {
	var locations []struct {
		URI   protocol.DocumentURI `json:"uri"`
		Range protocol.Range       `json:"range"`
	}
	result := []domain.Reference{}
	if json.Unmarshal(raw, &locations) != nil {
		return result
	}
	texts := map[string]string{}
	for _, location := range locations[:min(len(locations), maxReferences)] {
		path := uriToPath(location.URI)
		text, known := texts[path]
		if !known {
			text = textOf(path)
			texts[path] = text
		}
		where := rangeIn(text, path, location.Range)
		result = append(result, domain.Reference{Range: where, Preview: preview(text, where.Start.Line)})
	}
	sort.SliceStable(result, func(i, j int) bool { return lessReference(result[i], result[j]) })
	return result
}

func lessReference(a, b domain.Reference) bool {
	if a.Range.Start.File != b.Range.Start.File {
		return a.Range.Start.File < b.Range.Start.File
	}
	if a.Range.Start.Line != b.Range.Start.Line {
		return a.Range.Start.Line < b.Range.Start.Line
	}
	return a.Range.Start.Column < b.Range.Start.Column
}

// preview is the trimmed text of a 1-based line, cut to a reasonable length.
func preview(text string, line int) string {
	runes := []rune(strings.TrimSpace(lineOf(text, line-1)))
	if len(runes) > maxPreview {
		return string(runes[:maxPreview]) + "…"
	}
	return string(runes)
}
