package cpp

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

// Where full-cpp ships its tools, relative to the folder of the executable: llvm-mingw and Ninja
// in bin, CMake in its own folder, vcpkg in its root.
const (
	bundledBin   = "toolchain/cpp/bin"
	bundledCMake = "toolchain/cpp/cmake/bin"
	bundledVcpkg = "toolchain/cpp/vcpkg"
)

// compilerNames is the PATH order: the courses use g++; clang++ and c++ cover LLVM and macOS.
var compilerNames = []string{"g++", "clang++", "c++"}

// Compiler is the C++ compiler the IDE builds with.
type Compiler struct {
	Path    string
	Family  Family
	Version string // first line of --version
	Source  domain.ToolSource
}

// Options configures a Locator. All of them are optional.
type Options struct {
	Settings        app.SettingsStore // Settings.ToolPaths["cxx"], ["lldb-dap"], ...
	AppDir          string            // the folder of the executable (default: the running one)
	BaseEnvironment []string          // default: os.Environ()
	// Probe runs "<compiler> --version" and returns its output (default: runs it).
	Probe func(ctx context.Context, path string) (string, error)
	// XcodeReady reports whether macOS has the Command Line Tools (xcode-select -p). Calling
	// clang++ without them opens an install dialog, so they are checked first.
	XcodeReady func() bool
}

// Locator finds the compiler and the LLVM tools.
type Locator struct {
	options Options
	fixed   *toollocator.Locator // configured, then bundled
	onPath  *toollocator.Locator // PATH only
	mu      sync.Mutex
	found   *Compiler
}

// NewLocator creates the locator.
func NewLocator(options Options) *Locator {
	if options.BaseEnvironment == nil {
		options.BaseEnvironment = os.Environ()
	}
	if options.Probe == nil {
		options.Probe = runVersion
	}
	if options.XcodeReady == nil {
		options.XcodeReady = xcodeReady
	}
	l := &Locator{options: options}
	l.fixed = toollocator.New(applicationDirectory(options.AppDir), "", l.configured)
	l.onPath = toollocator.New("", searchPath(options.BaseEnvironment), nil)
	return l
}

// Compiler returns the compiler (configured → bundled → PATH g++, clang++, c++), probing each
// candidate with --version. It fails with app.MissingTool("cxx").
func (l *Locator) Compiler(ctx context.Context) (Compiler, error) {
	if runtime.GOOS == "darwin" && !l.options.XcodeReady() {
		return Compiler{}, app.MissingTool(ToolCompiler)
	}
	for _, candidate := range l.compilerCandidates() {
		output, err := l.options.Probe(ctx, candidate.Path)
		if err != nil {
			continue
		}
		candidate.Family = FamilyOf(candidate.Path, output)
		candidate.Version = strings.TrimSpace(strings.SplitN(output, "\n", 2)[0])
		l.mu.Lock()
		l.found = &candidate
		l.mu.Unlock()
		return candidate, nil
	}
	return Compiler{}, app.MissingTool(ToolCompiler)
}

func (l *Locator) compilerCandidates() []Compiler {
	var found []Compiler
	add := func(status domain.ToolStatus) {
		if status.Source != domain.ToolMissing {
			found = append(found, Compiler{Path: status.Path, Source: status.Source})
		}
	}
	add(l.fixed.Locate(toolFor(ToolCompiler, "clang++")))
	for _, name := range compilerNames {
		add(l.onPath.Locate(toolFor(ToolCompiler, name)))
	}
	return found
}

// Tool finds lldb-dap, clangd, clang-format, cmake, ninja or vcpkg: configured → bundled → next to the compiler (the
// same bin folder of an LLVM or MinGW install) → PATH. Version is left empty.
func (l *Locator) Tool(ctx context.Context, id string) domain.ToolStatus {
	tool := toolFor(id, id)
	if status := l.fixed.Locate(tool); status.Source != domain.ToolMissing {
		return status
	}
	if compiler, err := l.Compiler(ctx); err == nil {
		if path := beside(compiler.Path, id); path != "" {
			return domain.ToolStatus{ID: id, CodeLanguage: domain.CodeLanguageCpp, Role: tool.Spec.Role, Source: compiler.Source, Path: path}
		}
	}
	return l.onPath.Locate(tool)
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

func toolFor(id, name string) toollocator.Tool {
	tool := toollocator.Tool{Language: domain.CodeLanguageCpp, Name: name, BundledDirectories: bundledDirectories(id)}
	for _, spec := range Profile.Tools {
		if spec.ID == id {
			tool.Spec = spec
		}
	}
	return tool
}

// beside returns the executable name next to the compiler, or "".
func beside(compilerPath, name string) string {
	if runtime.GOOS == "windows" {
		name += ".exe"
	}
	path := filepath.Join(filepath.Dir(compilerPath), name)
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

// bundledDirectories are where the installer may put a tool, in order of preference.
func bundledDirectories(id string) []string {
	switch id {
	case ToolCMake:
		return []string{bundledCMake}
	case ToolNinja:
		return []string{bundledCMake, bundledBin}
	case ToolVcpkg:
		return []string{bundledVcpkg}
	}
	return []string{bundledBin}
}
