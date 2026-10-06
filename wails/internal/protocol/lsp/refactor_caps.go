package lsp

import (
	"encoding/json"
	"sync/atomic"
)

// refactorCaps is what the initialize result announced for renaming and finding references.
type refactorCaps struct {
	rename     atomic.Bool // renameProvider
	prepare    atomic.Bool // renameProvider.prepareProvider
	references atomic.Bool // referencesProvider
}

// store reads the capabilities of an initialize result.
func (c *refactorCaps) store(initializeResult json.RawMessage) {
	var result struct {
		Capabilities struct {
			Rename     json.RawMessage `json:"renameProvider"`
			References json.RawMessage `json:"referencesProvider"`
		} `json:"capabilities"`
	}
	if json.Unmarshal(initializeResult, &result) != nil {
		return
	}
	c.rename.Store(announced(result.Capabilities.Rename))
	c.references.Store(announced(result.Capabilities.References))
	var options struct {
		Prepare bool `json:"prepareProvider"`
	}
	_ = json.Unmarshal(result.Capabilities.Rename, &options)
	c.prepare.Store(options.Prepare)
}

// announced tells whether a provider is true or an options object.
func announced(provider json.RawMessage) bool {
	text := string(provider)
	return text != "" && text != "null" && text != "false"
}
