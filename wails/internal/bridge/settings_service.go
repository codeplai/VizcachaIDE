package bridge

import (
	"fmt"
	"sync"

	"github.com/codeplai/VizcachaIDE/wails/internal/app"
	"github.com/codeplai/VizcachaIDE/wails/internal/domain"
)

// SettingsService reads and saves the user preferences. Emits settings:changed
// after every successful Save. Owned by track F2 on the frontend side.
type SettingsService struct {
	sink  app.EventSink
	store app.SettingsStore
}

// NewSettingsService creates the service.
func NewSettingsService(sink app.EventSink, store app.SettingsStore) *SettingsService {
	return &SettingsService{sink: sink, store: store}
}

// Get returns the current settings.
func (s *SettingsService) Get() (domain.Settings, error) {
	settings, err := s.store.Load()
	if err != nil {
		return domain.DefaultSettings(), fmt.Errorf("load settings: %w", err)
	}
	return settings, nil
}

// Save stores the settings and tells the frontend.
func (s *SettingsService) Save(settings domain.Settings) error {
	if err := s.store.Save(settings); err != nil {
		return fmt.Errorf("save settings: %w", err)
	}
	s.sink.SettingsChanged(settings)
	return nil
}

// MemorySettingsStore keeps the settings in memory. W0 STUB: the real adapter
// (JSON in the user's config folder) goes in internal/adapters/settings.
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
