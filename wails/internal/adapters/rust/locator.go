package rust

import (
	"context"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync"

	"github.com/codeplai/VizcachaIDE/wails/internal/app"
	"github.com/codeplai/VizcachaIDE/wails/internal/domain"
	"github.com/codeplai/VizcachaIDE/wails/internal/protocol/toollocator"
)

// ErrNoToolchain means rustup is installed but has no default toolchain: its rustc proxy answers
// with an error that asks for "rustup default". It wraps app.MissingTool("rustc"), and the
// frontend shows errors.rustupNoToolchain.
var ErrNoToolchain = errors.New("rustup has no default toolchain")

// executables maps a tool id to its executable: clippy is driven by cargo-clippy.
var executables = map[string]string{ToolClippy: "cargo-clippy"}

// bundledLldb is where full-cpp ships lldb-dap (docs/PLAN_CPP.md section 7); Rust shares it.
const bundledLldb = "toolchain/cpp/bin"

// bundledRust is the bin folder of the sysroot the full variant ships (packaging/fetch_rust.py):
// rustc, cargo, clippy-driver, cargo-clippy, rustfmt and rust-analyzer, no rustup proxies.
const bundledRust = "toolchain/rust/bin"

// Options configures a Locator. All of them are optional.
type Options struct {
	Settings        app.SettingsStore // Settings.ToolPaths["rustc"], ["cargo"], ...
	AppDir          string            // the folder of the executable (default: the running one)
	BaseEnvironment []string          // default: os.Environ()
	// CargoHome is the writable folder the bundled toolchain uses as CARGO_HOME when the
	// environment sets none (the install folder may be read-only). Default: <user cache>/VizcachaIDE/cargo.
	CargoHome string
	// Probe runs a tool with arguments and returns what it printed, stdout and stderr together
	// (default: runs it with a time limit).
	Probe func(ctx context.Context, path string, args ...string) (string, error)
}

// Locator finds the tools of the rustup toolchain and lldb-dap, and reads the toolchain once.
type Locator struct {
	options Options
	fixed   *toollocator.Locator // configured, then bundled (toolchain/rust/bin; lldb-dap in toolchain/cpp/bin)
	cargo   *toollocator.Locator // CARGO_HOME/bin, where rustup puts its proxies
	onPath  *toollocator.Locator
	mu      sync.Mutex
	found   *Toolchain
}

// NewLocator creates the locator.
func NewLocator(options Options) *Locator {
	if options.BaseEnvironment == nil {
		options.BaseEnvironment = os.Environ()
	}
	if options.Probe == nil {
		base := options.BaseEnvironment // rustup proxies need CARGO_HOME and RUSTUP_HOME
		options.Probe = func(ctx context.Context, path string, args ...string) (string, error) {
			return runTool(ctx, base, path, args...)
		}
	}
	l := &Locator{options: options}
	l.fixed = toollocator.New(applicationDirectory(options.AppDir), "", l.configured)
	l.cargo = toollocator.New("", cargoBin(options.BaseEnvironment), nil)
	l.onPath = toollocator.New("", variable(options.BaseEnvironment, "PATH"), nil)
	return l
}

// Tool finds a tool: configured → bundled (toolchain/rust/bin) → CARGO_HOME/bin → PATH. Version is empty.
func (l *Locator) Tool(id string) domain.ToolStatus {
	tool := toolFor(id)
	for _, locator := range []*toollocator.Locator{l.fixed, l.cargo, l.onPath} {
		if status := locator.Locate(tool); status.Source != domain.ToolMissing {
			return status
		}
	}
	return domain.ToolStatus{ID: id, CodeLanguage: domain.CodeLanguageRust, Role: tool.Spec.Role, Source: domain.ToolMissing}
}

// Toolchain reads the rustc in use ("rustc -vV" and "--print sysroot") and keeps the answer. A
// missing rustc is app.MissingTool("rustc"); a rustup without toolchain is ErrNoToolchain.
func (l *Locator) Toolchain(ctx context.Context) (Toolchain, error) {
	l.mu.Lock()
	defer l.mu.Unlock()
	if l.found != nil {
		return *l.found, nil
	}
	rustc := l.Tool(ToolRustc)
	if rustc.Source == domain.ToolMissing {
		return Toolchain{}, app.MissingTool(ToolRustc)
	}
	verbose, err := l.options.Probe(ctx, rustc.Path, "-vV")
	if err != nil {
		if strings.Contains(verbose, "rustup") {
			return Toolchain{}, fmt.Errorf("%w: %w", ErrNoToolchain, app.MissingTool(ToolRustc))
		}
		return Toolchain{}, fmt.Errorf("%w: %w", err, app.MissingTool(ToolRustc))
	}
	found, err := parseVerbose(verbose)
	if err != nil {
		return Toolchain{}, err
	}
	sysroot, err := l.options.Probe(ctx, rustc.Path, "--print", "sysroot")
	if err != nil {
		return Toolchain{}, fmt.Errorf("rustc --print sysroot: %w", err)
	}
	found.Rustc, found.Sysroot = rustc.Path, strings.TrimSpace(sysroot)
	l.found = &found
	return found, nil
}

// Environment is the environment of every Rust process: Environment(base) with the folder of
// cargo first in PATH, so cargo, rustc and the proxies find each other, and on a Windows GNU host
// the linker and dlltool cargo must use (windowsgnu.go).
func (l *Locator) Environment() []string {
	env := append(Environment(l.options.BaseEnvironment), l.gnuVariables()...)
	env = append(env, l.ShellVariables(context.Background())...)
	cargo := l.Tool(ToolCargo)
	if cargo.Source == domain.ToolMissing {
		return env
	}
	return withPathFirst(env, filepath.Dir(cargo.Path))
}

func (l *Locator) configured(toolID string) string {
	if l.options.Settings == nil {
		return ""
	}
	settings, err := l.options.Settings.Load()
	if err != nil {
		return ""
	}
	return strings.TrimSpace(settings.ToolPaths[toolID])
}

func toolFor(id string) toollocator.Tool {
	tool := toollocator.Tool{Language: domain.CodeLanguageRust, Name: executables[id], BundledDirectories: []string{bundledRust}}
	if id == ToolLldbDap {
		tool.BundledDirectories = []string{bundledLldb}
	}
	for _, spec := range Profile.Tools {
		if spec.ID == id {
			tool.Spec = spec
		}
	}
	return tool
}
