// Package analyzer is the Rust flavor of the shared LSP client (protocol/lsp): it starts
// rust-analyzer with the options a beginner needs (docs/PLAN_RUST.md section 4.5).
package analyzer

import (
	"os"
	"path/filepath"
	"slices"
	"strings"
	"sync"

	"github.com/codeplai/VizcachaIDE/wails/internal/adapters/rust"
	"github.com/codeplai/VizcachaIDE/wails/internal/app"
	"github.com/codeplai/VizcachaIDE/wails/internal/domain"
	"github.com/codeplai/VizcachaIDE/wails/internal/protocol/lsp"
)

const (
	serverName = "rust-analyzer"
	languageID = "rust"
)

// Config says where rust-analyzer is.
type Config struct {
	Locator *rust.Locator
}

// Flavor implements lsp.Flavor for rust-analyzer.
type Flavor struct {
	cfg Config
	// loose lists the .rs files without Cargo.toml that RootOf has seen: rust-analyzer only
	// serves them when they are listed as detached files (see settings).
	mu    sync.Mutex
	loose []string
}

var (
	_ lsp.Flavor              = (*Flavor)(nil)
	_ lsp.PulledConfiguration = (*Flavor)(nil)
)

// PullsConfiguration is true: rust-analyzer asks for its settings, and a new loose file must
// make it ask again (lsp.PulledConfiguration).
func (f *Flavor) PullsConfiguration() bool { return true }

// NewFlavor creates the rust-analyzer flavor.
func NewFlavor(cfg Config) *Flavor { return &Flavor{cfg: cfg} }

// New creates the Rust language server: rust-analyzer starts on the first OpenDocument.
func New(sink app.EventSink, cfg Config, options lsp.Options) *lsp.Server {
	options.Name, options.LanguageID, options.CodeLanguage = serverName, languageID, domain.CodeLanguageRust
	return lsp.New(sink, NewFlavor(cfg), options)
}

// Command is rust-analyzer without arguments (it speaks LSP on stdio). A missing one is
// app.MissingTool("rust-analyzer").
func (f *Flavor) Command(map[string]string) (string, []string, error) {
	status := f.cfg.Locator.Tool(rust.ToolRustAnalyzer)
	if status.Source == domain.ToolMissing || status.Path == "" {
		return "", nil, app.MissingTool(rust.ToolRustAnalyzer)
	}
	return status.Path, nil, nil
}

// RootOf is the root of the Cargo workspace the file belongs to, so one rust-analyzer serves
// every member. A loose file has no Cargo.toml: its root is its folder and it is remembered as
// a detached file.
func (f *Flavor) RootOf(path string) string {
	abs, err := filepath.Abs(path)
	if err != nil {
		abs = path
	}
	project, found, err := rust.FindProject(abs)
	f.mu.Lock()
	defer f.mu.Unlock()
	if err != nil || !found {
		if !slices.Contains(f.loose, abs) {
			f.loose = append(f.loose, abs)
		}
		return filepath.Dir(abs)
	}
	return project.Workspace
}

// Environment is what the Rust environment changes in the one of the IDE: CARGO_HOME,
// RUSTUP_HOME, the cargo folder first in PATH and plain output. Only the variables that differ
// are returned, spelled as the process has them ("Path" on Windows), because the client merges
// them into os.Environ() by exact name.
func (f *Flavor) Environment() map[string]string {
	current := map[string]string{}
	spelling := map[string]string{}
	for _, entry := range os.Environ() {
		if name, value, ok := cut(entry); ok {
			current[name], spelling[strings.ToUpper(name)] = value, name
		}
	}
	changed := map[string]string{}
	for _, entry := range f.cfg.Locator.Environment() {
		name, value, ok := cut(entry)
		if !ok {
			continue
		}
		if existing, found := spelling[strings.ToUpper(name)]; found {
			name = existing
		}
		if current[name] != value {
			changed[name] = value
		}
	}
	return changed
}

// cut splits "NAME=value"; Windows has hidden entries like "=C:=C:\" whose name starts with "=".
func cut(entry string) (name, value string, ok bool) {
	i := strings.Index(entry[min(1, len(entry)):], "=")
	if i < 0 {
		return "", "", false
	}
	i++
	return entry[:i], entry[i+1:], true
}
