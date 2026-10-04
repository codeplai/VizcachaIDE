package app

import (
	"context"
	"testing"
	"time"

	"github.com/codeplai/VizcachaIDE/wails/internal/domain"
)

type fakeUpdater struct {
	result     domain.UpdateStatus
	checks     int
	downloads  int
	checkError error
}

func (f *fakeUpdater) State() domain.UpdateState { return domain.UpdateState{Status: f.result} }
func (f *fakeUpdater) Check(context.Context) (domain.UpdateState, error) {
	f.checks++
	return domain.UpdateState{Status: f.result}, f.checkError
}
func (f *fakeUpdater) Download(context.Context) error { f.downloads++; return nil }
func (f *fakeUpdater) Install() error                 { return nil }

type memoryStore struct{ settings domain.Settings }

func (m *memoryStore) Load() (domain.Settings, error) { return m.settings, nil }
func (m *memoryStore) Save(s domain.Settings) error   { m.settings = s; return nil }

func TestAutoUpdateChecksOnceADayAndDownloads(t *testing.T) {
	now := time.Date(2026, 10, 4, 12, 0, 0, 0, time.UTC)
	store := &memoryStore{settings: domain.DefaultSettings()}
	updater := &fakeUpdater{result: domain.UpdateAvailable}

	if err := AutoUpdate(context.Background(), updater, store, now); err != nil {
		t.Fatal(err)
	}
	if updater.checks != 1 || updater.downloads != 1 || store.settings.LastUpdateCheck != "2026-10-04T12:00:00Z" {
		t.Fatalf("checks %d, downloads %d, last %q", updater.checks, updater.downloads, store.settings.LastUpdateCheck)
	}
	_ = AutoUpdate(context.Background(), updater, store, now.Add(23*time.Hour))
	if updater.checks != 1 {
		t.Errorf("checked again before a day passed")
	}
	_ = AutoUpdate(context.Background(), updater, store, now.Add(25*time.Hour))
	if updater.checks != 2 {
		t.Errorf("did not check after a day")
	}
}

func TestAutoUpdateRespectsTheSettingAndDoesNotDownloadWhenUpToDate(t *testing.T) {
	store := &memoryStore{settings: domain.DefaultSettings()}
	store.settings.CheckUpdates = false
	updater := &fakeUpdater{result: domain.UpdateAvailable}
	_ = AutoUpdate(context.Background(), updater, store, time.Now())
	if updater.checks != 0 {
		t.Error("checked although the setting is off")
	}
	store.settings.CheckUpdates = true
	updater.result = domain.UpdateUpToDate
	_ = AutoUpdate(context.Background(), updater, store, time.Now())
	if updater.checks != 1 || updater.downloads != 0 {
		t.Errorf("checks %d, downloads %d", updater.checks, updater.downloads)
	}
}
