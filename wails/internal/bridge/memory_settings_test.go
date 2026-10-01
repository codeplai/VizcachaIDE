package bridge

import (
	"sync"

	"github.com/codeplai/VizcachaIDE/wails/internal/app"
	"github.com/codeplai/VizcachaIDE/wails/internal/domain"
)

// MemorySettingsStore keeps the settings in memory, for tests.
type MemorySettingsStore struct {
	mu       sync.Mutex
	settings domain.Settings
}

var _ app.SettingsStore = (*MemorySettingsStore)(nil)

// NewMemorySettingsStore starts with the default settings.
func NewMemorySettingsStore() *MemorySettingsStore {
	return &MemorySettingsStore{settings: domain.DefaultSettings()}
}

// Load implements app.SettingsStore.
func (m *MemorySettingsStore) Load() (domain.Settings, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	return m.settings, nil
}

// Save implements app.SettingsStore.
func (m *MemorySettingsStore) Save(settings domain.Settings) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.settings = settings
	return nil
}
