package vcpkg

import (
	"bytes"
	"context"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
	"time"

	"github.com/codeplai/VizcachaIDE/wails/internal/protocol/process"
)

// cmakeConfigurer is the minimum of cmake.Builder that the integration needs: the command line of
// docs/PLAN_CPP_CMAKE.md section 8 with the arguments and the environment of Setup.
type cmakeConfigurer struct {
	setup *Setup
	bin   string // the folder of cmake and ninja
}

func (c cmakeConfigurer) Configure(ctx context.Context, root string) (string, error) {
	if err := c.setup.Prepare(ctx); err != nil {
		return "", err
	}
	args := []string{
		"-S", root, "-B", filepath.Join(root, "build"), "-G", "Ninja", "-DCMAKE_BUILD_TYPE=Debug",
		"-DCMAKE_CXX_COMPILER=clang++", "-DCMAKE_MAKE_PROGRAM=" + filepath.Join(c.bin, "ninja.exe"),
	}
	args = append(args, c.setup.ConfigureArgs(root)...)
	command := exec.CommandContext(ctx, filepath.Join(c.bin, "cmake.exe"), args...)
	command.Dir = root
	command.Env = c.setup.Environment(os.Environ())
	process.HideConsole(command)
	var output bytes.Buffer
	command.Stdout, command.Stderr = &output, &output
	err := command.Run()
	return output.String(), err
}

func TestAddFmtBuildsRunsAndRemoves(t *testing.T) {
	rootDir := os.Getenv("VIZCACHA_TEST_VCPKG_ROOT")
	seed := os.Getenv("VIZCACHA_TEST_VCPKG_SEED")
	cmakeBin := os.Getenv("VIZCACHA_TEST_CMAKE_BIN")
	llvmBin := os.Getenv("VIZCACHA_TEST_LLVM_BIN")
	if runtime.GOOS != "windows" || rootDir == "" || cmakeBin == "" || llvmBin == "" {
		t.Skip("set VIZCACHA_TEST_VCPKG_ROOT, _VCPKG_SEED, _CMAKE_BIN and _LLVM_BIN (Windows) to run")
	}
	cache := os.Getenv("VIZCACHA_TEST_VCPKG_CACHE") // a shared folder keeps reruns fast
	if cache == "" {
		cache = t.TempDir()
	}
	locator := NewLocator(LocatorOptions{
		Settings: settingsWith(rootDir), AppDir: t.TempDir(), CacheDir: cache, SeedDir: seed,
		CompilerBin: func() string { return llvmBin }, ToolBins: func() []string { return []string{cmakeBin} },
	})
	setup := NewSetup(locator)
	events := newEvents()
	manager := NewManager(locator, ManagerOptions{Configurer: cmakeConfigurer{setup: setup, bin: cmakeBin}, Events: events})

	project := filepath.Join(t.TempDir(), "Ñandú")
	mustWrite(t, filepath.Join(project, "CMakeLists.txt"), "cmake_minimum_required(VERSION 3.25)\nproject(nandu LANGUAGES CXX)\n\n"+
		"set(CMAKE_CXX_STANDARD 17)\nset(CMAKE_CXX_STANDARD_REQUIRED ON)\nset(CMAKE_CXX_EXTENSIONS OFF)\nset(CMAKE_EXPORT_COMPILE_COMMANDS ON)\n\n"+
		"file(GLOB SOURCES CONFIGURE_DEPENDS *.cpp *.cc *.cxx)\nadd_executable(nandu ${SOURCES})\ntarget_compile_options(nandu PRIVATE -Wall -Wextra)\n\n"+
		LibrariesBegin+"\n"+LibrariesEnd+"\n")
	mustWrite(t, filepath.Join(project, "vcpkg.json"), "{ \"name\": \"nandu\", \"version\": \"0.1.0\", \"dependencies\": [] }\n")
	mustWrite(t, filepath.Join(project, "main.cpp"), "#include <fmt/core.h>\nint main() { fmt::print(\"hola {}\\n\", 42); }\n")

	started := time.Now()
	if err := manager.Add(context.Background(), project, "fmt"); err != nil {
		t.Fatal(err)
	}
	if code := waitFor(t, events, 15*time.Minute); code != 0 {
		t.Fatalf("Add exit code %d:\n%s", code, events.text())
	}
	t.Logf("Add(fmt) took %s", time.Since(started).Round(time.Second))

	cmake := readText(t, filepath.Join(project, "CMakeLists.txt"))
	for _, want := range []string{"find_package(fmt CONFIG REQUIRED)", "target_link_libraries(nandu PRIVATE fmt::fmt)"} {
		if !strings.Contains(cmake, want) {
			t.Errorf("CMakeLists lacks %q:\n%s", want, cmake)
		}
	}
	build := exec.Command(filepath.Join(cmakeBin, "cmake.exe"), "--build", filepath.Join(project, "build"))
	build.Env = setup.Environment(os.Environ())
	process.HideConsole(build)
	if output, err := build.CombinedOutput(); err != nil {
		t.Fatalf("cmake --build: %v\n%s", err, output)
	}
	program := exec.Command(filepath.Join(project, "build", "nandu.exe"))
	program.Env = setup.Environment(os.Environ()) // the llvm-mingw runtime DLLs are on PATH
	process.HideConsole(program)
	if output, err := program.CombinedOutput(); err != nil || !strings.Contains(string(output), "hola 42") {
		t.Errorf("program: %v %q", err, output)
	}

	events.reset()
	if err := manager.List(context.Background(), project); err != nil {
		t.Fatal(err)
	}
	waitFor(t, events, time.Minute)
	if got := events.text(); !strings.HasPrefix(got, "fmt ") || strings.TrimSpace(strings.TrimPrefix(got, "fmt ")) == "" {
		t.Errorf("List = %q, want fmt with a version", got)
	}

	events.reset()
	if err := manager.Remove(context.Background(), project, "fmt"); err != nil {
		t.Fatal(err)
	}
	if code := waitFor(t, events, 5*time.Minute); code != 0 {
		t.Fatalf("Remove exit code %d:\n%s", code, events.text())
	}
	if names, _ := Dependencies(project); len(names) != 0 {
		t.Errorf("dependencies after Remove = %v", names)
	}
	if got := readText(t, filepath.Join(project, "CMakeLists.txt")); strings.Contains(got, "fmt") {
		t.Errorf("block still mentions fmt:\n%s", got)
	}
}

