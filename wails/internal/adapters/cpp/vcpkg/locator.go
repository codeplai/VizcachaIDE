// Package vcpkg brings C++ libraries to the IDE: it finds vcpkg, prepares its environment and
// cache, searches its ports without internet, and adds or removes a library in the project
// (vcpkg.json plus the marked block of CMakeLists.txt). The project is built by adapters/cpp/cmake,
// which asks Setup for the arguments and the environment; vcpkg itself runs inside "cmake -S".
package vcpkg

import (
	"os"
	"path/filepath"
	"runtime"
	"strings"

	"github.com/codeplai/VizcachaIDE/wails/internal/app"
)

const (
	// ToolID is the ID of vcpkg in Settings.ToolPaths (adapters/cpp/profile.go declares it).
	ToolID = "vcpkg"
	// bundledRoot and bundledSeed are where full-cpp ships vcpkg, relative to the executable.
	bundledRoot = "toolchain/cpp/vcpkg"
	bundledSeed = "toolchain/cpp/vcpkg-seed"
	// bundledCompilerBin is where full-cpp ships llvm-mingw.
	bundledCompilerBin = "toolchain/cpp/bin"
)

// LocatorOptions configures a Locator. All of them are optional.
type LocatorOptions struct {
	Settings        app.SettingsStore // Settings.ToolPaths["vcpkg"]
	AppDir          string            // the folder of the executable (default: the running one)
	BaseEnvironment []string          // default: os.Environ()
	// SeedDir overrides the bundled seed folder (tests).
	SeedDir string
	// CacheDir is where downloads, binary archives and build trees live (default:
	// <UserCacheDir>/VizcachaIDE).
	CacheDir string
	// CompilerBin returns the folder of the C++ compiler. On Windows it goes first on PATH:
	// vcpkg's mingw toolchain looks for x86_64-w64-mingw32-g++ there (default: the bundled one).
	CompilerBin func() string
	// ToolBins returns the folders of cmake and ninja. vcpkg looks for cmake on PATH, and when it
	// is not there it downloads its own copy; with them first on PATH it uses the IDE's.
	ToolBins func() []string
	// GOOS and GOARCH choose the triplet (default: the running ones). Tests set them.
	GOOS, GOARCH string
}

// pathFolders are the existing tool folders that go on PATH.
func (o LocatorOptions) pathFolders() []string {
	if o.ToolBins == nil {
		return nil
	}
	var folders []string
	for _, folder := range o.ToolBins() {
		if folder != "" {
			folders = append(folders, folder)
		}
	}
	return folders
}

// Locator finds vcpkg, its seed and the triplet.
type Locator struct{ options LocatorOptions }

// NewLocator creates the locator.
func NewLocator(options LocatorOptions) *Locator {
	if options.BaseEnvironment == nil {
		options.BaseEnvironment = os.Environ()
	}
	if options.GOOS == "" {
		options.GOOS = runtime.GOOS
	}
	if options.GOARCH == "" {
		options.GOARCH = runtime.GOARCH
	}
	if options.AppDir == "" {
		options.AppDir = applicationDirectory()
	}
	return &Locator{options: options}
}

// Root returns the vcpkg root (configured, then bundled, then VCPKG_ROOT) or "" when none is
// valid: a valid root has .vcpkg-root and the vcpkg executable.
func (l *Locator) Root() string {
	candidates := []string{
		l.configured(),
		filepath.Join(l.options.AppDir, filepath.FromSlash(bundledRoot)),
		environmentValue(l.options.BaseEnvironment, "VCPKG_ROOT"),
	}
	for _, candidate := range candidates {
		if l.valid(candidate) {
			return candidate
		}
	}
	return ""
}

// Seed returns the folder with the downloads vcpkg needs, shipped with the installer ("" if
// there is none).
func (l *Locator) Seed() string {
	if l.options.SeedDir != "" {
		return l.options.SeedDir
	}
	seed := filepath.Join(l.options.AppDir, filepath.FromSlash(bundledSeed))
	if info, err := os.Stat(seed); err == nil && info.IsDir() {
		return seed
	}
	return ""
}

// Triplet is the vcpkg triplet of the target and of the host. On Windows both are
// x64-mingw-static, so no MSVC is needed.
func (l *Locator) Triplet() string {
	arm := l.options.GOARCH == "arm64"
	switch l.options.GOOS {
	case "windows":
		return "x64-mingw-static"
	case "darwin":
		if arm {
			return "arm64-osx"
		}
		return "x64-osx"
	default:
		if arm {
			return "arm64-linux"
		}
		return "x64-linux"
	}
}

// CacheRoot is <cache>/vcpkg, where the IDE keeps what vcpkg downloads and builds.
func (l *Locator) CacheRoot() string {
	return filepath.Join(l.cacheDir(), "vcpkg")
}

func (l *Locator) cacheDir() string {
	if l.options.CacheDir != "" {
		return l.options.CacheDir
	}
	if cache, err := os.UserCacheDir(); err == nil {
		return filepath.Join(cache, "VizcachaIDE")
	}
	return filepath.Join(os.TempDir(), "VizcachaIDE")
}

func (l *Locator) compilerBin() string {
	if l.options.CompilerBin != nil {
		return l.options.CompilerBin()
	}
	bin := filepath.Join(l.options.AppDir, filepath.FromSlash(bundledCompilerBin))
	if info, err := os.Stat(bin); err == nil && info.IsDir() {
		return bin
	}
	return ""
}

func (l *Locator) configured() string {
	if l.options.Settings == nil {
		return ""
	}
	settings, err := l.options.Settings.Load()
	if err != nil {
		return ""
	}
	return strings.TrimSpace(settings.ToolPaths[ToolID])
}

func (l *Locator) valid(root string) bool {
	if root == "" {
		return false
	}
	executable := "vcpkg"
	if l.options.GOOS == "windows" {
		executable += ".exe"
	}
	return isFile(filepath.Join(root, ".vcpkg-root")) && isFile(filepath.Join(root, executable))
}

func isFile(path string) bool {
	info, err := os.Stat(path)
	return err == nil && !info.IsDir()
}

func applicationDirectory() string {
	executable, err := os.Executable()
	if err != nil {
		return "."
	}
	return filepath.Dir(executable)
}

// environmentValue returns the value of a variable in a "KEY=value" list, "" if absent.
func environmentValue(environment []string, key string) string {
	for index := len(environment) - 1; index >= 0; index-- {
		name, value, found := strings.Cut(environment[index], "=")
		if found && strings.EqualFold(name, key) {
			return value
		}
	}
	return ""
}
