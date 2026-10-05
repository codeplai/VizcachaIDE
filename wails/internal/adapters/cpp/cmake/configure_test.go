package cmake

import (
	"context"
	"errors"
	"os"
	"path/filepath"
	"runtime"
	"slices"
	"strings"
	"testing"
	"time"

	"github.com/codeplai/VizcachaIDE/wails/internal/adapters/cpp/cpptest"
)

// fakeDependencies stands for vcpkg.
type fakeDependencies struct {
	prepared int
	fail     error
}

func (f *fakeDependencies) ConfigureArgs(root string) []string {
	return []string{"-DCMAKE_TOOLCHAIN_FILE=/vcpkg/vcpkg.cmake", "-DFROM=" + filepath.Base(root)}
}
func (f *fakeDependencies) Environment(base []string) []string {
	return append(base, "VCPKG_ROOT=/vcpkg")
}
func (f *fakeDependencies) Prepare(context.Context) error {
	f.prepared++
	return f.fail
}

func newProject(t *testing.T) string {
	t.Helper()
	root := filepath.Join(t.TempDir(), "Ñandú")
	for name, text := range Files("Ñandú", "en") {
		write(t, filepath.Join(root, name), text)
	}
	write(t, filepath.Join(root, "main.cpp"), "int main() { return 0; }\n")
	return root
}

func TestPlanIsTheR0Command(t *testing.T) {
	root := newProject(t)
	deps := &fakeDependencies{}
	builder := New(Options{CMake: "cmake", Ninja: "ninja", Compiler: filepath.Join("llvm", "bin", "clang++"), Dependencies: deps,
		Environment: []string{"Path=C:\\Windows", "PATH=/usr/bin", "OTHER=1"}, SupportDir: t.TempDir()})
	plan, err := builder.Plan(context.Background(), root, false)
	if err != nil || plan.Configure == nil || deps.prepared != 1 {
		t.Fatalf("plan = %+v, %v, prepared %d", plan, err, deps.prepared)
	}
	args := plan.Configure.Args
	for _, want := range []string{"-S", root, "-B", filepath.Join(root, "build"), "-G", "Ninja", "-DCMAKE_BUILD_TYPE=Debug",
		"-DCMAKE_CXX_COMPILER=" + filepath.Join("llvm", "bin", "clang++"), "-DCMAKE_MAKE_PROGRAM=ninja", "-DCMAKE_TOOLCHAIN_FILE=/vcpkg/vcpkg.cmake", "-DFROM=Ñandú"} {
		if !slices.Contains(args, want) {
			t.Errorf("configure args %v lack %q", args, want)
		}
	}
	if runtime.GOOS == "windows" && !slices.ContainsFunc(args, func(a string) bool {
		return strings.HasPrefix(a, "-DCMAKE_PROJECT_TOP_LEVEL_INCLUDES=") && strings.HasSuffix(a, "vizcacha.cmake")
	}) {
		t.Errorf("configure args %v lack the vizcacha.cmake include", args)
	}
	if got := plan.Build.Args; !slices.Equal(got, []string{"--build", filepath.Join(root, "build")}) {
		t.Errorf("build args = %v", got)
	}
	if !slices.Contains(plan.Build.Env, "VCPKG_ROOT=/vcpkg") || !strings.HasPrefix(pathOf(plan.Build.Env), filepath.Join("llvm", "bin")) {
		t.Errorf("env = %v, want the compiler's folder first on PATH and the libraries' variables", plan.Build.Env)
	}
	if !isFile(filepath.Join(apiFolder(filepath.Join(root, "build")), "query", "codemodel-v2")) {
		t.Error("the File API query must exist before configure")
	}
}

func pathOf(env []string) string {
	for _, entry := range env {
		if name, value, ok := strings.Cut(entry, "="); ok && strings.EqualFold(name, "PATH") && strings.Contains(value, "llvm") {
			return value
		}
	}
	return ""
}

