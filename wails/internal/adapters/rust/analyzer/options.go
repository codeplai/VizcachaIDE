package analyzer

// InitializationOptions are the settings of rust-analyzer at the handshake (see settings).
func (f *Flavor) InitializationOptions() any { return f.settings() }

// Configuration is the same settings again. rust-analyzer ignores what comes in
// workspace/didChangeConfiguration and asks the client with workspace/configuration instead;
// a client that answers that request with this value makes the loose files opened after the
// first one detached too (docs/PLAN_RUST.md section 4.5; CCR to protocol/lsp).
func (f *Flavor) Configuration() any { return f.settings() }

// settings are the options of docs/PLAN_RUST.md section 4.5: clippy on every save, so the live
// problems of a project include its advice; build scripts and proc macros on, so derives and
// generated code resolve; type and parameter hints on (the editor that draws them is track
// R6's), chaining and closing-brace hints off to keep the code readable for a beginner. The
// loose .rs files are detached files: without that rust-analyzer gives nothing for them.
func (f *Flavor) settings() map[string]any {
	options := map[string]any{
		"checkOnSave": true,
		"check":       map[string]any{"command": "clippy"},
		"cargo":       map[string]any{"buildScripts": map[string]any{"enable": true}},
		"procMacro":   map[string]any{"enable": true},
		"inlayHints": map[string]any{
			"typeHints":         map[string]any{"enable": true},
			"parameterHints":    map[string]any{"enable": true},
			"chainingHints":     map[string]any{"enable": false},
			"closingBraceHints": map[string]any{"enable": false},
		},
	}
	f.mu.Lock()
	defer f.mu.Unlock()
	if len(f.loose) > 0 {
		options["detachedFiles"] = append([]string{}, f.loose...)
	}
	return options
}
