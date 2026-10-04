package lsp

import (
	"context"
	"encoding/json"

	"go.lsp.dev/protocol"

	"github.com/codeplai/VizcachaIDE/wails/internal/domain"
)

// LSP InlayHintKind values.
const (
	lspInlayType      = 1
	lspInlayParameter = 2
)

// announcesInlayHints tells whether an initialize result has an inlayHintProvider (true or options).
func announcesInlayHints(initializeResult json.RawMessage) bool {
	var result struct {
		Capabilities struct {
			Provider json.RawMessage `json:"inlayHintProvider"`
		} `json:"capabilities"`
	}
	if json.Unmarshal(initializeResult, &result) != nil {
		return false
	}
	provider := string(result.Capabilities.Provider)
	return provider != "" && provider != "null" && provider != "false"
}

// InlayHints returns the hints of the lines from visible.Start.Line to visible.End.Line (both
// included, whole lines) of an open document. A server without inlayHintProvider is not asked.
// workspace/inlayHint/refresh needs no handling: the connection answers every server request
// with null and the editor asks again after each edit.
func (s *Server) InlayHints(ctx context.Context, visible domain.SourceRange) ([]domain.InlayHint, error) {
	doc, open := s.docs.get(visible.Start.File)
	if !open {
		return []domain.InlayHint{}, nil
	}
	lastLine := max(visible.End.Line, visible.Start.Line)
	params := map[string]any{
		"textDocument": protocol.TextDocumentIdentifier{URI: pathToURI(doc.path)},
		"range": protocol.Range{
			Start: toLSPPosition(doc.text, visible.Start.Line, 1),
			End:   endOfLine(doc.text, lastLine),
		},
	}
	raw, ok := s.requestIf(ctx, "textDocument/inlayHint", params, &s.hints)
	if !ok {
		return []domain.InlayHint{}, nil
	}
	return toInlayHints(raw, doc.text), nil
}

// endOfLine is the LSP position after the last character of a 1-based line.
func endOfLine(text string, line int) protocol.Position {
	index := max(line-1, 0)
	return protocol.Position{Line: uint32(index), Character: uint32(utf16Offset(lineOf(text, index), 1<<30))}
}

type inlayWire struct {
	Position     protocol.Position `json:"position"`
	Label        json.RawMessage   `json:"label"`
	Kind         int               `json:"kind"`
	PaddingLeft  bool              `json:"paddingLeft"`
	PaddingRight bool              `json:"paddingRight"`
}

// toInlayHints maps an inlayHint result; never nil.
func toInlayHints(raw json.RawMessage, text string) []domain.InlayHint {
	hints := []domain.InlayHint{}
	var wire []inlayWire
	if json.Unmarshal(raw, &wire) != nil {
		return hints
	}
	for _, item := range wire {
		label := inlayLabel(item.Label)
		if label == "" {
			continue
		}
		line, column := fromLSPPosition(text, item.Position)
		hints = append(hints, domain.InlayHint{
			Line: line, Column: column, Label: label, Kind: inlayKind(item.Kind),
			PaddingLeft: item.PaddingLeft, PaddingRight: item.PaddingRight,
		})
	}
	return hints
}

// inlayLabel is a string, or the joined values of InlayHintLabelPart[].
func inlayLabel(raw json.RawMessage) string {
	var text string
	if json.Unmarshal(raw, &text) == nil {
		return text
	}
	var parts []struct {
		Value string `json:"value"`
	}
	if json.Unmarshal(raw, &parts) != nil {
		return ""
	}
	for _, part := range parts {
		text += part.Value
	}
	return text
}

func inlayKind(kind int) domain.InlayHintKind {
	switch kind {
	case lspInlayType:
		return domain.InlayHintType
	case lspInlayParameter:
		return domain.InlayHintParameter
	}
	return domain.InlayHintOther
}
