package app

import "github.com/codeplai/VizcachaIDE/wails/internal/domain"

// WindowStateStore persists the window geometry. It is separate from SettingsStore so
// the frontend, which saves the whole settings object, cannot overwrite it.
type WindowStateStore interface {
	// Load returns the saved state, or false when there is none or it is unreadable.
	Load() (domain.WindowState, bool)
	Save(state domain.WindowState) error
}
