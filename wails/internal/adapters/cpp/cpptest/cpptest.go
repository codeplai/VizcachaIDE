// Package cpptest gives the tests of the C++ adapter real compilers, or skips them. It is imported
// only from _test.go files.
package cpptest

import (
	"os"
	"path/filepath"
	"runtime"
	"testing"
)

// Variables that name the bin folders to test with (docs/PLAN_CPP.md section 0).
const (
	LLVMVariable = "VIZCACHA_TEST_LLVM_BIN"
	GCCVariable  = "VIZCACHA_TEST_GCC_BIN"
	// CMakeVariable names the bin folder with cmake and ninja (docs/PLAN_CPP_CMAKE.md section 8).
	CMakeVariable = "VIZCACHA_TEST_CMAKE_BIN"
)

// LLVMBin returns a bin folder with clang++, lldb-dap, clangd and clang-format, or skips.
func LLVMBin(t *testing.T) string {
	t.Helper()
	return binFolder(t, LLVMVariable, "clang++", func(dev string) []string {
		matches, _ := filepath.Glob(filepath.Join(dev, "llvm-mingw-*", "bin"))
		return matches
	})
}

// GCCBin returns a bin folder with a real GCC g++ (not clang under that name), or skips.
func GCCBin(t *testing.T) string {
	t.Helper()
	return binFolder(t, GCCVariable, "g++", func(dev string) []string {
		return []string{filepath.Join(dev, "mingw64", "bin")}
	})
}

// CMakeBin returns a bin folder with cmake and ninja, or skips.
func CMakeBin(t *testing.T) string {
	t.Helper()
	return binFolder(t, CMakeVariable, "cmake", func(dev string) []string {
		return []string{filepath.Join(dev, "cmake", "bin")}
	})
}

// Exe adds .exe on Windows.
func Exe(name string) string {
	if runtime.GOOS == "windows" {
		return name + ".exe"
	}
	return name
}

func binFolder(t *testing.T, variable, compiler string, candidates func(dev string) []string) string {
	t.Helper()
	if dir := os.Getenv(variable); dir != "" {
		return dir
	}
	if dev := developmentToolchains(); dev != "" {
		for _, dir := range candidates(dev) {
			if _, err := os.Stat(filepath.Join(dir, Exe(compiler))); err == nil {
				return dir
			}
		}
	}
	t.Skipf("no %s to test with: set %s or unpack it in wails/.toolchain-dev (docs/PLAN_CPP.md section 0)", compiler, variable)
	return ""
}

// developmentToolchains finds wails/.toolchain-dev walking up from the test's folder.
func developmentToolchains() string {
	dir, err := os.Getwd()
	if err != nil {
		return ""
	}
	for {
		candidate := filepath.Join(dir, ".toolchain-dev")
		if info, err := os.Stat(candidate); err == nil && info.IsDir() {
			return candidate
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			return ""
		}
		dir = parent
	}
}
