package lsp

import (
	"encoding/json"
	"fmt"
	"strings"

	"go.lsp.dev/protocol"

	"github.com/codeplai/VizcachaIDE/wails/internal/domain"
)

const (
	maxCompletions = 60
)

var severities = map[protocol.DiagnosticSeverity]domain.Severity{
	protocol.DiagnosticSeverityError:       domain.SeverityError,
	protocol.DiagnosticSeverityWarning:     domain.SeverityWarning,
	protocol.DiagnosticSeverityInformation: domain.SeverityInfo,
	protocol.DiagnosticSeverityHint:        domain.SeverityHint,
}

var completionKinds = map[protocol.CompletionItemKind]domain.CompletionKind{
	protocol.CompletionItemKindMethod:        domain.CompletionMethod,
	protocol.CompletionItemKindFunction:      domain.CompletionFunction,
	protocol.CompletionItemKindConstructor:   domain.CompletionFunction,
	protocol.CompletionItemKindField:         domain.CompletionField,
	protocol.CompletionItemKindProperty:      domain.CompletionField,
	protocol.CompletionItemKindVariable:      domain.CompletionVariable,
	protocol.CompletionItemKindConstant:      domain.CompletionConstant,
	protocol.CompletionItemKindEnumMember:    domain.CompletionConstant,
	protocol.CompletionItemKindClass:         domain.CompletionType,
	protocol.CompletionItemKindInterface:     domain.CompletionType,
	protocol.CompletionItemKindStruct:        domain.CompletionType,
	protocol.CompletionItemKindEnum:          domain.CompletionType,
	protocol.CompletionItemKindTypeParameter: domain.CompletionType,
	protocol.CompletionItemKindModule:        domain.CompletionPackage,
	protocol.CompletionItemKindKeyword:       domain.CompletionKeyword,
}

// markupText is the plain text of a string | MarkupContent | MarkedString | list value.
func markupText(raw json.RawMessage) string {
	var text string
	if json.Unmarshal(raw, &text) == nil {
		return text
	}
	var object struct {
		Value string `json:"value"`
	}
	if json.Unmarshal(raw, &object) == nil && object.Value != "" {
		return object.Value
	}
	var parts []json.RawMessage
	if json.Unmarshal(raw, &parts) != nil {
		return ""
	}
	texts := make([]string, 0, len(parts))
	for _, part := range parts {
		texts = append(texts, markupText(part))
	}
	return strings.Join(texts, "\n\n")
}

// toDiagnostics maps the params of textDocument/publishDiagnostics. file is the path
// the editor knows and text its current content (needed to convert UTF-16 columns).
// source names the server (Diagnostic.Source).
func toDiagnostics(params protocol.PublishDiagnosticsParams, file, text, source string) []domain.Diagnostic {
	result := make([]domain.Diagnostic, 0, len(params.Diagnostics))
	for _, item := range params.Diagnostics {
		start := locationIn(text, file, item.Range.Start)
		end := locationIn(text, file, item.Range.End)
		severity, known := severities[item.Severity]
		if !known {
			severity = domain.SeverityError
		}
		result = append(result, domain.Diagnostic{
			Location: &start, End: &end, Severity: severity,
			Message: item.Message, RawText: item.Message,
			Source: source, Code: codeText(item.Code),
		})
	}
	return result
}

func codeText(code interface{}) string {
	if code == nil {
		return ""
	}
	if number, ok := code.(float64); ok {
		return fmt.Sprintf("%d", int64(number))
	}
	return fmt.Sprint(code)
}

type textEditWire struct {
	NewText string `json:"newText"`
}

type completionWire struct {
	Label         string          `json:"label"`
	Kind          int             `json:"kind"`
	Detail        string          `json:"detail"`
	Documentation json.RawMessage `json:"documentation"`
	InsertText    string          `json:"insertText"`
	TextEdit      *textEditWire   `json:"textEdit"`
}

// toCompletionItems maps a completion result (list, or object with items).
func toCompletionItems(raw json.RawMessage) []domain.CompletionItem {
	var items []completionWire
	if json.Unmarshal(raw, &items) != nil {
		var list struct {
			Items []completionWire `json:"items"`
		}
		_ = json.Unmarshal(raw, &list)
		items = list.Items
	}
	items = items[:min(len(items), maxCompletions)]
	result := make([]domain.CompletionItem, 0, len(items))
	for _, item := range items {
		result = append(result, toCompletionItem(item))
	}
	return result
}

func toCompletionItem(item completionWire) domain.CompletionItem {
	// clangd's detailed style pads the label with a space (or a "•" when it would add an include).
	item.Label = strings.TrimLeft(item.Label, " •")
	insert := item.InsertText
	if item.TextEdit != nil && item.TextEdit.NewText != "" {
		insert = item.TextEdit.NewText
	}
	if insert == item.Label {
		insert = ""
	}
	kind, known := completionKinds[protocol.CompletionItemKind(item.Kind)]
	if !known {
		kind = domain.CompletionOther
	}
	return domain.CompletionItem{
		Label: item.Label, Kind: kind, Detail: item.Detail,
		Documentation: markupText(item.Documentation), InsertText: insert,
	}
}

// toHoverText maps a hover result; "" when there is nothing to show.
func toHoverText(raw json.RawMessage) string {
	var hover struct {
		Contents json.RawMessage `json:"contents"`
	}
	if json.Unmarshal(raw, &hover) != nil {
		return ""
	}
	return strings.TrimSpace(markupText(hover.Contents))
}