func TestSearchRealPorts(t *testing.T) {
	rootDir := os.Getenv("VIZCACHA_TEST_VCPKG_ROOT")
	if rootDir == "" {
		t.Skip("set VIZCACHA_TEST_VCPKG_ROOT to search the real ports")
	}
	locator := NewLocator(LocatorOptions{Settings: settingsWith(rootDir), AppDir: t.TempDir(), CacheDir: t.TempDir()})
	search := NewSearch(locator)
	started := time.Now()
	found, err := search.Search(context.Background(), "fmt")
	if err != nil || len(found) == 0 || found[0].Name != "fmt" || found[0].Version == "" {
		t.Fatalf("fmt: %+v, %v", found, err)
	}
	t.Logf("first search (ports scanned): %s", time.Since(started).Round(time.Millisecond))
	started = time.Now()
	if _, err := NewSearch(locator).Search(context.Background(), "json"); err != nil {
		t.Fatal(err)
	}
	t.Logf("a search with the index read from the cache file: %s", time.Since(started).Round(time.Millisecond))
	again, _ := search.Search(context.Background(), "sdl")
	t.Logf("sdl -> %v", names(again))
}

func waitFor(t *testing.T, events *recordedEvents, limit time.Duration) int {
	t.Helper()
	select {
	case code := <-events.finished:
		return code
	case <-time.After(limit):
		t.Fatalf("timed out after %s:\n%s", limit, events.text())
		return -1
	}
}
