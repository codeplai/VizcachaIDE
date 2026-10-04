package cpp

import (
	"path/filepath"
	"strings"
)

// Family is the compiler family: both are supported by the same adapter.
type Family string

// Family values.
const (
	GCC   Family = "gcc"
	Clang Family = "clang"
)

// commonFlags: C++17, debug information without optimisation (the debugger shows every
// variable), the warnings beginners need and plain diagnostics for the parser.
var commonFlags = []string{"-std=c++17", "-g", "-O0", "-Wall", "-Wextra", "-fdiagnostics-color=never"}

// FamilyOf decides the family from the executable name and confirms it with the output of
// "<compiler> --version": "c++" and "g++" are clang on macOS and in llvm-mingw.
func FamilyOf(path, versionOutput string) Family {
	if strings.Contains(strings.ToLower(versionOutput), "clang") {
		return Clang
	}
	name := strings.ToLower(strings.TrimSuffix(filepath.Base(path), filepath.Ext(path)))
	if strings.Contains(name, "clang") {
		return Clang
	}
	return GCC
}

// CompileFlags are the flags of every compilation for a family on a system. On Windows the
// executable is linked statically so it runs outside the IDE and under the debugger without the
// compiler's DLLs on PATH (libstdc++ for GCC, libc++ and libunwind for llvm-mingw: without it
// the program exits with 0xc0000135, see PLAN_CPP.md section 3.1).
func CompileFlags(_ Family, goos string) []string {
	flags := append([]string{}, commonFlags...)
	if goos == "windows" {
		flags = append(flags, "-static")
	}
	return flags
}
