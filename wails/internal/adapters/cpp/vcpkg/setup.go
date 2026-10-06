package vcpkg

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// Setup prepares the build to use vcpkg. It has the three methods of cmake.Dependencies, which
// adapters/cpp/cmake declares (the compile-time check lives where both packages are wired).
type Setup struct{ locator *Locator }

// NewSetup creates the Setup over a locator.
func NewSetup(locator *Locator) *Setup { return &Setup{locator: locator} }

// ConfigureArgs returns the "-D" arguments that make CMake use vcpkg in manifest mode. Without
// vcpkg (no valid root) it returns nil and plain CMake still works. vcpkg writes its build trees
// into its own root unless told otherwise, and that root is read-only in Program Files, so they
// go to the cache.
func (s *Setup) ConfigureArgs(string) []string {
	root := s.locator.Root()
	if root == "" {
		return nil
	}
	cache := s.locator.CacheRoot()
	triplet := s.locator.Triplet()
	options := "--x-buildtrees-root=" + cmakePath(filepath.Join(cache, "buildtrees")) +
		";--x-packages-root=" + cmakePath(filepath.Join(cache, "packages"))
	return []string{
		"-DCMAKE_TOOLCHAIN_FILE=" + cmakePath(filepath.Join(root, "scripts", "buildsystems", "vcpkg.cmake")),
		"-DVCPKG_TARGET_TRIPLET=" + triplet,
		"-DVCPKG_HOST_TRIPLET=" + triplet,
		"-DVCPKG_INSTALL_OPTIONS=" + options,
	}
}

// Environment returns base plus the variables vcpkg reads. On Windows the folder of the compiler
// goes first on PATH. Without a valid vcpkg root base comes back unchanged.
func (s *Setup) Environment(base []string) []string {
	root := s.locator.Root()
	if root == "" {
		return base
	}
	cache := s.locator.CacheRoot()
	archives := filepath.Join(cache, "archives")
	_ = os.MkdirAll(archives, 0o755) // vcpkg refuses a binary cache folder that does not exist
	environment := withVariables(base, map[string]string{
		"VCPKG_ROOT":                 root,
		"VCPKG_DOWNLOADS":            filepath.Join(cache, "downloads"),
		"VCPKG_DEFAULT_BINARY_CACHE": archives,
		"VCPKG_DISABLE_METRICS":      "1",
	})
	folders := s.locator.options.pathFolders()
	if bin := s.locator.compilerBin(); bin != "" && s.locator.options.GOOS == "windows" {
		folders = append([]string{bin}, folders...)
	}
	if len(folders) == 0 {
		return environment
	}
	folders = append(folders, environmentValue(environment, "PATH"))
	return withVariables(environment, map[string]string{"PATH": strings.Join(folders, string(os.PathListSeparator))})
}

// Prepare puts the seed in the downloads folder of vcpkg (seed.go); it is a no-op without seed.
func (s *Setup) Prepare(ctx context.Context) error {
	seed := s.locator.Seed()
	if seed == "" || s.locator.Root() == "" {
		return nil
	}
	downloads := filepath.Join(s.locator.CacheRoot(), "downloads")
	if err := copySeed(ctx, seed, downloads); err != nil {
		return fmt.Errorf("prepare the vcpkg downloads: %w", err)
	}
	return nil
}

// withVariables returns environment with the given variables set, replacing any previous value
// whatever its case (Windows spells PATH "Path").
func withVariables(environment []string, values map[string]string) []string {
	result := make([]string, 0, len(environment)+len(values))
	for _, entry := range environment {
		name, _, _ := strings.Cut(entry, "=")
		if !hasFold(values, name) {
			result = append(result, entry)
		}
	}
	for name, value := range values {
		result = append(result, name+"="+value)
	}
	return result
}

func hasFold(values map[string]string, name string) bool {
	for key := range values {
		if strings.EqualFold(key, name) {
			return true
		}
	}
	return false
}

// cmakePath writes a path with forward slashes: CMake reads backslashes as escapes.
func cmakePath(path string) string { return strings.ReplaceAll(path, `\`, "/") }
