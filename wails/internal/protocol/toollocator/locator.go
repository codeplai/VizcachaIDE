// Package toollocator finds the executables of a language's tools: the path chosen in Settings,
// then the toolchain shipped next to the IDE, then PATH.
//
// Each language passes the candidates it declares (Tool). Its own validation, such as the
// minimum Python version, happens in the language adapter before it accepts the result.
package toollocator

import (
	"os"
	"path/filepath"
	"runtime"

	"github.com/codeplai/VizcachaIDE/wails/internal/domain"
)

// Tool is what a language asks the locator to find.
type Tool struct {
	Spec     domain.ToolSpec
	Language domain.CodeLanguage
	// Name is the executable name without extension; it defaults to Spec.ID.
	Name string
	// BundledDirectories are the folders where the IDE may ship the tool, relative to the
	// folder of the IDE executable (for example "toolchain/go/bin"), in order of preference.
	BundledDirectories []string
}

// Locator resolves tools.
type Locator struct {
	appDir     string
	searchPath string
	configured func(toolID string) string
}

// New creates a locator. appDir is the folder of the IDE executable, searchPath the PATH
// variable (a list of folders) and configured returns Settings.ToolPaths[id] (it may be nil).
func New(appDir, searchPath string, configured func(toolID string) string) *Locator {
	return &Locator{appDir: appDir, searchPath: searchPath, configured: configured}
}

// Locate returns where the tool is. Source is domain.ToolMissing and Path is empty when it is
// not found anywhere. Version is left empty: the language adapter asks the tool for it.
func (l *Locator) Locate(tool Tool) domain.ToolStatus {
	status := domain.ToolStatus{
		ID: tool.Spec.ID, CodeLanguage: tool.Language, Role: tool.Spec.Role, Source: domain.ToolMissing,
	}
	name := executableName(tool)
	if l.configured != nil {
		if path := l.configured(tool.Spec.ID); isFile(path) {
			status.Source, status.Path = domain.ToolConfigured, path
			return status
		}
	}
	if path := l.bundled(tool, name); path != "" {
		status.Source, status.Path = domain.ToolBundled, path
		return status
	}
	if path := l.onPath(name); path != "" {
		status.Source, status.Path = domain.ToolOnPath, path
	}
	return status
}

// ExistingDirectories returns the folders, given relative to the IDE executable's folder, that
// exist. Languages put them in front of PATH when they run their bundled tools.
func (l *Locator) ExistingDirectories(relative ...string) []string {
	found := []string{}
	for _, dir := range relative {
		full := filepath.Join(l.appDir, filepath.FromSlash(dir))
		if info, err := os.Stat(full); err == nil && info.IsDir() {
			found = append(found, full)
		}
	}
	return found
}

// Join returns a path inside the folder of the IDE executable.
func (l *Locator) Join(relative string) string {
	return filepath.Join(l.appDir, filepath.FromSlash(relative))
}

func (l *Locator) bundled(tool Tool, name string) string {
	for _, dir := range tool.BundledDirectories {
		if candidate := filepath.Join(l.Join(dir), name); isFile(candidate) {
			return candidate
		}
	}
	return ""
}

func (l *Locator) onPath(name string) string {
	for _, dir := range filepath.SplitList(l.searchPath) {
		if dir == "" {
			continue
		}
		if candidate := filepath.Join(dir, name); isFile(candidate) {
			return candidate
		}
	}
	return ""
}

func executableName(tool Tool) string {
	name := tool.Name
	if name == "" {
		name = tool.Spec.ID
	}
	if runtime.GOOS == "windows" {
		return name + ".exe"
	}
	return name
}

func isFile(path string) bool {
	if path == "" {
		return false
	}
	info, err := os.Stat(path)
	return err == nil && !info.IsDir()
}