func TestPlanWithoutDependenciesAndReleaseFolder(t *testing.T) {
	root := newProject(t)
	plan, err := New(Options{CMake: "cmake", Ninja: "ninja", Compiler: "clang++"}).Plan(context.Background(), root, true)
	if err != nil || plan.BuildDir != filepath.Join(root, "build", "release") || !slices.Contains(plan.Configure.Args, "-DCMAKE_BUILD_TYPE=Release") {
		t.Fatalf("plan = %+v, %v", plan, err)
	}
	failing := New(Options{Dependencies: &fakeDependencies{fail: errors.New("no seed")}})
	if _, err := failing.Plan(context.Background(), root, false); err == nil {
		t.Fatal("a Prepare failure must stop the plan")
	}
}

func TestStaleBuildFolder(t *testing.T) {
	root := newProject(t)
	build := BuildDir(root, false)
	if !needsConfigure(root, build) {
		t.Fatal("a project that was never configured needs it")
	}
	cache := filepath.Join(build, "CMakeCache.txt")
	write(t, cache, "CMAKE_HOME_DIRECTORY:INTERNAL="+filepath.ToSlash(root)+"\n")
	if !needsConfigure(root, build) {
		t.Fatal("without the File API reply CMake has not answered")
	}
	write(t, filepath.Join(apiFolder(build), "reply", "index-1.json"), "{}")
	old := time.Now().Add(-time.Hour)
	for _, name := range inputs {
		_ = os.Chtimes(filepath.Join(root, name), old, old)
	}
	if needsConfigure(root, build) {
		t.Fatal("a cache newer than the project files is up to date")
	}
	later := time.Now().Add(time.Hour)
	if err := os.Chtimes(filepath.Join(root, "vcpkg.json"), later, later); err != nil || !needsConfigure(root, build) {
		t.Fatalf("a newer vcpkg.json makes it stale (%v)", err)
	}
}

func TestMovedProjectDropsTheBuildFolder(t *testing.T) {
	root := newProject(t)
	build := BuildDir(root, false)
	write(t, filepath.Join(build, "CMakeCache.txt"), "CMAKE_HOME_DIRECTORY:INTERNAL=/somewhere/else\n")
	write(t, filepath.Join(apiFolder(build), "reply", "index-1.json"), "{}")
	if !needsConfigure(root, build) {
		t.Fatal("a cache of another folder is stale")
	}
	if isFile(filepath.Join(build, "CMakeCache.txt")) {
		t.Fatal("the stale build folder must be removed: CMake refuses to reuse it")
	}
}

func realBuilder(t *testing.T) *Builder {
	t.Helper()
	llvm, cmakeBin := cpptest.LLVMBin(t), cpptest.CMakeBin(t)
	return New(Options{
		CMake: filepath.Join(cmakeBin, cpptest.Exe("cmake")), Ninja: filepath.Join(cmakeBin, cpptest.Exe("ninja")),
		Compiler: filepath.Join(llvm, cpptest.Exe("clang++")), Environment: os.Environ(), SupportDir: t.TempDir(),
	})
}

func TestConfigureAndBuildWithRealCMake(t *testing.T) {
	builder := realBuilder(t)
	root := newProject(t)
	ctx := context.Background()

	output, err := builder.Configure(ctx, root)
	if err != nil {
		t.Fatalf("Configure = %v\n%s", err, output)
	}
	if !isFile(filepath.Join(root, "build", "compile_commands.json")) {
		t.Error("clangd needs build/compile_commands.json")
	}
	if again, err := builder.Configure(ctx, root); err != nil || again != "" {
		t.Fatalf("a configured project is not configured again: %q, %v", again, err)
	}
	if output, err := builder.Build(ctx, root, false); err != nil {
		t.Fatalf("Build = %v\n%s", err, output)
	}
	executables, err := Executables(BuildDir(root, false))
	if err != nil || len(executables) != 1 || executables[0].Name != "nandu" || !isFile(executables[0].Path) {
		t.Fatalf("executables = %+v, %v", executables, err)
	}
}

func TestConfigureFailureWrapsErrConfigureFailed(t *testing.T) {
	builder := realBuilder(t)
	root := newProject(t)
	write(t, filepath.Join(root, ListsFile), "cmake_minimum_required(VERSION 3.25)\nproject(x LANGUAGES CXX)\nfind_package(NoExiste REQUIRED)\n")
	output, err := builder.Configure(context.Background(), root)
	if !errors.Is(err, ErrConfigureFailed) || !strings.Contains(output, "CMake Error at CMakeLists.txt:3 (find_package)") {
		t.Fatalf("Configure = %v\n%s", err, output)
	}
}
