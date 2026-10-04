package domain

// Language values stored in Settings.Language (the UI language, not a CodeLanguage).
const (
	LanguageAuto = "auto"
	LanguageEN   = "en"
	LanguageES   = "es"
)

// Theme values stored in Settings.Theme.
const (
	ThemeSystem = "system"
	ThemeLight  = "light"
	ThemeDark   = "dark"
)

// ServerStatus is the state of the language server (event lsp:status).
type ServerStatus string

// ServerStatus values.
const (
	ServerStarting    ServerStatus = "starting"
	ServerReady       ServerStatus = "ready"
	ServerUnavailable ServerStatus = "unavailable"
)

// Settings are the user preferences saved between sessions.
type Settings struct {
	// Language is the UI language (LanguageAuto, LanguageEN or LanguageES).
	Language string `json:"language"`
	Theme    string `json:"theme"`
	FontSize int    `json:"fontSize"`
	// ToolPaths are the executables chosen in Settings, by ToolSpec.ID ("go", "dlv", "gopls"...).
	// A missing or empty entry means "find it automatically".
	ToolPaths  map[string]string `json:"toolPaths"`
	FirstRun   bool              `json:"firstRun"`
	LastFolder string            `json:"lastFolder"`
	// DefaultCodeLanguage is the language of new files.
	DefaultCodeLanguage CodeLanguage `json:"defaultCodeLanguage"`
	// EnabledCodeLanguages are the languages chosen in the first-run wizard; empty means all.
	EnabledCodeLanguages []CodeLanguage `json:"enabledCodeLanguages"`
	// FormatOnSave formats a file every time it is saved, when its language can format.
	FormatOnSave bool `json:"formatOnSave"`
	// RecentFiles are the last opened files, newest first (at most MaxRecentFiles).
	RecentFiles []string `json:"recentFiles"`
}

// MaxRecentFiles is how many entries Settings.RecentFiles keeps.
const MaxRecentFiles = 10

// DefaultSettings returns the settings of a fresh installation.
func DefaultSettings() Settings {
	return Settings{
		Language: LanguageAuto, Theme: ThemeSystem, FontSize: 14, ToolPaths: map[string]string{},
		FirstRun: true, DefaultCodeLanguage: CodeLanguageGo, EnabledCodeLanguages: []CodeLanguage{},
		FormatOnSave: true, RecentFiles: []string{},
	}
}
