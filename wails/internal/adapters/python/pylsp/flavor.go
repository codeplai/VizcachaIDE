// Package pylsp is the Python flavor of the shared LSP client (protocol/lsp): it starts
// "python -m pylsp" and tells python-lsp-server which plugins a beginner needs.
package pylsp

import (
	"context"
	"os"
	"path/filepath"

	"github.com/codeplai/VizcachaIDE/wails/internal/adapters/python"
	"github.com/codeplai/VizcachaIDE/wails/internal/app"
	"github.com/codeplai/VizcachaIDE/wails/internal/domain"
	"github.com/codeplai/VizcachaIDE/wails/internal/protocol/lsp"
)

const (
	serverName = "pylsp"
	languageID = "python"
)

// Config says where Python is.
type Config struct {
	Locator *python.Locator
	// BaseEnvironment is the "NAME=value" list the server starts from. Nil means os.Environ().
	BaseEnvironment []string
}

// Flavor implements lsp.Flavor for "python -m pylsp".
type Flavor struct{ cfg Config }

var _ lsp.Flavor = Flavor{}

// NewFlavor creates the pylsp flavor.
func NewFlavor(cfg Config) Flavor {
	if cfg.BaseEnvironment == nil {
		cfg.BaseEnvironment = os.Environ()
	}
	return Flavor{cfg: cfg}
}

// New creates the Python language server: pylsp starts on the first OpenDocument.
func New(sink app.EventSink, cfg Config, options lsp.Options) *lsp.Server {
	options.SourceExtensions = []string{".py", ".pyi"}
	options.Name, options.LanguageID, options.CodeLanguage = serverName, languageID, domain.CodeLanguagePython
	return lsp.New(sink, NewFlavor(cfg), options)
}

// Command is "<python> -m pylsp". A missing interpreter is app.MissingTool("python") and a
// missing module app.MissingTool("pylsp").
func (f Flavor) Command(env map[string]string) (string, []string, error) {
	interpreter, err := f.cfg.Locator.Find(context.Background(), "")
	if err != nil {
		return "", nil, err
	}
	versions := python.ModuleVersions(context.Background(), interpreter, env)
	if versions != nil && versions[python.ToolPylsp] == "" {
		return "", nil, app.MissingTool(python.ToolPylsp)
	}
	return interpreter.Path, []string{"-m", "pylsp"}, nil
}

// RootOf is the folder of the file.
func (Flavor) RootOf(path string) string {
	abs, err := filepath.Abs(path)
	if err != nil {
		abs = path
	}
	return filepath.Dir(abs)
}

// InitializationOptions are none: pylsp is configured with Configuration.
func (Flavor) InitializationOptions() any { return nil }

// Configuration is sent as workspace/didChangeConfiguration. pyflakes reports syntax errors,
// undefined names and unused imports; pycodestyle, mccabe and pylint stay off on purpose:
// style warnings overwhelm a beginner (docs/PLAN_PYTHON.md section 4.6).
func (Flavor) Configuration() any {
	return map[string]any{"pylsp": map[string]any{"plugins": map[string]any{
		"pyflakes":            map[string]any{"enabled": true},
		"pycodestyle":         map[string]any{"enabled": false},
		"mccabe":              map[string]any{"enabled": false},
		"pylint":              map[string]any{"enabled": false},
		"jedi_completion":     map[string]any{"include_params": false},
		"jedi_signature_help": map[string]any{"enabled": true},
		"jedi_symbols":        map[string]any{"enabled": true},
	}}}
}

// Environment is the environment of every Python process of the IDE.
func (f Flavor) Environment() map[string]string {
	interpreter, err := f.cfg.Locator.Find(context.Background(), "")
	if err != nil {
		return python.Environment(f.cfg.BaseEnvironment, python.Interpreter{})
	}
	return python.Environment(f.cfg.BaseEnvironment, interpreter)
}
