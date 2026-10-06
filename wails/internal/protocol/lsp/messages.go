package lsp

import (
	"os"
	"path/filepath"

	"go.lsp.dev/protocol"
)

const clientName = "VizcachaIDE"

var plainText = []string{"plaintext"}

// clientCapabilities mirrors the 1.0 client: plain text, no snippets, hierarchical symbols.
func clientCapabilities() map[string]any {
	return map[string]any{
		"workspace": map[string]any{"workspaceFolders": true, "workspaceEdit": map[string]any{"documentChanges": true}, "didChangeWatchedFiles": map[string]any{"dynamicRegistration": false}},
		"general":   map[string]any{"positionEncodings": []string{"utf-16"}},
		"textDocument": map[string]any{
			"synchronization":    map[string]any{"didSave": false},
			"completion":         map[string]any{"completionItem": map[string]any{"snippetSupport": false, "documentationFormat": plainText}},
			"hover":              map[string]any{"contentFormat": plainText},
			"definition":         map[string]any{"linkSupport": false},
			"documentHighlight":  map[string]any{},
			"documentSymbol":     map[string]any{"hierarchicalDocumentSymbolSupport": true},
			"publishDiagnostics": map[string]any{},
			"inlayHint":          map[string]any{},
			"references":         map[string]any{},
			"rename":             map[string]any{"prepareSupport": true},
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

// initializeParams builds the handshake; initializationOptions are the flavor's (nil: none).
func initializeParams(root string, initializationOptions any, pullsConfiguration bool) map[string]any {
	capabilities := clientCapabilities()
	if pullsConfiguration { // the server may ask workspace/configuration (pulled.go)
		capabilities["workspace"] = map[string]any{
			"workspaceFolders": true, "configuration": true, "workspaceEdit": map[string]any{"documentChanges": true},
			"didChangeWatchedFiles": map[string]any{"dynamicRegistration": false},
		}
	}
	params := map[string]any{
		"processId":        os.Getpid(),
		"clientInfo":       map[string]any{"name": clientName},
		"rootUri":          string(pathToURI(root)),
		"workspaceFolders": []any{workspaceFolder(root)},
		"capabilities":     capabilities,
	}
	if initializationOptions != nil {
		params["initializationOptions"] = initializationOptions
	}
	return params
}

func addFolderParams(root string) map[string]any {
	return map[string]any{"event": map[string]any{
		"added": []any{workspaceFolder(root)}, "removed": []any{},
	}}
}

func didOpenParams(doc document, languageID string) protocol.DidOpenTextDocumentParams {
	return protocol.DidOpenTextDocumentParams{TextDocument: protocol.TextDocumentItem{
		URI: pathToURI(doc.path), LanguageID: protocol.LanguageIdentifier(languageID), Version: doc.version, Text: doc.text,
	}}
}

// fullTextChange is a change event without a range: the whole document is replaced.
// protocol.TextDocumentContentChangeEvent cannot be used for this: its Range is not a pointer, so
// it always serializes as {0:0 - 0:0} and gopls inserts the text at the top (a duplicated file).
type fullTextChange struct {
	Text string `json:"text"`
}

type didChangeFullParams struct {
	TextDocument   protocol.VersionedTextDocumentIdentifier `json:"textDocument"`
	ContentChanges []fullTextChange                         `json:"contentChanges"`
}

// didChangeParams sends the whole document, versioned.
func didChangeParams(doc document) didChangeFullParams {
	return didChangeFullParams{
		TextDocument: protocol.VersionedTextDocumentIdentifier{
			TextDocumentIdentifier: protocol.TextDocumentIdentifier{URI: pathToURI(doc.path)},
			Version:                doc.version,
		},
		ContentChanges: []fullTextChange{{Text: doc.text}},
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
