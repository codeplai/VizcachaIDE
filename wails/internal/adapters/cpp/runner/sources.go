package runner

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/codeplai/VizcachaIDE/wails/internal/app"
	"github.com/codeplai/VizcachaIDE/wails/internal/domain"
)

// ErrHeaderOnly means the user ran a header (.h .hpp .hh) in a folder without a source file.
// It wraps app.ErrUnsupported and its message carries the key errors.cppHeaderOnly, which the
// frontend shows translated. Run, Build, Check and CompileForDebug return it.
var ErrHeaderOnly = fmt.Errorf("%w: errors.cppHeaderOnly", app.ErrUnsupported)

var (
	sourceExtensions = map[string]bool{".cpp": true, ".cc": true, ".cxx": true, ".c++": true}
	headerExtensions = map[string]bool{".h": true, ".hpp": true, ".hh": true}
)

func isSource(path string) bool { return sourceExtensions[strings.ToLower(filepath.Ext(path))] }

func isHeader(path string) bool { return headerExtensions[strings.ToLower(filepath.Ext(path))] }

// Configure implements app.ProgramRunner. A folder with more than one source is a project
// (every source is compiled together) and so is a header's folder; otherwise the file runs alone.
// A header whose folder has no sources is still a project: Run answers ErrHeaderOnly.
func (r *Runner) Configure(path string, programArgs []string) domain.RunConfiguration {
	config := domain.NewFileRunConfiguration(domain.CodeLanguageCpp, path, programArgs)
	folder := config.WorkingDir
	config.Project = &domain.ProjectContext{Root: folder, Kind: domain.ProjectFolder}
	if len(listSources(folder)) > 1 || isHeader(path) {
		config.Mode = domain.RunProject
		config.Target = folder
		config.Project.Name = filepath.Base(folder)
	}
	return config
}

// listSources returns the sources of a folder (not its subfolders), sorted by name.
func listSources(folder string) []string {
	entries, err := os.ReadDir(folder)
	if err != nil {
		return nil
	}
	var sources []string
	for _, entry := range entries {
		if !entry.IsDir() && isSource(entry.Name()) {
			sources = append(sources, filepath.Join(folder, entry.Name()))
		}
	}
	sort.Strings(sources)
	return sources
}

// folderOf is where the configuration lives.
func folderOf(config domain.RunConfiguration) string {
	if config.Project != nil && config.Project.Root != "" {
		return config.Project.Root
	}
	return config.WorkingDir
}

// sourcesOf returns the files the compiler gets: the whole folder for a project, the file alone
// otherwise. A folder without sources gives ErrHeaderOnly.
func sourcesOf(config domain.RunConfiguration) ([]string, error) {
	if config.Mode == domain.RunFile && isSource(config.Target) {
		return []string{config.Target}, nil
	}
	sources := listSources(folderOf(config))
	if len(sources) == 0 {
		return nil, ErrHeaderOnly
	}
	return sources, nil
}

// programName is the name of the executable without extension: the file's for a file, the
// folder's for a project.
func programName(config domain.RunConfiguration) string {
	if config.Mode == domain.RunFile {
		base := filepath.Base(config.Target)
		return strings.TrimSuffix(base, filepath.Ext(base))
	}
	return filepath.Base(folderOf(config))
}
