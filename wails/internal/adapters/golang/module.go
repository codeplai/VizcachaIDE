package golang

import (
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/codeplai/VizcachaIDE/wails/internal/domain"
)

// GoModFileName is the file that marks the root of a Go module.
const GoModFileName = "go.mod"

var moduleDirective = regexp.MustCompile(`(?m)^\s*module\s+"?([^"\s]+)"?`)

// ParseModulePath returns the module path declared in the text of a go.mod ("" if none).
func ParseModulePath(goModText string) string {
	match := moduleDirective.FindStringSubmatch(goModText)
	if match == nil {
		return ""
	}
	return match[1]
}

// FindGoModule returns the nearest module in startDir or any of its parents, or nil.
func FindGoModule(startDir string) *domain.ProjectContext {
	dir := startDir
	for {
		info, err := os.Stat(filepath.Join(dir, GoModFileName))
		if err == nil && !info.IsDir() {
			text, _ := os.ReadFile(filepath.Join(dir, GoModFileName))
			return &domain.ProjectContext{Root: dir, Kind: domain.ProjectGoModule, Name: ParseModulePath(string(text))}
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			return nil
		}
		dir = parent
	}
}

// ConfigurationForFile decides how to run the active file: "go run ." in its folder
// when it lives inside a module, "go run file.go" otherwise.
func ConfigurationForFile(path string, programArgs []string) domain.RunConfiguration {
	if absolute, err := filepath.Abs(path); err == nil {
		path = absolute
	}
	config := domain.NewFileRunConfiguration(domain.CodeLanguageGo, path, programArgs)
	module := FindGoModule(config.WorkingDir)
	if module == nil {
		return config
	}
	config.Mode = domain.RunProject
	config.Project = module
	return config
}

// GoTargetArgument is the argument passed to "go run" or "go build" from WorkingDir.
func GoTargetArgument(config domain.RunConfiguration) string {
	if config.Mode == domain.RunProject {
		return "."
	}
	return filepath.Base(config.Target)
}

// GoExecutableName is the file name "go build" produces for the configuration.
func GoExecutableName(config domain.RunConfiguration, windows bool) string {
	base := strings.TrimSuffix(filepath.Base(config.Target), filepath.Ext(config.Target))
	if config.Mode == domain.RunProject {
		base = filepath.Base(config.WorkingDir)
	}
	if windows {
		return base + ".exe"
	}
	return base
}
