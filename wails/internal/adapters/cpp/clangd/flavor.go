// Package clangd is the C++ flavor of the shared LSP client (protocol/lsp): it starts clangd
// with the options a beginner needs (docs/PLAN_CPP.md section 4.6).
package clangd

import (
	"context"
	"path/filepath"

	"github.com/codeplai/VizcachaIDE/wails/internal/adapters/cpp"
	"github.com/codeplai/VizcachaIDE/wails/internal/app"
	"github.com/codeplai/VizcachaIDE/wails/internal/domain"
	"github.com/codeplai/VizcachaIDE/wails/internal/protocol/lsp"
)

const (
	serverName = "clangd"
	languageID = "cpp"
)

// Config says where the compiler and clangd are.
type Config struct {
	Locator *cpp.Locator
	// BaseEnvironment is the "NAME=value" list the server starts from. Nil means os.Environ().
	BaseEnvironment []string
}

// Flavor implements lsp.Flavor for clangd.
type Flavor struct{ cfg Config }

var _ lsp.Flavor = Flavor{}

// NewFlavor creates the clangd flavor.
func NewFlavor(cfg Config) Flavor { return Flavor{cfg: cfg} }

// New creates the C++ language server: clangd starts on the first OpenDocument.
func New(sink app.EventSink, cfg Config, options lsp.Options) *lsp.Server {
	options.Name, options.LanguageID, options.CodeLanguage = serverName, languageID, domain.CodeLanguageCpp
	return lsp.New(sink, NewFlavor(cfg), options)
}

// Command is clangd with its fixed arguments. With a GCC compiler it adds --query-driver so
// clangd asks g++ for its system headers (otherwise <iostream> is not found). A missing clangd
// is app.MissingTool("clangd").
func (f Flavor) Command(map[string]string) (string, []string, error) {
	ctx := context.Background()
	status := f.cfg.Locator.Tool(ctx, cpp.ToolClangd)
	if status.Source == domain.ToolMissing || status.Path == "" {
		return "", nil, app.MissingTool(cpp.ToolClangd)
	}
	args := []string{"--background-index=false", "--header-insertion=never", "--completion-style=detailed", "--log=error"}
	if compiler, err := f.cfg.Locator.Compiler(ctx); err == nil && compiler.Family == cpp.GCC {
		args = append(args, "--query-driver="+compiler.Path)
	}
	return status.Path, args, nil
}

// RootOf is the folder of the file: a student's exercise is a folder, and a teacher's
// compile_flags.txt or compile_commands.json there wins over the fallback flags.
func (Flavor) RootOf(path string) string {
	abs, err := filepath.Abs(path)
	if err != nil {
		abs = path
	}
	return filepath.Dir(abs)
}

// InitializationOptions are the flags clangd uses when the folder has no compilation database.
func (Flavor) InitializationOptions() any {
	return map[string]any{"fallbackFlags": []string{"-std=c++17", "-Wall", "-Wextra"}}
}

// Configuration is none: clangd reads .clangd and compile_flags.txt by itself.
func (Flavor) Configuration() any { return nil }

// Environment adds nothing to the process environment.
func (Flavor) Environment() map[string]string { return nil }
