// Package rusttest gives the tests of the Rust adapter a real rustup toolchain, or skips them. It
// is imported only from _test.go files.
//
// The development toolchain lives in wails/.toolchain-dev (git-ignored), installed with
// rustup-init --no-modify-path, RUSTUP_HOME=.toolchain-dev/rustup and CARGO_HOME=.toolchain-dev/cargo
// (docs/PLAN_RUST.md section 0); VIZCACHA_TEST_CARGO_HOME and VIZCACHA_TEST_RUSTUP_HOME name
// another one.
package rusttest

import (
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
)

// Variables that name the toolchain to test with.
const (
	CargoHomeVariable  = "VIZCACHA_TEST_CARGO_HOME"
	RustupHomeVariable = "VIZCACHA_TEST_RUSTUP_HOME"
	LLVMVariable       = "VIZCACHA_TEST_LLVM_BIN"
)

// Environment returns os.Environ() with CARGO_HOME and RUSTUP_HOME of the test toolchain and its
// cargo/bin first in PATH, or skips the test.
func Environment(t *testing.T) []string {
	t.Helper()
	cargoHome, rustupHome := homes()
	if !isFile(filepath.Join(cargoHome, "bin", Exe("rustc"))) {
		t.Skipf("no Rust to test with: set %s and %s or install rustup in wails/.toolchain-dev (docs/PLAN_RUST.md section 0)", CargoHomeVariable, RustupHomeVariable)
	}
	env := []string{}
	for _, entry := range os.Environ() {
		name, _, _ := strings.Cut(entry, "=")
		switch strings.ToUpper(name) {
		case "CARGO_HOME", "RUSTUP_HOME", "PATH":
			continue
		}
		env = append(env, entry)
	}
	path := filepath.Join(cargoHome, "bin") + string(os.PathListSeparator) + os.Getenv("PATH")
	return append(env, "CARGO_HOME="+cargoHome, "RUSTUP_HOME="+rustupHome, "PATH="+path)
}

// Rustc returns the rustc proxy of the test toolchain, or skips the test.
func Rustc(t *testing.T) string {
	t.Helper()
	Environment(t)
	cargoHome, _ := homes()
	return filepath.Join(cargoHome, "bin", Exe("rustc"))
}

// LldbDap returns an lldb-dap (llvm-mingw's, as in docs/PLAN_CPP.md section 0), or skips the test.
func LldbDap(t *testing.T) string {
	t.Helper()
	dirs := []string{os.Getenv(LLVMVariable)}
	if dev := developmentToolchains(); dev != "" {
		matches, _ := filepath.Glob(filepath.Join(dev, "llvm-mingw-*", "bin"))
		dirs = append(dirs, matches...)
	}
	for _, dir := range dirs {
		if path := filepath.Join(dir, Exe("lldb-dap")); dir != "" && isFile(path) {
			return path
		}
	}
	t.Skipf("no lldb-dap to test with: set %s", LLVMVariable)
	return ""
}

// Exe adds .exe on Windows.
func Exe(name string) string {
	if runtime.GOOS == "windows" {
		return name + ".exe"
	}
	return name
}

func homes() (cargoHome, rustupHome string) {
	cargoHome, rustupHome = os.Getenv(CargoHomeVariable), os.Getenv(RustupHomeVariable)
	if cargoHome != "" {
		return cargoHome, rustupHome
	}
	if dev := developmentToolchains(); dev != "" {
		return filepath.Join(dev, "cargo"), filepath.Join(dev, "rustup")
	}
	return "", ""
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

func isFile(path string) bool {
	info, err := os.Stat(path)
	return err == nil && !info.IsDir()
}
