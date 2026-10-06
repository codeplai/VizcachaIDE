package cmake

import (
	"context"
	"path/filepath"
	"runtime"
	"strings"
)

// Dependencies is what the libraries of a project need from the build. The vcpkg package
// implements it; a nil Dependencies is a plain CMake project without libraries.
type Dependencies interface {
	// ConfigureArgs are the extra arguments of the configure command (the toolchain file, the
	// triplets and where vcpkg installs).
	ConfigureArgs(root string) []string
	// Environment is base plus what the tools of the libraries need (VCPKG_ROOT, caches, PATH).
	Environment(base []string) []string
	// Prepare readies what has to exist before the first configure (vcpkg's downloaded tools).
	// It is idempotent and cheap after the first time.
	Prepare(ctx context.Context) error
}

// Options configures a Builder.
type Options struct {
	CMake    string // path of cmake
	Ninja    string // path of ninja
	Compiler string // path of clang++ or g++
	// Dependencies supplies the libraries' arguments and environment; nil for none.
	Dependencies Dependencies
	// Environment is the "NAME=value" list the tools start from. Nil means no variables.
	Environment []string
	// SupportDir is where the IDE's CMake script and helper are written (any writable folder).
	SupportDir string
}

// Builder configures and builds the CMake projects.
type Builder struct{ options Options }

// New creates a Builder.
func New(options Options) *Builder { return &Builder{options: options} }

// Command is one process the build runs: configure or build.
type Command struct {
	Path string
	Args []string
	Dir  string
	Env  []string
}

// Plan is what a build needs to run: the configure command, nil when the build folder is up to
// date, and the build command.
type Plan struct {
	Root      string
	Release   bool
	BuildDir  string
	Configure *Command
	Build     Command
}

// BuildDir is the build folder of a project: build/ for Debug, build/release for Release.
func BuildDir(root string, release bool) string {
	if release {
		return filepath.Join(root, "build", "release")
	}
	return filepath.Join(root, "build")
}

// Plan prepares a build of the project in root: it asks CMake for the File API reply, readies
// the libraries' tools when a configure is needed and returns the commands to run.
func (b *Builder) Plan(ctx context.Context, root string, release bool) (Plan, error) {
	dir := BuildDir(root, release)
	stale := needsConfigure(root, dir) // before the query: it may delete a build folder of another place
	if err := writeQuery(dir); err != nil {
		return Plan{}, err
	}
	env := b.environment()
	plan := Plan{Root: root, Release: release, BuildDir: dir}
	plan.Build = Command{Path: b.options.CMake, Args: []string{"--build", dir}, Dir: root, Env: env}
	if !stale {
		return plan, nil
	}
	if b.options.Dependencies != nil {
		if err := b.options.Dependencies.Prepare(ctx); err != nil {
			return Plan{}, err
		}
	}
	plan.Configure = &Command{Path: b.options.CMake, Args: b.configureArgs(root, dir, release), Dir: root, Env: env}
	return plan, nil
}

// configureArgs is the configure command of a build folder.
func (b *Builder) configureArgs(root, dir string, release bool) []string {
	kind := "Debug"
	if release {
		kind = "Release"
	}
	args := []string{
		"-S", root, "-B", dir, "-G", "Ninja",
		"-DCMAKE_BUILD_TYPE=" + kind,
		"-DCMAKE_CXX_COMPILER=" + b.options.Compiler,
		"-DCMAKE_MAKE_PROGRAM=" + b.options.Ninja,
	}
	if runtime.GOOS == "windows" {
		// The program must run outside the IDE and under the debugger without the compiler's DLLs
		// (libc++ and libunwind for llvm-mingw, libstdc++ for GCC): 0xc0000135 otherwise.
		args = append(args, "-DCMAKE_EXE_LINKER_FLAGS=-static")
	}
	if include := Include(b.options.SupportDir); include != "" {
		args = append(args, "-DCMAKE_PROJECT_TOP_LEVEL_INCLUDES="+filepath.ToSlash(include))
	}
	if b.options.Dependencies != nil {
		args = append(args, b.options.Dependencies.ConfigureArgs(root)...)
	}
	return args
}

// environment starts from the base with the compiler's folder first on PATH (llvm-mingw's
// clang++ finds its linker and runtime next to itself, and vcpkg's mingw toolchain looks for
// the compiler by name), then lets the libraries add theirs.
func (b *Builder) environment() []string {
	env := prependPath(b.options.Environment, filepath.Dir(b.options.Compiler))
	if b.options.Dependencies != nil {
		return b.options.Dependencies.Environment(env)
	}
	return env
}

// prependPath returns env with dir at the front of PATH ("Path" on Windows).
func prependPath(env []string, dir string) []string {
	result := make([]string, 0, len(env)+1)
	found := false
	for _, entry := range env {
		name, value, ok := strings.Cut(entry, "=")
		if ok && name != "" && isPathName(name) {
			entry, found = name+"="+dir+string(filepath.ListSeparator)+value, true
		}
		result = append(result, entry)
	}
	if !found {
		result = append(result, "PATH="+dir)
	}
	return result
}

func isPathName(name string) bool {
	if runtime.GOOS == "windows" {
		return strings.EqualFold(name, "PATH")
	}
	return name == "PATH"
}
