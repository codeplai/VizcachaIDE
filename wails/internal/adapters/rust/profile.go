// Package rust is the Rust language support: its profile, how to find rustc, cargo and the
// other tools of a rustup toolchain, what that toolchain is (version, host, sysroot) and the Cargo
// project around a file. The parts live in subfolders: runner, lldbdap, analyzer, clippy,
// rustfmt, packages and errors (docs/PLAN_RUST.md section 4).
package rust

import "github.com/codeplai/VizcachaIDE/wails/internal/domain"

// Tool ids of the Rust profile.
const (
	ToolRustc        = "rustc"
	ToolCargo        = "cargo"
	ToolRustAnalyzer = "rust-analyzer"
	ToolLldbDap      = "lldb-dap"
	ToolClippy       = "clippy"
	ToolRustfmt      = "rustfmt"
)

// rustupURL is where Rust is installed from (free, MIT or Apache-2.0).
const rustupURL = "https://rustup.rs"

// Profile is everything the frontend needs to know about Rust. clippy and rustfmt are components
// of the toolchain (ProvidedBy cargo): they have a status row but no "Choose" button.
var Profile = domain.LanguageProfile{
	ID: domain.CodeLanguageRust, NameKey: "codeLanguage.rust", Extensions: []string{".rs"},
	Indent: domain.IndentStyle{UseTabs: false, Size: 4},
	Capabilities: domain.Capabilities{
		Build: true, Console: false, Format: true, Check: true, DebugInput: true,
		// Cargo has no "tidy": dependencies are added and removed one by one.
		PackageActions: []domain.PackageAction{domain.PackageInit, domain.PackageAdd, domain.PackageRemove, domain.PackageList},
		ThreadsLabel:   "debug.threads",
	},
	Tools: []domain.ToolSpec{
		{
			ID: ToolRustc, Role: domain.RoleCompiler, LabelKey: "settings.toolRustc",
			MissingKey: "errors.rustNotFound", InstallURL: rustupURL,
		},
		{
			ID: ToolCargo, Role: domain.RoleRuntime, LabelKey: "settings.toolCargo",
			MissingKey: "errors.cargoNotFound", InstallURL: rustupURL,
		},
		{
			ID: ToolRustAnalyzer, Role: domain.RoleLanguageServer, LabelKey: "settings.toolRustAnalyzer",
			MissingKey: "errors.rustAnalyzerNotFound", InstallCommand: "rustup component add rust-analyzer",
		},
		{
			ID: ToolLldbDap, Role: domain.RoleDebugAdapter, LabelKey: "settings.toolLldbDap",
			MissingKey: "errors.lldbDapNotFound", InstallURL: "https://github.com/mstorsjo/llvm-mingw/releases",
		},
		{
			ID: ToolClippy, Role: domain.RoleCompiler, ProvidedBy: ToolCargo, LabelKey: "settings.toolClippy",
			MissingKey: "errors.clippyNotFound", InstallCommand: "rustup component add clippy",
		},
		{
			ID: ToolRustfmt, Role: domain.RoleFormatter, ProvidedBy: ToolCargo, LabelKey: "settings.toolRustfmt",
			MissingKey: "errors.rustfmtNotFound", InstallCommand: "rustup component add rustfmt",
		},
	},
}
