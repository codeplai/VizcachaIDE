package cmake

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
)

// ErrNoExecutable means the project declares no executable target (or the build folder has no
// File API reply yet).
var ErrNoExecutable = errors.New("the CMake project has no executable")

// Executable is a program target of a project, as the CMake File API (codemodel-v2) reports it.
type Executable struct {
	Name    string
	Path    string   // absolute path of the program
	Sources []string // absolute paths of its source files
}

// apiFolder is where CMake reads queries and writes replies, inside a build folder.
func apiFolder(buildDir string) string {
	return filepath.Join(buildDir, ".cmake", "api", "v1")
}

// writeQuery asks CMake for the codemodel on the next configure: an empty file named after it.
func writeQuery(buildDir string) error {
	dir := filepath.Join(apiFolder(buildDir), "query")
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return fmt.Errorf("create the CMake query folder: %w", err)
	}
	file, err := os.OpenFile(filepath.Join(dir, "codemodel-v2"), os.O_CREATE|os.O_WRONLY, 0o644)
	if err != nil {
		return fmt.Errorf("write the CMake query: %w", err)
	}
	return file.Close()
}

// latestReply is the newest reply index of a build folder ("" when CMake has not answered yet).
func latestReply(buildDir string) string {
	matches, _ := filepath.Glob(filepath.Join(apiFolder(buildDir), "reply", "index-*.json"))
	if len(matches) == 0 {
		return ""
	}
	sort.Strings(matches) // the names carry a timestamp
	return matches[len(matches)-1]
}

// Executables reads the reply of the last configure of buildDir: the executable targets with
// their program and sources.
func Executables(buildDir string) ([]Executable, error) {
	index := latestReply(buildDir)
	if index == "" {
		return nil, ErrNoExecutable
	}
	reply := filepath.Dir(index)
	var listing struct {
		Objects []struct {
			Kind     string `json:"kind"`
			JSONFile string `json:"jsonFile"`
		} `json:"objects"`
	}
	if err := readJSON(index, &listing); err != nil {
		return nil, err
	}
	for _, object := range listing.Objects {
		if object.Kind == "codemodel" {
			return executablesOf(buildDir, reply, filepath.Join(reply, object.JSONFile))
		}
	}
	return nil, ErrNoExecutable
}

func executablesOf(buildDir, reply, codemodel string) ([]Executable, error) {
	var model struct {
		Paths struct {
			Source string `json:"source"`
		} `json:"paths"`
		Configurations []struct {
			Targets []struct {
				JSONFile string `json:"jsonFile"`
			} `json:"targets"`
		} `json:"configurations"`
	}
	if err := readJSON(codemodel, &model); err != nil {
		return nil, err
	}
	var found []Executable
	for _, configuration := range model.Configurations[:min(1, len(model.Configurations))] {
		for _, target := range configuration.Targets {
			executable, ok, err := readTarget(buildDir, model.Paths.Source, filepath.Join(reply, target.JSONFile))
			if err != nil {
				return nil, err
			}
			if ok {
				found = append(found, executable)
			}
		}
	}
	return found, nil
}

// readTarget reads one target file; ok is false when it is not an executable.
func readTarget(buildDir, sourceDir, path string) (Executable, bool, error) {
	var target struct {
		Name      string `json:"name"`
		Type      string `json:"type"`
		Artifacts []struct {
			Path string `json:"path"`
		} `json:"artifacts"`
		Sources []struct {
			Path string `json:"path"`
		} `json:"sources"`
	}
	if err := readJSON(path, &target); err != nil {
		return Executable{}, false, err
	}
	if target.Type != "EXECUTABLE" || len(target.Artifacts) == 0 {
		return Executable{}, false, nil
	}
	executable := Executable{Name: target.Name, Path: absolute(buildDir, target.Artifacts[0].Path)}
	for _, source := range target.Sources {
		executable.Sources = append(executable.Sources, absolute(sourceDir, source.Path))
	}
	return executable, true, nil
}

func absolute(base, path string) string {
	if filepath.IsAbs(path) {
		return filepath.Clean(path)
	}
	return filepath.Join(base, filepath.FromSlash(path))
}

func readJSON(path string, into any) error {
	data, err := os.ReadFile(path)
	if err != nil {
		return fmt.Errorf("read the CMake reply: %w", err)
	}
	if err := json.Unmarshal(data, into); err != nil {
		return fmt.Errorf("parse the CMake reply %s: %w", filepath.Base(path), err)
	}
	return nil
}

// Choose picks the program to run: the executable whose sources contain the active file, else
// the one named like the project's main target, else the first. ok is false without executables.
func Choose(executables []Executable, active, target string) (Executable, bool) {
	if len(executables) == 0 {
		return Executable{}, false
	}
	if active != "" {
		for _, executable := range executables {
			if hasSource(executable, active) {
				return executable, true
			}
		}
	}
	for _, executable := range executables {
		if executable.Name == target {
			return executable, true
		}
	}
	return executables[0], true
}

func hasSource(executable Executable, path string) bool {
	wanted := filepath.ToSlash(filepath.Clean(path))
	for _, source := range executable.Sources {
		if strings.EqualFold(filepath.ToSlash(source), wanted) {
			return true
		}
	}
	return false
}
