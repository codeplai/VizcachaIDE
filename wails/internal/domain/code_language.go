package domain

// CodeLanguage is a programming language the IDE supports. "Language" alone means the UI
// language (en/es, see settings.go); see PLAN_NUCLEO_MULTILENGUAJE.md section 3.
type CodeLanguage string

// CodeLanguage values.
const (
	CodeLanguageGo     CodeLanguage = "go"
	CodeLanguagePython CodeLanguage = "python"
	CodeLanguageCpp    CodeLanguage = "cpp"
	CodeLanguageRust   CodeLanguage = "rust"
)

// PackageAction is one verb of a language's package manager.
type PackageAction string

// PackageAction values.
const (
	PackageInit   PackageAction = "init"
	PackageAdd    PackageAction = "add"
	PackageRemove PackageAction = "remove"
	PackageTidy   PackageAction = "tidy"
	PackageList   PackageAction = "list"
)

// IndentStyle is how the editor indents files of a language.
type IndentStyle struct {
	UseTabs bool `json:"useTabs"`
	Size    int  `json:"size"`
}

// Capabilities says which buttons and panels apply to a language. The language registry checks
// that they match the ports of the language support, so they can never disagree.
type Capabilities struct {
	Build   bool `json:"build"`
	Console bool `json:"console"`
	Format  bool `json:"format"`
	Check   bool `json:"check"`
	// DebugInput says whether the program can read the keyboard while it is debugged (Python
	// through debugpy's runInTerminal in a pseudoterminal); Delve cannot, so Go says false.
	DebugInput     bool            `json:"debugInput"`
	PackageActions []PackageAction `json:"packageActions"` // empty = no package manager
	ThreadsLabel   string          `json:"threadsLabel"`   // i18n key: "debug.goroutines" or "debug.threads"
}

// ToolRole is what a tool does for its language.
type ToolRole string

// ToolRole values.
const (
	RoleRuntime        ToolRole = "runtime"
	RoleCompiler       ToolRole = "compiler"
	RoleDebugAdapter   ToolRole = "debugAdapter"
	RoleLanguageServer ToolRole = "languageServer"
	RoleFormatter      ToolRole = "formatter"
)

// ToolSpec describes one tool a language needs and how to get it when it is missing.
type ToolSpec struct {
	ID             string   `json:"id"` // "go", "dlv", "gopls", "python", "debugpy", "cxx", "lldb-dap"...
	Role           ToolRole `json:"role"`
	LabelKey       string   `json:"labelKey"`       // i18n key of the name shown in Settings
	MissingKey     string   `json:"missingKey"`     // i18n key of the "not found" notice
	InstallURL     string   `json:"installUrl"`     // "" when there is none
	InstallCommand string   `json:"installCommand"` // "" when there is none
	// ProvidedBy is the id of the tool that contains this one ("python" for debugpy): it has a
	// status row in Settings but no path of its own, so it cannot be picked.
	ProvidedBy string `json:"providedBy"`
}

// LanguageProfile is everything the frontend needs to know about a language.
type LanguageProfile struct {
	ID           CodeLanguage `json:"id"`
	NameKey      string       `json:"nameKey"`    // i18n key: "codeLanguage.go"...
	Extensions   []string     `json:"extensions"` // lower case, with the dot
	Indent       IndentStyle  `json:"indent"`
	Capabilities Capabilities `json:"capabilities"`
	Tools        []ToolSpec   `json:"tools"`
}

// ToolSource says where a tool comes from (settings, bundled toolchain or PATH).
type ToolSource string

// ToolSource values.
const (
	ToolConfigured ToolSource = "configured"
	ToolBundled    ToolSource = "bundled"
	ToolOnPath     ToolSource = "path"
	ToolMissing    ToolSource = "missing"
)

// ToolStatus is what the IDE found for one tool.
type ToolStatus struct {
	ID           string       `json:"id"`
	CodeLanguage CodeLanguage `json:"codeLanguage"`
	Role         ToolRole     `json:"role"`
	Version      string       `json:"version"` // "" when missing
	Source       ToolSource   `json:"source"`
	Path         string       `json:"path"`
}
