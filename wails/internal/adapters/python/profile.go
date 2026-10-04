// Package python is the Python language support: its profile, how to find the interpreter and
// the environment every Python process runs with. The parts live in subfolders: runner, debugpy,
// pylsp, ruff, repl, packages and errors (docs/PLAN_PYTHON.md section 4).
package python

import "github.com/codeplai/VizcachaIDE/wails/internal/domain"

// Module tool ids: debugpy, pylsp and ruff are modules of the interpreter ("python -m ...").
const (
	ToolPython  = "python"
	ToolDebugpy = "debugpy"
	ToolPylsp   = "pylsp"
	ToolRuff    = "ruff"
)

// Profile is everything the frontend needs to know about Python.
var Profile = domain.LanguageProfile{
	ID: domain.CodeLanguagePython, NameKey: "codeLanguage.python", Extensions: []string{".py", ".pyw"},
	Indent: domain.IndentStyle{UseTabs: false, Size: 4},
	Capabilities: domain.Capabilities{
		Build: false, Console: true, Format: true, Check: true, DebugInput: true,
		PackageActions: []domain.PackageAction{domain.PackageAdd, domain.PackageRemove, domain.PackageList},
		ThreadsLabel:   "debug.threads",
	},
	Tools: []domain.ToolSpec{
		{
			ID: ToolPython, Role: domain.RoleRuntime, LabelKey: "settings.toolPython",
			MissingKey: "errors.pythonNotFound", InstallURL: "https://www.python.org/downloads/",
		},
		{
			ID: ToolDebugpy, Role: domain.RoleDebugAdapter, ProvidedBy: ToolPython, LabelKey: "settings.modulePython",
			MissingKey: "errors.debugpyMissing", InstallCommand: "python -m pip install debugpy",
		},
		{
			ID: ToolPylsp, Role: domain.RoleLanguageServer, ProvidedBy: ToolPython, LabelKey: "settings.modulePython",
			MissingKey: "errors.pylspMissing", InstallCommand: "python -m pip install python-lsp-server",
		},
		{
			ID: ToolRuff, Role: domain.RoleFormatter, ProvidedBy: ToolPython, LabelKey: "settings.modulePython",
			MissingKey: "errors.ruffMissing", InstallCommand: "python -m pip install ruff",
		},
	},
}
