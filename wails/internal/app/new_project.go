package app

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/codeplai/VizcachaIDE/wails/internal/domain"
)

// Errors of "New project". Each message is the i18n key of the text the frontend shows (the
// convention of run.chooseMember), so the key survives the wrapping with %w.
var (
	// ErrProjectNameEmpty means the project has no name.
	ErrProjectNameEmpty = errors.New("project.errorNameEmpty")
	// ErrProjectNameInvalid means the name has characters a folder cannot have, or only dots.
	ErrProjectNameInvalid = errors.New("project.errorNameInvalid")
	// ErrProjectLocationMissing means the location is empty or is not an existing folder.
	ErrProjectLocationMissing = errors.New("project.errorLocation")
	// ErrProjectFolderTaken means the project folder exists and is not empty.
	ErrProjectFolderTaken = errors.New("project.errorExists")
	// ErrProjectNoScaffold means the language cannot create projects.
	ErrProjectNoScaffold = errors.New("project.errorNoScaffold")
)

// ProjectScaffold is the starting project of a language: the files of the "Hello + keyboard
// input" template. It is pure, it does no I/O (CreateProject writes the files).
type ProjectScaffold interface {
	// Scaffold returns the files (slash-separated relative path -> content) of a project called
	// name and the path of the file to open first.
	Scaffold(name string) (files map[string]string, mainFile string)
}

// ValidateProjectName checks a project name, which is also the name of its folder.
func ValidateProjectName(name string) error {
	trimmed := strings.TrimSpace(name)
	if trimmed == "" {
		return ErrProjectNameEmpty
	}
	if strings.ContainsAny(trimmed, "/\\:*?\"<>|") || strings.Trim(trimmed, ". ") == "" || strings.HasSuffix(trimmed, ".") {
		return fmt.Errorf("%w: %q", ErrProjectNameInvalid, name)
	}
	return nil
}

// CreateProject makes the folder <location>/<name> and writes the scaffold's files in it. It
// never touches a folder that already has something in it.
func CreateProject(scaffold ProjectScaffold, location, name string) (domain.NewProject, error) {
	if scaffold == nil {
		return domain.NewProject{}, ErrProjectNoScaffold
	}
	if err := ValidateProjectName(name); err != nil {
		return domain.NewProject{}, err
	}
	if info, err := os.Stat(location); location == "" || err != nil || !info.IsDir() {
		return domain.NewProject{}, fmt.Errorf("%w: %q", ErrProjectLocationMissing, location)
	}
	name = strings.TrimSpace(name)
	root := filepath.Join(location, name)
	if err := checkFolderFree(root); err != nil {
		return domain.NewProject{}, err
	}
	files, mainFile := scaffold.Scaffold(name)
	for relative, text := range files {
		if err := writeProjectFile(filepath.Join(root, filepath.FromSlash(relative)), text); err != nil {
			return domain.NewProject{}, err
		}
	}
	return domain.NewProject{Root: root, MainFile: filepath.Join(root, filepath.FromSlash(mainFile))}, nil
}

// checkFolderFree accepts a folder that does not exist or is empty.
func checkFolderFree(root string) error {
	info, err := os.Stat(root)
	if errors.Is(err, os.ErrNotExist) {
		return nil
	}
	if err != nil {
		return fmt.Errorf("project folder: %w", err)
	}
	if !info.IsDir() {
		return fmt.Errorf("%w: %q", ErrProjectFolderTaken, root)
	}
	entries, err := os.ReadDir(root)
	if err != nil {
		return fmt.Errorf("project folder: %w", err)
	}
	if len(entries) > 0 {
		return fmt.Errorf("%w: %q", ErrProjectFolderTaken, root)
	}
	return nil
}

func writeProjectFile(path, text string) error {
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return fmt.Errorf("project folder: %w", err)
	}
	if err := os.WriteFile(path, []byte(text), 0o644); err != nil {
		return fmt.Errorf("project file: %w", err)
	}
	return nil
}
