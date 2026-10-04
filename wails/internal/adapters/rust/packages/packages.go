// Package packages is the Rust package manager: it implements app.PackageManager with Cargo
// ("cargo init", "add", "remove" and "tree"), run through the shared process supervisor so the
// commands emit run events. Cargo has no tidy: that verb returns app.ErrUnsupported.
package packages

import (
	"context"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"unicode"

	"golang.org/x/text/runes"
	"golang.org/x/text/transform"
	"golang.org/x/text/unicode/norm"

	"github.com/codeplai/VizcachaIDE/wails/internal/adapters/rust"
	"github.com/codeplai/VizcachaIDE/wails/internal/adapters/rust/runner"
	"github.com/codeplai/VizcachaIDE/wails/internal/app"
	"github.com/codeplai/VizcachaIDE/wails/internal/protocol/process"
)

var (
	// ErrNeedsProject means Add, Remove or List ran outside a Cargo crate (or on the root of a
	// virtual workspace, which has no crate of its own). It wraps app.ErrUnsupported and carries
	// the key errors.rustNeedsProject; the dialog then offers Init first.
	ErrNeedsProject = fmt.Errorf("%w: errors.rustNeedsProject", app.ErrUnsupported)
	// ErrAlreadyProject means Init found a Cargo.toml in the folder.
	ErrAlreadyProject = errors.New("the folder already has a Cargo.toml")
)

// Manager is the Rust adapter of app.PackageManager.
type Manager struct {
	supervisor *process.Supervisor
	runner     *runner.Runner
}

var _ app.PackageManager = (*Manager)(nil)

// New creates the package manager. It shares the supervisor with the runner and takes from it
// the location of cargo and its environment.
func New(supervisor *process.Supervisor, rustRunner *runner.Runner) *Manager {
	return &Manager{supervisor: supervisor, runner: rustRunner}
}

// Init implements app.PackageManager: "cargo init --vcs none --name <name>" in dir. The name is
// the given one or the folder's, made a valid crate name.
func (m *Manager) Init(ctx context.Context, dir, name string) error {
	if _, err := os.Stat(filepath.Join(dir, "Cargo.toml")); err == nil {
		return ErrAlreadyProject
	}
	if strings.TrimSpace(name) == "" {
		name = filepath.Base(dir)
	}
	return m.start(ctx, dir, []string{"init", "--vcs", "none", "--name", CrateName(name)})
}

// Add implements app.PackageManager: "cargo add <pkg>" on the crate of dir ("serde",
// "serde@1").
func (m *Manager) Add(ctx context.Context, dir, pkg string) error {
	name, err := app.SingleWordArgument(pkg, "package")
	if err != nil {
		return err
	}
	return m.onCrate(ctx, dir, "add", name)
}

// Remove implements app.PackageManager: "cargo remove <pkg>" on the crate of dir.
func (m *Manager) Remove(ctx context.Context, dir, pkg string) error {
	name, err := app.SingleWordArgument(pkg, "package")
	if err != nil {
		return err
	}
	return m.onCrate(ctx, dir, "remove", name)
}

// List implements app.PackageManager: "cargo tree --depth 1" on the crate of dir.
func (m *Manager) List(ctx context.Context, dir string) error {
	return m.onCrate(ctx, dir, "tree", "--depth", "1")
}

// Tidy implements app.PackageManager.
func (m *Manager) Tidy(context.Context, string) error { return app.ErrUnsupported }

// onCrate runs "cargo <verb> --manifest-path <Cargo.toml of the crate> <rest>" from the crate's
// folder. In a workspace the crate is the member that holds dir (or the file dir names).
func (m *Manager) onCrate(ctx context.Context, dir, verb string, rest ...string) error {
	anchor := dir
	if info, err := os.Stat(dir); err == nil && info.IsDir() {
		anchor = filepath.Join(dir, "Cargo.toml")
	}
	project, ok, err := rust.FindProject(anchor)
	if err != nil {
		return fmt.Errorf("read Cargo.toml: %w", err)
	}
	if !ok || project.Virtual {
		return ErrNeedsProject
	}
	args := append([]string{verb, "--manifest-path", project.Manifest}, rest...)
	return m.start(ctx, project.Root, args)
}

func (m *Manager) start(ctx context.Context, dir string, args []string) error {
	job, err := m.runner.CommandJob(dir, args)
	if err != nil {
		return err
	}
	if err := m.supervisor.Start(ctx, job); err != nil {
		return fmt.Errorf("cargo %s: %w", args[0], err)
	}
	return nil
}

// reserved are names cargo refuses for a crate: Rust keywords and the libraries it ships.
var reserved = map[string]bool{
	"test": true, "core": true, "std": true, "alloc": true, "proc_macro": true, "self": true,
	"crate": true, "super": true, "fn": true, "let": true, "mod": true, "use": true, "type": true,
	"struct": true, "enum": true, "impl": true, "trait": true, "match": true, "loop": true,
	"while": true, "for": true, "if": true, "else": true, "return": true, "move": true, "ref": true,
	"static": true, "const": true, "pub": true, "mut": true, "in": true, "as": true, "where": true,
	"async": true, "await": true, "dyn": true, "unsafe": true, "extern": true, "break": true,
	"continue": true, "true": true, "false": true,
}

// CrateName turns a folder name into a valid crate name: lower case letters, digits, "_" and
// "-", not starting with a digit and never a Rust keyword ("Mi Programa" gives "mi_programa").
func CrateName(folder string) string {
	var name strings.Builder
	for _, letter := range strings.ToLower(withoutAccents(folder)) {
		switch {
		case letter < unicode.MaxASCII && (unicode.IsLetter(letter) || unicode.IsDigit(letter) || letter == '-'):
			name.WriteRune(letter)
		default:
			name.WriteRune('_')
		}
	}
	clean := strings.Trim(name.String(), "_-")
	if clean == "" {
		return "app"
	}
	if unicode.IsDigit(rune(clean[0])) || reserved[clean] {
		return "app_" + clean
	}
	return clean
}

// withoutAccents turns "Ñandú" into "Nandu": cargo accepts only ASCII crate names, and a student's
// folder name should still be recognisable.
func withoutAccents(text string) string {
	plain, _, err := transform.String(transform.Chain(norm.NFD, runes.Remove(runes.In(unicode.Mn)), norm.NFC), text)
	if err != nil {
		return text
	}
	return plain
}
