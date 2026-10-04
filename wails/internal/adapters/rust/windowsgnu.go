package rust

import (
	"context"
	"path/filepath"
	"strings"
)

// llvmDlltool is LLVM's dlltool. Crates that link Windows DLLs by name (getrandom, so rand and
// most crates that need random numbers) make rustc build import libraries with a dlltool; the
// one of rustup's GNU toolchain calls an assembler (as.exe) the toolchain does not bring, so the
// build fails with "dlltool could not create import library". LLVM's needs none: llvm-mingw has it
// (found in the multi-file experiment after M3).
const llvmDlltool = "llvm-dlltool.exe"

// Dlltool is the llvm-dlltool to give rustc on a Windows GNU host ("-C dlltool="): the one of the
// bundled llvm-mingw, the one next to the chosen lldb-dap, or one on PATH; "" when there is none
// (then crates such as rand fail to build, and the Assistant explains how to get llvm-mingw).
func (l *Locator) Dlltool(toolchain Toolchain) string {
	if !strings.HasSuffix(toolchain.Host, "-windows-gnu") {
		return ""
	}
	folders := []string{filepath.Join(applicationDirectory(l.options.AppDir), filepath.FromSlash(bundledLldb))}
	if lldb := l.configured(ToolLldbDap); lldb != "" {
		folders = append(folders, filepath.Dir(lldb))
	}
	folders = append(folders, filepath.SplitList(variable(l.options.BaseEnvironment, "PATH"))...)
	for _, folder := range folders {
		if folder == "" {
			continue
		}
		if candidate := filepath.Join(folder, llvmDlltool); isFile(candidate) {
			return candidate
		}
	}
	return ""
}

// gnuVariables are the variables cargo needs on a Windows GNU host, the same for every cargo the
// IDE runs (build, clippy, rust-analyzer's check), so they share target/ without rebuilding: the
// toolchain's own MinGW linker (an llvm-mingw on PATH would shadow it) and LLVM's dlltool.
func (l *Locator) gnuVariables() []string {
	toolchain, err := l.Toolchain(context.Background())
	if err != nil || !strings.HasSuffix(toolchain.Host, "-windows-gnu") {
		return nil
	}
	prefix := "CARGO_TARGET_" + strings.ToUpper(strings.NewReplacer("-", "_", ".", "_").Replace(toolchain.Host))
	var variables []string
	if gcc := toolchain.SelfContainedGCC(); isFile(gcc) {
		variables = append(variables, prefix+"_LINKER="+gcc)
	}
	// CARGO_ENCODED_RUSTFLAGS takes its flags whole (RUSTFLAGS would split a path with spaces,
	// such as the bundled llvm-dlltool under C:/Program Files/VizcachaIDE).
	if dlltool := l.Dlltool(toolchain); dlltool != "" {
		variables = append(variables, "CARGO_ENCODED_RUSTFLAGS=-Cdlltool="+dlltool)
	}
	return variables
}
