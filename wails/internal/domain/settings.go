package domain

// Language values stored in Settings.Language.
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
	Language   string `json:"language"`
	Theme      string `json:"theme"`
	FontSize   int    `json:"fontSize"`
	GoPath     string `json:"goPath"`
	DelvePath  string `json:"delvePath"`
	GoplsPath  string `json:"goplsPath"`
	FirstRun   bool   `json:"firstRun"`
	LastFolder string `json:"lastFolder"`
}

// DefaultSettings returns the settings of a fresh installation.
func DefaultSettings() Settings {
	return Settings{Language: LanguageAuto, Theme: ThemeSystem, FontSize: 14, FirstRun: true}
}
