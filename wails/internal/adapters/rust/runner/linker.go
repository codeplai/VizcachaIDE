package runner

import (
	"os"
	"path/filepath"
	"runtime"
	"strings"

	"github.com/codeplai/VizcachaIDE/wails/internal/adapters/rust"
)

// bundledClang is where full-cpp ships llvm-mingw (docs/PLAN_CPP.md section 7).
const bundledClang = "toolchain/cpp/bin"

// linkerFor is the linker to pass to rustc ("-C linker=") and cargo, or "" to let them use their
// own (docs/PLAN_RUST.md section 3.2 point 2: llvm-mingw's clang is only a fallback).
func (r *Runner) linkerFor(toolchain rust.Toolchain) string {
	return chooseLinker(toolchain, runtime.GOOS, isFile, r.clang)
}

// chooseLinker pins the linker of a Windows GNU toolchain: its own MinGW gcc when it has one (an
// llvm-mingw on PATH would otherwise be picked up and fail on -lgcc_eh), llvm-mingw's clang when it
// does not, "" elsewhere. exists and clang are parameters so the decision is testable.
func chooseLinker(toolchain rust.Toolchain, goos string, exists func(string) bool, clang func() string) string {
	own := toolchain.SelfContainedGCC()
	if goos != "windows" || own == "" {
		return ""
	}
	if exists(own) {
		return own
	}
	return clang()
}

// clang finds llvm-mingw's clang: bundled with the IDE, then on PATH. "" when there is none.
func (r *Runner) clang() string {
	name := executableName("clang")
	candidates := []string{filepath.Join(r.options.AppDir, filepath.FromSlash(bundledClang), name)}
	for _, entry := range r.base {
		if key, value, ok := strings.Cut(entry, "="); ok && strings.EqualFold(key, "PATH") {
			for _, dir := range filepath.SplitList(value) {
				candidates = append(candidates, filepath.Join(dir, name))
			}
		}
	}
	for _, candidate := range candidates {
		if isFile(candidate) {
			return candidate
		}
	}
	return ""
}

// cargoLinkerVariable is the environment variable that sets cargo's linker for a host triple:
// x86_64-pc-windows-gnu gives CARGO_TARGET_X86_64_PC_WINDOWS_GNU_LINKER.
func cargoLinkerVariable(host string) string {
	triple := strings.ToUpper(strings.NewReplacer("-", "_", ".", "_").Replace(host))
	return "CARGO_TARGET_" + triple + "_LINKER"
}

func isFile(path string) bool {
	info, err := os.Stat(path)
	return err == nil && !info.IsDir()
}
