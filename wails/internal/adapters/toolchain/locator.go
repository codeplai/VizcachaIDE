package toolchain

import (
	"os"
	"path/filepath"
	"runtime"
)

// Tool names the IDE looks for.
const (
	ToolGo    = "go"
	ToolDelve = "dlv"
	ToolGopls = "gopls"
)

// Origin says where a tool was found.
type Origin string

// Origin values.
const (
	OriginConfigured Origin = "configured"
	OriginBundled    Origin = "bundled"
	OriginPath       Origin = "path"
	OriginMissing    Origin = "missing"
)

const bundleDirectory = "toolchain"

// Location is where one tool comes from. Path is empty when the origin is OriginMissing.
type Location struct {
	Tool   string
	Origin Origin
	Path   string
}

// Locator finds Go tools: configured path, then the toolchain shipped next to the
// executable (toolchain/go/bin, toolchain/bin), then PATH.
type Locator struct {
	appDir     string
	searchPath string
	configured func(tool string) string
}

// NewLocator creates a locator. configured may be nil.
func NewLocator(appDir, searchPath string, configured func(tool string) string) *Locator {
	return &Locator{appDir: appDir, searchPath: searchPath, configured: configured}
}

// BundledGoroot is the GOROOT of the toolchain shipped with the app.
func (l *Locator) BundledGoroot() string {
	return filepath.Join(l.appDir, bundleDirectory, "go")
}

// BundledBinDirectories lists the shipped folders that exist, to put in front of PATH.
func (l *Locator) BundledBinDirectories() []string {
	candidates := []string{
		filepath.Join(l.BundledGoroot(), "bin"),
		filepath.Join(l.appDir, bundleDirectory, "bin"),
	}
	found := []string{}
	for _, dir := range candidates {
		if info, err := os.Stat(dir); err == nil && info.IsDir() {
			found = append(found, dir)
		}
	}
	return found
}

// Locate resolves one tool.
func (l *Locator) Locate(tool string) Location {
	if l.configured != nil {
		if path := l.configured(tool); isFile(path) {
			return Location{Tool: tool, Origin: OriginConfigured, Path: path}
		}
	}
	if path := l.bundledPath(tool); isFile(path) {
		return Location{Tool: tool, Origin: OriginBundled, Path: path}
	}
	if path := l.onPath(tool); path != "" {
		return Location{Tool: tool, Origin: OriginPath, Path: path}
	}
	return Location{Tool: tool, Origin: OriginMissing}
}

func (l *Locator) bundledPath(tool string) string {
	dir := filepath.Join(l.appDir, bundleDirectory, "bin")
	if tool == ToolGo {
		dir = filepath.Join(l.BundledGoroot(), "bin")
	}
	return filepath.Join(dir, executableName(tool))
}

func (l *Locator) onPath(tool string) string {
	for _, dir := range filepath.SplitList(l.searchPath) {
		if dir == "" {
			continue
		}
		if candidate := filepath.Join(dir, executableName(tool)); isFile(candidate) {
			return candidate
		}
	}
	return ""
}

func executableName(tool string) string {
	if runtime.GOOS == "windows" {
		return tool + ".exe"
	}
	return tool
}

func isFile(path string) bool {
	if path == "" {
		return false
	}
	info, err := os.Stat(path)
	return err == nil && !info.IsDir()
}
