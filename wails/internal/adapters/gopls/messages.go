package gopls

import (
	"os"
	"path/filepath"

	"go.lsp.dev/protocol"
)

const (
	clientName = "VizcachaIDE"
	languageID = "go"
)

var plainText = []string{"plaintext"}

// clientCapabilities mirrors the 1.0 client: plain text, no snippets, hierarchical symbols.
func clientCapabilities() map[string]any {
	return map[string]any{
		"workspace": map[string]any{"workspaceFolders": true},
		"general":   map[string]any{"positionEncodings": []string{"utf-16"}},
		"textDocument": map[string]any{
			"synchronization":    map[string]any{"didSave": false},
			"completion":         map[string]any{"completionItem": map[string]any{"snippetSupport": false, "documentationFormat": plainText}},
			"hover":              map[string]any{"contentFormat": plainText},
			"definition":         map[string]any{"linkSupport": false},
			"documentHighlight":  map[string]any{},
			"documentSymbol":     map[string]any{"hierarchicalDocumentSymbolSupport": true},
			"publishDiagnostics": map[string]any{},
			"signatureHelp": map[string]any{"signatureInformation": map[string]any{
				"documentationFormat":    plainText,
				"parameterInformation":   map[string]any{"labelOffsetSupport": true},
				"activeParameterSupport": true,
			}},
		},
	}
}

func workspaceFolder(root string) map[string]any {
	name := filepath.Base(root)
	return map[string]any{"uri": string(pathToURI(root)), "name": name}
}

func initializeParams(root string) map[string]any {
	return map[string]any{
		"processId":        os.Getpid(),
		"clientInfo":       map[string]any{"name": clientName},
		"rootUri":          string(pathToURI(root)),
		"workspaceFolders": []any{workspaceFolder(root)},
		"capabilities":     clientCapabilities(),
	}
}

func addFolderParams(root string) map[string]any {
	return map[string]any{"event": map[string]any{
		"added": []any{workspaceFolder(root)}, "removed": []any{},
	}}
}

func didOpenParams(doc document) protocol.DidOpenTextDocumentParams {
	return protocol.DidOpenTextDocumentParams{TextDocument: protocol.TextDocumentItem{
		URI: pathToURI(doc.path), LanguageID: languageID, Version: doc.version, Text: doc.text,
	}}
}

// didChangeParams sends the whole document, versioned.
func didChangeParams(doc document) protocol.DidChangeTextDocumentParams {
	return protocol.DidChangeTextDocumentParams{
		TextDocument: protocol.VersionedTextDocumentIdentifier{
			TextDocumentIdentifier: protocol.TextDocumentIdentifier{URI: pathToURI(doc.path)},
			Version:                doc.version,
		},
		ContentChanges: []protocol.TextDocumentContentChangeEvent{{Text: doc.text}},
	}
}

func didCloseParams(path string) protocol.DidCloseTextDocumentParams {
	return protocol.DidCloseTextDocumentParams{
		TextDocument: protocol.TextDocumentIdentifier{URI: pathToURI(path)},
	}
}

func positionParams(doc document, line, column int) protocol.TextDocumentPositionParams {
	return protocol.TextDocumentPositionParams{
		TextDocument: protocol.TextDocumentIdentifier{URI: pathToURI(doc.path)},
		Position:     toLSPPosition(doc.text, line, column),
	}
}

func documentParams(path string) protocol.DocumentSymbolParams {
	return protocol.DocumentSymbolParams{
		TextDocument: protocol.TextDocumentIdentifier{URI: pathToURI(path)},
	}
}
