// Package golang is the Go language support: its profile and the module helpers that every
// part of the adapter shares. The parts live in subfolders: runner, packages, errors, console,
// delve and gopls.
package golang

import "github.com/codeplai/VizcachaIDE/wails/internal/domain"

// Profile is everything the frontend needs to know about Go.
var Profile = domain.LanguageProfile{
	ID: domain.CodeLanguageGo, NameKey: "codeLanguage.go", Extensions: []string{".go"},
	Indent: domain.IndentStyle{UseTabs: true, Size: 4},
	Capabilities: domain.Capabilities{
		Build: true, Console: true, Format: true, Check: true,
		PackageActions: []domain.PackageAction{domain.PackageInit, domain.PackageAdd, domain.PackageTidy},
		ThreadsLabel:   "debug.goroutines",
	},
	Tools: []domain.ToolSpec{
		{
			ID: "go", Role: domain.RoleRuntime, LabelKey: "settings.toolGo",
			MissingKey: "errors.goNotFound", InstallURL: "https://go.dev/dl/",
		},
		{
			ID: "dlv", Role: domain.RoleDebugAdapter, LabelKey: "settings.toolDelve",
			MissingKey: "errors.delveNotFound", InstallCommand: "go install github.com/go-delve/delve/cmd/dlv@latest",
		},
		{
			ID: "gopls", Role: domain.RoleLanguageServer, LabelKey: "settings.toolGopls",
			MissingKey: "errors.goplsNotFound", InstallCommand: "go install golang.org/x/tools/gopls@latest",
		},
	},
}
