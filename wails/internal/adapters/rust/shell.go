package rust

import (
	"context"
	"path/filepath"

	"github.com/codeplai/VizcachaIDE/wails/internal/domain"
)

// ShellPaths implements app.ShellPaths: the folder of cargo (configured, else CARGO_HOME/bin where
// rustup puts cargo, rustc and rustup), so the terminal finds them even when the PATH of the
// IDE does not list it.
func (l *Locator) ShellPaths(context.Context) []string {
	cargo := l.Tool(ToolCargo)
	if cargo.Source == domain.ToolMissing {
		return nil
	}
	return []string{filepath.Dir(cargo.Path)}
}
