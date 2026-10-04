package main

import (
	"github.com/codeplai/VizcachaIDE/wails/internal/app"
	"github.com/codeplai/VizcachaIDE/wails/internal/domain"
)

// Provisional profile of C++ (docs/PLAN_CPP.md section 4.1, without Tools: no tool can be detected
// without an adapter). It exists so the menus, the file dialogs and the language registry know the
// language; running, debugging and building answer ErrUnsupported. M2 deletes it when the C++
// adapter arrives (Python got its adapter in M1, see support_python.go).

var provisionalCppProfile = domain.LanguageProfile{
	ID: domain.CodeLanguageCpp, NameKey: "codeLanguage.cpp",
	Extensions: []string{".cpp", ".cc", ".cxx", ".c++", ".h", ".hpp", ".hh"},
	Indent:     domain.IndentStyle{UseTabs: false, Size: 4},
	Capabilities: domain.Capabilities{
		Build: true, Console: false, Format: true, Check: true, DebugInput: true,
		PackageActions: nil,
		ThreadsLabel:   "debug.threads",
	},
	Tools: []domain.ToolSpec{},
}

// newUnavailableSupports returns the languages that have a profile but no adapter yet.
func newUnavailableSupports() []app.LanguageSupport {
	return []app.LanguageSupport{
		app.UnavailableSupport(provisionalCppProfile),
	}
}
