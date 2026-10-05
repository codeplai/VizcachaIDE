// Package cpp is the C++ language support: its profile, the compiler family (GCC or Clang) and
// how to find the compiler and the LLVM tools. The parts live in subfolders: runner, lldbdap,
// clangd, clangformat and errors (docs/PLAN_CPP.md section 4).
package cpp

import "github.com/codeplai/VizcachaIDE/wails/internal/domain"

// Tool ids of the C++ profile.
const (
	ToolCompiler    = "cxx"
	ToolLldbDap     = "lldb-dap"
	ToolClangd      = "clangd"
	ToolClangFormat = "clang-format"
	ToolCMake       = "cmake"
	ToolNinja       = "ninja"
	ToolVcpkg       = "vcpkg" // located by the vcpkg package (its root folder, not a bin folder)
)

// Profile is everything the frontend needs to know about C++. The install hints are the Windows
// ones; support_cpp.go adapts them on macOS and Linux.
var Profile = domain.LanguageProfile{
	ID: domain.CodeLanguageCpp, NameKey: "codeLanguage.cpp",
	Extensions: []string{".cpp", ".cc", ".cxx", ".c++", ".h", ".hpp", ".hh"},
	Indent:     domain.IndentStyle{UseTabs: false, Size: 4},
	Capabilities: domain.Capabilities{
		Build: true, Console: false, Format: true, Check: true, DebugInput: true,
		PackageActions: []domain.PackageAction{domain.PackageAdd, domain.PackageRemove, domain.PackageList},
		ThreadsLabel:   "debug.threads",
	},
	Tools: []domain.ToolSpec{
		{
			ID: ToolCompiler, Role: domain.RoleCompiler, LabelKey: "settings.toolCxx",
			MissingKey: "errors.cxxNotFound", InstallURL: "https://winlibs.com/",
		},
		{
			ID: ToolLldbDap, Role: domain.RoleDebugAdapter, LabelKey: "settings.toolLldbDap",
			MissingKey: "errors.lldbDapNotFound", InstallURL: "https://github.com/mstorsjo/llvm-mingw/releases",
		},
		{
			ID: ToolClangd, Role: domain.RoleLanguageServer, LabelKey: "settings.toolClangd",
			MissingKey: "errors.clangdNotFound", InstallURL: "https://clangd.llvm.org/installation",
		},
		{
			ID: ToolClangFormat, Role: domain.RoleFormatter, LabelKey: "settings.toolClangFormat",
			MissingKey: "errors.clangFormatNotFound", InstallURL: "https://github.com/mstorsjo/llvm-mingw/releases",
		},
		{
			ID: ToolCMake, Role: domain.RoleBuildTool, LabelKey: "settings.toolCMake",
			MissingKey: "errors.cmakeNotFound", InstallURL: "https://cmake.org/download/",
		},
		{
			ID: ToolNinja, Role: domain.RoleBuildTool, LabelKey: "settings.toolNinja",
			MissingKey: "errors.ninjaNotFound", InstallURL: "https://github.com/ninja-build/ninja/releases",
		},
		{
			ID: ToolVcpkg, Role: domain.RoleBuildTool, LabelKey: "settings.toolVcpkg",
			MissingKey: "errors.vcpkgNotFound", InstallURL: "https://github.com/microsoft/vcpkg",
		},
	},
}
