// Package gopls is the Go flavor of the shared LSP client (protocol/lsp): where gopls is,
// how it is started and which folder is a project.
package gopls

import (
	"os"
	"path/filepath"

	"github.com/codeplai/VizcachaIDE/wails/internal/app"
	"github.com/codeplai/VizcachaIDE/wails/internal/protocol/lsp"
)

const (
	toolID         = "gopls"
	goModuleFile   = "go.mod"
	languageID     = "go"
	diagnosticName = "gopls"
)

// Config says where gopls is. The zero value searches for gopls.
type Config struct {
	// Executable returns the path of gopls each time the server starts. Nil or an empty
	// result means "search for it".
	Executable func() string
	// Environment returns the variables of the Go toolchain (may be nil).
	Environment func() map[string]string
}

// Flavor implements lsp.Flavor for "gopls serve".
type Flavor struct{ cfg Config }

var _ lsp.Flavor = Flavor{}

// NewFlavor creates the gopls flavor.
func NewFlavor(cfg Config) Flavor { return Flavor{cfg: cfg} }

// New creates the Go language server: gopls starts on the first OpenDocument.
func New(sink app.EventSink, cfg Config, options lsp.Options) *lsp.Server {
	options.Name, options.LanguageID = diagnosticName, languageID
	return lsp.New(sink, NewFlavor(cfg), options)
}

// Command locates gopls (configured path, toolchain dirs, PATH). A missing gopls is
// reported as app.MissingTool("gopls").
func (f Flavor) Command(env map[string]string) (string, []string, error) {
	configured := ""
	if f.cfg.Executable != nil {
		configured = f.cfg.Executable()
	}
	executable, err := locateGopls(configured, env)
	if err != nil {
		return "", nil, err
	}
	return executable, []string{"serve"}, nil
}

// RootOf is the folder with the nearest go.mod above path, or the file's own folder.
func (Flavor) RootOf(path string) string { return moduleRoot(path) }

// InitializationOptions are none: gopls works with its defaults.
func (Flavor) InitializationOptions() any { return nil }

// Configuration is none: gopls needs no workspace/didChangeConfiguration.
func (Flavor) Configuration() any { return nil }

// Environment returns the Go toolchain variables.
func (f Flavor) Environment() map[string]string {
	if f.cfg.Environment == nil {
		return nil
	}
	return f.cfg.Environment()
}

// moduleRoot is the folder with the nearest go.mod above file, or the file's own folder.
func moduleRoot(file string) string {
	abs, err := filepath.Abs(file)
	if err != nil {
		abs = file
	}
	folder := filepath.Dir(abs)
	for candidate := folder; ; candidate = filepath.Dir(candidate) {
		if info, err := os.Stat(filepath.Join(candidate, goModuleFile)); err == nil && !info.IsDir() {
			return candidate
		}
		if filepath.Dir(candidate) == candidate {
			return folder
		}
	}
}
