package lsp

import (
	"encoding/json"

	"go.lsp.dev/protocol"

	"github.com/codeplai/VizcachaIDE/wails/internal/domain"
)

var symbolKinds = map[protocol.SymbolKind]domain.SymbolKind{
	protocol.SymbolKindFunction:      domain.SymbolFunction,
	protocol.SymbolKindConstructor:   domain.SymbolFunction,
	protocol.SymbolKindMethod:        domain.SymbolMethod,
	protocol.SymbolKindStruct:        domain.SymbolStruct,
	protocol.SymbolKindInterface:     domain.SymbolInterface,
	protocol.SymbolKindClass:         domain.SymbolType,
	protocol.SymbolKindEnum:          domain.SymbolType,
	protocol.SymbolKindTypeParameter: domain.SymbolType,
	protocol.SymbolKindVariable:      domain.SymbolVariable,
	protocol.SymbolKindConstant:      domain.SymbolConstant,
	protocol.SymbolKindEnumMember:    domain.SymbolConstant,
	protocol.SymbolKindField:         domain.SymbolField,
	protocol.SymbolKindProperty:      domain.SymbolField,
	protocol.SymbolKindPackage:       domain.SymbolPackage,
	protocol.SymbolKindModule:        domain.SymbolPackage,
	protocol.SymbolKindNamespace:     domain.SymbolPackage,
}

type symbolWire struct {
	Name           string         `json:"name"`
	Detail         string         `json:"detail"`
	Kind           int            `json:"kind"`
	Range          protocol.Range `json:"range"`
	SelectionRange protocol.Range `json:"selectionRange"`
	Children       []symbolWire   `json:"children"`
}

// toDocumentSymbols maps a documentSymbol result made of nested DocumentSymbol.
func toDocumentSymbols(raw json.RawMessage, file, text string) []domain.DocumentSymbol {
	var symbols []symbolWire
	if json.Unmarshal(raw, &symbols) != nil {
		return []domain.DocumentSymbol{}
	}
	return mapSymbols(symbols, file, text)
}

func mapSymbols(symbols []symbolWire, file, text string) []domain.DocumentSymbol {
	result := make([]domain.DocumentSymbol, 0, len(symbols))
	for _, symbol := range symbols {
		kind, known := symbolKinds[protocol.SymbolKind(symbol.Kind)]
		if !known {
			kind = domain.SymbolOther
		}
		whole := rangeIn(text, file, symbol.Range)
		result = append(result, domain.DocumentSymbol{
			Name: symbol.Name, Kind: kind, Detail: symbol.Detail,
			Location: locationIn(text, file, symbol.SelectionRange.Start),
			Range:    &whole,
			Children: mapSymbols(symbol.Children, file, text),
		})
	}
	return result
}

// toHighlightRanges maps a documentHighlight result.
func toHighlightRanges(raw json.RawMessage, file, text string) []domain.SourceRange {
	var highlights []struct {
		Range protocol.Range `json:"range"`
	}
	result := []domain.SourceRange{}
	if json.Unmarshal(raw, &highlights) != nil {
		return result
	}
	for _, highlight := range highlights {
		result = append(result, rangeIn(text, file, highlight.Range))
	}
	return result
}

type targetWire struct {
	URI                  protocol.DocumentURI `json:"uri"`
	Range                protocol.Range       `json:"range"`
	TargetURI            protocol.DocumentURI `json:"targetUri"`
	TargetSelectionRange *protocol.Range      `json:"targetSelectionRange"`
}

// toDefinition maps the first target of a definition result (Location or LocationLink,
// single or list). textOf gives the text of the target file to convert its columns.
func toDefinition(raw json.RawMessage, textOf func(path string) string) *domain.SourceLocation {
	var targets []targetWire
	if json.Unmarshal(raw, &targets) != nil {
		var single targetWire
		if json.Unmarshal(raw, &single) != nil {
			return nil
		}
		targets = []targetWire{single}
	}
	if len(targets) == 0 {
		return nil
	}
	target := targets[0]
	targetURI, selected := target.URI, target.Range
	if target.TargetURI != "" {
		targetURI = target.TargetURI
	}
	if target.TargetSelectionRange != nil {
		selected = *target.TargetSelectionRange
	}
	if targetURI == "" {
		return nil
	}
	path := uriToPath(targetURI)
	location := locationIn(textOf(path), path, selected.Start)
	return &location
}
