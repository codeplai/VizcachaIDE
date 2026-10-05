package vcpkg

import (
	"context"
	"fmt"
	"sync"
	"sync/atomic"

	"github.com/codeplai/VizcachaIDE/wails/internal/app"
	"github.com/codeplai/VizcachaIDE/wails/internal/protocol/process"
)

// ErrNeedsProject means a package verb ran outside a CMake project. It wraps app.ErrUnsupported and
// carries the key errors.cppNeedsProject.
var ErrNeedsProject = fmt.Errorf("%w: errors.cppNeedsProject", app.ErrUnsupported)

// Configurer configures the CMake project of root, which makes vcpkg install the libraries of its
// vcpkg.json into build/vcpkg_installed. It is *cmake.Builder; the output is what the commands
// printed.
type Configurer interface {
	Configure(ctx context.Context, root string) (string, error)
}

// ManagerOptions configures a Manager.
type ManagerOptions struct {
	Configurer Configurer
	// Events receives run:started, run:output and run:finished, like the package commands of the
	// other languages (the shared EventSink satisfies it).
	Events process.JobEvents
	// SlowNotice returns the already translated text printed before a long install ("Compiling
	// the library, the first time takes a few minutes"). Nil prints the English text.
	SlowNotice func() string
}

// Manager is the C++ adapter of app.PackageManager: the libraries of vcpkg.json.
type Manager struct {
	locator *Locator
	options ManagerOptions
	busy    atomic.Bool
	mu      sync.Mutex
	cancel  context.CancelFunc
}

var _ app.PackageManager = (*Manager)(nil)

const defaultSlowNotice = "Compiling the library. The first time it takes a few minutes; later ones are instant."

// NewManager creates the package manager.
func NewManager(locator *Locator, options ManagerOptions) *Manager {
	return &Manager{locator: locator, options: options}
}

// Init implements app.PackageManager: a C++ project is created by the scaffold, not here.
func (m *Manager) Init(context.Context, string, string) error { return app.ErrUnsupported }

// Tidy implements app.PackageManager.
func (m *Manager) Tidy(context.Context, string) error { return app.ErrUnsupported }

// Add implements app.PackageManager: it adds the port to vcpkg.json, lets CMake install it, and
// writes the find_package and target_link_libraries lines of the library in CMakeLists.txt. If
// anything fails both files go back to what they were: a port that does not build must not break
// the project.
func (m *Manager) Add(ctx context.Context, dir, pkg string) error {
	port, err := app.SingleWordArgument(pkg, "package")
	if err != nil {
		return err
	}
	root, err := projectRoot(dir)
	if err != nil {
		return err
	}
	return m.start(ctx, root, "vcpkg add "+port, func(ctx context.Context, say func(string)) error {
		return m.add(ctx, root, port, say)
	})
}

// Remove implements app.PackageManager: it takes the port out of vcpkg.json and of the block of
// CMakeLists.txt, and configures again so build/vcpkg_installed drops it.
func (m *Manager) Remove(ctx context.Context, dir, pkg string) error {
	port, err := app.SingleWordArgument(pkg, "package")
	if err != nil {
		return err
	}
	root, err := projectRoot(dir)
	if err != nil {
		return err
	}
	return m.start(ctx, root, "vcpkg remove "+port, func(ctx context.Context, say func(string)) error {
		return m.remove(ctx, root, port, say)
	})
}

// List implements app.PackageManager: the libraries of vcpkg.json with the installed version.
func (m *Manager) List(ctx context.Context, dir string) error {
	root, err := projectRoot(dir)
	if err != nil {
		return err
	}
	return m.start(ctx, root, "vcpkg list", func(_ context.Context, say func(string)) error {
		ports, err := Dependencies(root)
		if err != nil {
			return err
		}
		versions := installedVersions(root)
		for _, port := range ports {
			say(port + " " + versions[port] + "\n")
		}
		return nil
	})
}

func (m *Manager) add(ctx context.Context, root, port string, say func(string)) (err error) {
	restore, err := snapshot(root)
	if err != nil {
		return err
	}
	defer func() {
		if err != nil {
			restore()
		}
	}()
	if _, err = AddDependency(root, port); err != nil {
		return err
	}
	say(m.slowNotice() + "\n")
	if err = m.configure(ctx, root, say); err != nil {
		return err
	}
	lines := UsageLines(root, m.locator.Triplet(), port, targetName(root))
	if len(lines) == 0 {
		return nil
	}
	if err = SetLibraryLines(root, port, lines); err != nil {
		return err
	}
	return m.configure(ctx, root, say)
}

func (m *Manager) remove(ctx context.Context, root, port string, say func(string)) (err error) {
	restore, err := snapshot(root)
	if err != nil {
		return err
	}
	defer func() {
		if err != nil {
			restore()
		}
	}()
	if _, err = RemoveDependency(root, port); err != nil {
		return err
	}
	if err = RemoveLibraryLines(root, port); err != nil {
		return err
	}
	return m.configure(ctx, root, say)
}

func (m *Manager) configure(ctx context.Context, root string, say func(string)) error {
	output, err := m.options.Configurer.Configure(ctx, root)
	if output != "" {
		say(output)
	}
	return err
}
