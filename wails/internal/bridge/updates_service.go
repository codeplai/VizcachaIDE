package bridge

import (
	"context"

	"github.com/codeplai/VizcachaIDE/wails/internal/app"
	"github.com/codeplai/VizcachaIDE/wails/internal/domain"
)

// UpdatesService lets the frontend see and drive the update of the IDE. Progress arrives as
// update:state events.
type UpdatesService struct {
	updater app.Updater
}

// NewUpdatesService creates the service.
func NewUpdatesService(updater app.Updater) *UpdatesService {
	return &UpdatesService{updater: updater}
}

// State returns the current update state.
func (s *UpdatesService) State() domain.UpdateState { return s.updater.State() }

// Check looks for a new version now (the "Check now" button).
func (s *UpdatesService) Check() (domain.UpdateState, error) {
	return s.updater.Check(context.Background())
}

// Download starts downloading the new version in the background and returns at once.
func (s *UpdatesService) Download() {
	go func() { _ = s.updater.Download(context.Background()) }() // failures arrive as update:state
}

// Install installs the downloaded version (closing the IDE) or shows the downloaded file.
func (s *UpdatesService) Install() error { return s.updater.Install() }
