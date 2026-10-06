package rust

import (
	"context"
	"os"
	"path/filepath"

	"github.com/codeplai/VizcachaIDE/wails/internal/domain"
)

// Bundled reports whether the Rust in use is the sysroot the full variant ships
// (<app>/toolchain/rust). That toolchain has no rustup proxies and may sit in a read-only folder.
func (l *Locator) Bundled() bool {
	return l.Tool(ToolCargo).Source == domain.ToolBundled
}

// ShellVariables implements app.ShellVariables: with the bundled toolchain, CARGO_HOME points to a
// folder the user can write (the crates registry and the download caches live there) unless the
// environment already sets one. The terminal and every Rust process of the IDE share it.
func (l *Locator) ShellVariables(context.Context) []string {
	if !l.Bundled() || variable(l.options.BaseEnvironment, "CARGO_HOME") != "" {
		return nil
	}
	home := l.options.CargoHome
	if home == "" {
		home = userCargoHome()
	}
	if home == "" {
		return nil
	}
	return []string{"CARGO_HOME=" + home}
}

// userCargoHome is <user cache dir>/VizcachaIDE/cargo, or "" when the system has no cache dir.
func userCargoHome() string {
	cache, err := os.UserCacheDir()
	if err != nil || cache == "" {
		return ""
	}
	return filepath.Join(cache, "VizcachaIDE", "cargo")
}
