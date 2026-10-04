package main

import (
	"github.com/codeplai/VizcachaIDE/wails/internal/app"
	"github.com/codeplai/VizcachaIDE/wails/internal/domain"
)

// Provisional profiles of Python and C++ (docs/PLAN_PYTHON.md and docs/PLAN_CPP.md section 4.1,
// without Tools: no tool can be detected without an adapter). They exist so the menus, the file
// dialogs and the language registry know the languages in 2.1; running, debugging and building
// answer ErrUnsupported. M1 deletes the Python one and M2 the C++ one when their adapters arrive.

var provisionalPythonProfile = domain.LanguageProfile{
	ID: domain.CodeLanguagePython, NameKey: "codeLanguage.python", Extensions: []string{".py", ".pyw"},
	Indent: domain.IndentStyle{UseTabs: false, Size: 4},
	Capabilities: domain.Capabilities{
		Build: false, Console: true, Format: true, Check: true, DebugInput: true,
		PackageActions: []domain.PackageAction{domain.PackageAdd, domain.PackageRemove, domain.PackageList},
		ThreadsLabel:   "debug.threads",
	},
	Tools: []domain.ToolSpec{},
}

var provisionalCppProfile = domain.LanguageProfile{
	ID: domain.CodeLanguageCpp, NameKey: "codeLanguage.cpp",
	Extensions: []string{".cpp", ".cc", ".cxx", ".c++", ".h", ".hpp", ".hh"},
	Indent:     domain.IndentStyle{UseTabs: false, Size: 4},
	Capabilities: domain.Capabilities{
		Build: true, Console: false, Format: true, Check: false, DebugInput: false,
		PackageActions: nil,
		ThreadsLabel:   "debug.threads",
	},
	Tools: []domain.ToolSpec{},
}

// newUnavailableSupports returns the languages that have a profile but no adapter yet.
func newUnavailableSupports() []app.LanguageSupport {
	return []app.LanguageSupport{
		app.UnavailableSupport(provisionalPythonProfile),
		app.UnavailableSupport(provisionalCppProfile),
	}
}
