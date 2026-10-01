package domain

import (
	"path/filepath"
	"strings"
)

// GoModule is a Go module found around the file being run.
type GoModule struct {
	Root       string `json:"root"`
	ModulePath string `json:"modulePath"`
}

// RunTarget says whether to run one file or a whole package.
type RunTarget string

// RunTarget values.
const (
	RunFile    RunTarget = "file"
	RunPackage RunTarget = "package"
)

// RunConfiguration is what the user wants to run: a single file or a Go package.
type RunConfiguration struct {
	Target      string    `json:"target"`
	WorkingDir  string    `json:"workingDir"`
	Mode        RunTarget `json:"mode"`
	ProgramArgs []string  `json:"programArgs"`
	Module      *GoModule `json:"module"`
}

// NewFileRunConfiguration builds the configuration that runs one file from its folder.
func NewFileRunConfiguration(path string, programArgs []string) RunConfiguration {
	if programArgs == nil {
		programArgs = []string{} // JSON "[]", not "null"
	}
	return RunConfiguration{
		Target:      path,
		WorkingDir:  filepath.Dir(path),
		Mode:        RunFile,
		ProgramArgs: programArgs,
	}
}

// GoTargetArgument is the argument passed to "go run" or "go build" from WorkingDir.
func (c RunConfiguration) GoTargetArgument() string {
	if c.Mode == RunPackage {
		return "."
	}
	return filepath.Base(c.Target)
}

// ExecutableName is the file name "go build" produces for this configuration.
func (c RunConfiguration) ExecutableName(windows bool) string {
	base := strings.TrimSuffix(filepath.Base(c.Target), filepath.Ext(c.Target))
	if c.Mode == RunPackage {
		base = filepath.Base(c.WorkingDir)
	}
	if windows {
		return base + ".exe"
	}
	return base
}

// FileNode is one entry of the project tree shown in the Files panel.
type FileNode struct {
	Name     string     `json:"name"`
	Path     string     `json:"path"`
	IsDir    bool       `json:"isDir"`
	Children []FileNode `json:"children"`
}

// ToolchainInfo describes the Go tools the IDE found. Empty means "not found".
type ToolchainInfo struct {
	GoVersion    string `json:"goVersion"`
	DelveVersion string `json:"delveVersion"`
	GoplsVersion string `json:"goplsVersion"`
}
