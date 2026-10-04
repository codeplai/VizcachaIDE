package app

import (
	"context"
	"errors"
	"time"

	"github.com/codeplai/VizcachaIDE/wails/internal/domain"
)

// ErrNoUpdate means there is no newer version to download or install.
var ErrNoUpdate = errors.New("there is no update to download")

// Updater keeps the IDE up to date from its published releases. Every change of its state is
// reported through an UpdateSink (event update:state).
type Updater interface {
	// State returns the current state without network access.
	State() domain.UpdateState
	// Check asks for the latest release; the state becomes available, upToDate or failed.
	Check(ctx context.Context) (domain.UpdateState, error)
	// Download fetches and verifies the latest release's file (ErrNoUpdate when there is none).
	Download(ctx context.Context) error
	// Install runs the downloaded installer and closes the IDE, or shows the downloaded file
	// when this copy cannot install itself (portable, macOS, Linux).
	Install() error
}

// UpdateSink is told about every change of the update state. It must not block.
type UpdateSink interface {
	UpdateState(state domain.UpdateState)
}

// UpdateInterval is how often the IDE looks for a new version by itself.
const UpdateInterval = 24 * time.Hour

// AutoUpdate checks for a new version when the settings allow it and the last check is older
// than UpdateInterval, remembers the check and downloads what it finds. Failures stay in the
// updater's state (the IDE keeps working); the error is returned for logging.
func AutoUpdate(ctx context.Context, updater Updater, store SettingsStore, now time.Time) error {
	settings, err := store.Load()
	if err != nil || !settings.CheckUpdates || !dueForCheck(settings.LastUpdateCheck, now) {
		return err
	}
	state, err := updater.Check(ctx)
	if err != nil {
		return err
	}
	if fresh, loadErr := store.Load(); loadErr == nil {
		fresh.LastUpdateCheck = now.UTC().Format(time.RFC3339)
		_ = store.Save(fresh) // best effort: at worst the next start checks again
	}
	if state.Status != domain.UpdateAvailable {
		return nil
	}
	return updater.Download(ctx)
}

func dueForCheck(last string, now time.Time) bool {
	checked, err := time.Parse(time.RFC3339, last)
	return err != nil || now.Sub(checked) >= UpdateInterval
}
