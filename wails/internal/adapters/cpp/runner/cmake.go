package runner

import (
	"context"
	"os"
	"path/filepath"

	"github.com/codeplai/VizcachaIDE/wails/internal/adapters/cpp"
	"github.com/codeplai/VizcachaIDE/wails/internal/adapters/cpp/cmake"
	"github.com/codeplai/VizcachaIDE/wails/internal/app"
	"github.com/codeplai/VizcachaIDE/wails/internal/domain"
)

// Every C++ project is a CMake project (docs/PLAN_CPP_CMAKE.md): a saved folder without a
// CMakeLists.txt gets one the first time it is configured. Only an untitled file, which has no
// folder of its own, is still compiled directly (untitled.go).

// configureProject is the project configuration of the CMake project around path, creating
// its files when the folder has sources and no CMakeLists.txt. ok is false when the file is not
// a project's (a header alone, a folder that cannot be written).
func (r *Runner) configureProject(path string, programArgs []string) (domain.RunConfiguration, bool) {
	project, found := cmake.FindProject(path, "")
	if !found {
		folder := containingFolder(path)
		if len(listSources(folder)) == 0 {
			return domain.RunConfiguration{}, false
		}
		var err error
		if project, err = cmake.Generate(folder, r.language()); err != nil {
			return domain.RunConfiguration{}, false
		}
	}
	r.rememberActive(project.Root, path)
	config := domain.NewFileRunConfiguration(domain.CodeLanguageCpp, path, programArgs)
	config.Mode, config.Target, config.WorkingDir = domain.RunProject, project.Root, project.Root
	config.Project = &domain.ProjectContext{Root: project.Root, Kind: domain.ProjectFolder, Name: filepath.Base(project.Root)}
	return config, true
}

// projectOf is the CMake project of a configuration, as Configure resolved it.
func projectOf(config domain.RunConfiguration) (cmake.Project, bool) {
	root := folderOf(config)
	return cmake.FindProject(root, root)
}

func containingFolder(path string) string {
	if info, err := os.Stat(path); err == nil && info.IsDir() {
		return path
	}
	return filepath.Dir(path)
}

func (r *Runner) language() string {
	if r.options.Language == nil {
		return "en"
	}
	return r.options.Language()
}

// rememberActive keeps the file the user is editing in a project: with several programs in it,
// the one that contains that file is the one to run.
func (r *Runner) rememberActive(root, file string) {
	r.mu.Lock()
	defer r.mu.Unlock()
	if r.active == nil {
		r.active = map[string]string{}
	}
	r.active[root] = file
}

func (r *Runner) activeFile(root string) string {
	r.mu.Lock()
	defer r.mu.Unlock()
	return r.active[root]
}

// Builder returns the CMake builder with the tools this machine has (app.MissingTool("cmake"),
// "ninja" or "cxx" when one is missing). The library manager configures through it.
func (r *Runner) Builder(ctx context.Context) (*cmake.Builder, error) {
	tools, err := r.cmakeTools(ctx)
	return tools.builder, err
}

// cmakeTools are the builder and the compiler it uses.
type cmakeTools struct {
	builder  *cmake.Builder
	compiler cpp.Compiler
}

func (r *Runner) cmakeTools(ctx context.Context) (cmakeTools, error) {
	compiler, err := r.locator.Compiler(ctx)
	if err != nil {
		return cmakeTools{}, err
	}
	paths := map[string]string{}
	for _, id := range []string{cpp.ToolCMake, cpp.ToolNinja} {
		status := r.locator.Tool(ctx, id)
		if status.Source == domain.ToolMissing {
			return cmakeTools{}, app.MissingTool(id)
		}
		paths[id] = status.Path
	}
	builder := cmake.New(cmake.Options{
		CMake: paths[cpp.ToolCMake], Ninja: paths[cpp.ToolNinja], Compiler: compiler.Path,
		Dependencies: r.options.Dependencies, Environment: r.base,
		SupportDir: filepath.Dir(r.buildFolder("")),
	})
	return cmakeTools{builder: builder, compiler: compiler}, nil
}

// filter is the build's output filter, with the "Compiling..." notice of this runner.
func (r *Runner) filter() *cmake.Filter {
	return &cmake.Filter{Notice: r.options.CompilingNotice, Delay: r.options.NoticeDelay}
}
