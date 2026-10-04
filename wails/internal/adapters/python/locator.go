package python

import (
	"context"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"sync"

	"github.com/codeplai/VizcachaIDE/wails/internal/app"
	"github.com/codeplai/VizcachaIDE/wails/internal/domain"
	"github.com/codeplai/VizcachaIDE/wails/internal/protocol/toollocator"
)

// Where the IDE ships CPython (python-build-standalone), relative to the folder of its executable:
// python.exe at the root on Windows, bin/python3 elsewhere.
var bundledDirectories = []string{"toolchain/python", "toolchain/python/bin"}

// Interpreter is the Python the IDE runs programs and tools with.
type Interpreter struct {
	Path    string
	Version string // "3.12.14"
	Source  domain.ToolSource
	// Venv is true when it is the .venv of the program's folder.
	Venv bool
}

// Options configures a Locator. All of them are optional.
type Options struct {
	// Settings supplies Settings.ToolPaths["python"].
	Settings app.SettingsStore
	// AppDir is the folder of the executable, where toolchain/ lives. Default: the running one.
	AppDir string
	// BaseEnvironment is the "NAME=value" list Python starts from. Default: os.Environ().
	BaseEnvironment []string
	// Probe asks an interpreter its version. Default: runs it (see probe.go).
	Probe Probe
	// Launcher returns the interpreter the Windows "py -3" launcher picks, or "". Default: asks py.
	Launcher func(ctx context.Context) string
}

// Locator finds a working Python 3.10 or newer.
type Locator struct {
	options  Options
	fixed    *toollocator.Locator // configured, then bundled
	onPath   *toollocator.Locator // PATH only
	mu       sync.Mutex
	versions map[string]string // probed paths that answered
}

// NewLocator creates the locator.
func NewLocator(options Options) *Locator {
	if options.BaseEnvironment == nil {
		options.BaseEnvironment = os.Environ()
	}
	if options.Probe == nil {
		options.Probe = runProbe
	}
	if options.Launcher == nil {
		options.Launcher = func(ctx context.Context) string { return launcherInterpreter(ctx, options.Probe) }
	}
	l := &Locator{options: options, versions: map[string]string{}}
	l.fixed = toollocator.New(applicationDirectory(options.AppDir), "", l.configured)
	l.onPath = toollocator.New("", searchPath(options.BaseEnvironment), nil)
	return l
}

// Find returns the interpreter for a program in folder ("" when there is no program yet). The
// order is: configured, bundled, the folder's .venv, the "py -3" launcher (Windows), PATH
// (python3, then python). Each candidate must answer and be 3.10 or newer, which also skips the
// Microsoft Store alias of Windows. It fails with app.MissingTool("python").
func (l *Locator) Find(ctx context.Context, folder string) (Interpreter, error) {
	for _, candidate := range l.candidates(ctx, folder) {
		if version, ok := l.usable(ctx, candidate.Path); ok {
			candidate.Version = version
			return candidate, nil
		}
	}
	return Interpreter{}, app.MissingTool(ToolPython)
}

func (l *Locator) candidates(ctx context.Context, folder string) []Interpreter {
	var found []Interpreter
	add := func(status domain.ToolStatus) {
		if status.Source != domain.ToolMissing {
			found = append(found, Interpreter{Path: status.Path, Source: status.Source})
		}
	}
	for _, name := range executableNames() {
		add(l.fixed.Locate(toolNamed(name)))
	}
	if venv := venvInterpreter(folder); venv != "" {
		found = append(found, Interpreter{Path: venv, Source: domain.ToolOnPath, Venv: true})
	}
	if runtime.GOOS == "windows" {
		if path := l.options.Launcher(ctx); path != "" {
			found = append(found, Interpreter{Path: path, Source: domain.ToolOnPath})
		}
	}
	for _, name := range executableNames() {
		add(l.onPath.Locate(toolNamed(name)))
	}
	return found
}

// usable probes a candidate once and remembers the ones that answered with a good version.
func (l *Locator) usable(ctx context.Context, path string) (string, bool) {
	l.mu.Lock()
	version, known := l.versions[path]
	l.mu.Unlock()
	if known {
		return version, true
	}
	version, err := l.options.Probe(ctx, path)
	if err != nil || !supportedVersion(version) {
		return "", false
	}
	l.mu.Lock()
	l.versions[path] = version
	l.mu.Unlock()
	return version, true
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

func toolNamed(name string) toollocator.Tool {
	return toollocator.Tool{
		Spec: Profile.Tools[0], Language: domain.CodeLanguagePython, Name: name,
		BundledDirectories: bundledDirectories,
	}
}

func executableNames() []string {
	if runtime.GOOS == "windows" {
		return []string{"python"}
	}
	return []string{"python3", "python"}
}

// venvInterpreter returns the interpreter of folder/.venv, or "".
func venvInterpreter(folder string) string {
	if folder == "" {
		return ""
	}
	path := filepath.Join(folder, ".venv", "bin", "python")
	if runtime.GOOS == "windows" {
		path = filepath.Join(folder, ".venv", "Scripts", "python.exe")
	}
	if info, err := os.Stat(path); err == nil && !info.IsDir() {
		return path
	}
	return ""
}

func applicationDirectory(given string) string {
	if given != "" {
		return given
	}
	executable, err := os.Executable()
	if err != nil {
		return "."
	}
	return filepath.Dir(executable)
}
