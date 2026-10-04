package domain

import "path/filepath"

// ProjectKind says what marks the project around the file being run.
type ProjectKind string

// ProjectKind values.
const (
	ProjectGoModule  ProjectKind = "gomod"
	ProjectFolder    ProjectKind = "folder"
	ProjectPyProject ProjectKind = "pyproject"
	ProjectCargo     ProjectKind = "cargo"
)

// ProjectContext is the project found around the file being run. For Go, Name is the
// module path.
type ProjectContext struct {
	Root string      `json:"root"`
	Kind ProjectKind `json:"kind"`
	Name string      `json:"name"`
}

// RunTarget says whether to run one file or its whole project.
type RunTarget string

// RunTarget values.
const (
	RunFile    RunTarget = "file"
	RunProject RunTarget = "project"
)

// RunConfiguration is what the user wants to run: a single file or a project.
type RunConfiguration struct {
	CodeLanguage CodeLanguage    `json:"codeLanguage"`
	Target       string          `json:"target"`
	WorkingDir   string          `json:"workingDir"`
	Mode         RunTarget       `json:"mode"`
	ProgramArgs  []string        `json:"programArgs"`
	Project      *ProjectContext `json:"project"`
	// Echo is true when the program runs in a pseudoterminal, which already echoes what the
	// user types; the frontend then does not repeat the input in Output.
	Echo bool `json:"echo"`
}

// NewFileRunConfiguration builds the configuration that runs one file from its folder.
func NewFileRunConfiguration(language CodeLanguage, path string, programArgs []string) RunConfiguration {
	if programArgs == nil {
		programArgs = []string{} // JSON "[]", not "null"
	}
	return RunConfiguration{
		CodeLanguage: language,
		Target:       path,
		WorkingDir:   filepath.Dir(path),
		Mode:         RunFile,
		ProgramArgs:  programArgs,
	}
}

// FileNode is one entry of the project tree shown in the Files panel.
type FileNode struct {
	Name     string     `json:"name"`
	Path     string     `json:"path"`
	IsDir    bool       `json:"isDir"`
	Children []FileNode `json:"children"`
}
