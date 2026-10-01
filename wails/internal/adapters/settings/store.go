package settings

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"sync"

	"github.com/codeplai/VizcachaIDE/wails/internal/app"
	"github.com/codeplai/VizcachaIDE/wails/internal/domain"
)

const (
	folderName = "VizcachaIDE"
	fileName   = "settings.json"
	backupName = "settings.json.bak"
)

// Store keeps the settings in one JSON file. It is safe for concurrent use.
type Store struct {
	path string
	mu   sync.Mutex
}

var _ app.SettingsStore = (*Store)(nil)

// NewStore creates a store that reads and writes the given file.
func NewStore(path string) *Store { return &Store{path: path} }

// DefaultPath is os.UserConfigDir()/VizcachaIDE/settings.json.
func DefaultPath() (string, error) {
	dir, err := os.UserConfigDir()
	if err != nil {
		return "", fmt.Errorf("find the user config folder: %w", err)
	}
	return filepath.Join(dir, folderName, fileName), nil
}

// NewDefaultStore creates the store in the user's config folder.
func NewDefaultStore() (*Store, error) {
	path, err := DefaultPath()
	if err != nil {
		return nil, err
	}
	return NewStore(path), nil
}

// Load returns the saved settings. A missing file gives the defaults; fields missing in
// an older file take their default; a corrupt file is renamed to settings.json.bak and
// the defaults are returned.
func (s *Store) Load() (domain.Settings, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	data, err := os.ReadFile(s.path)
	if errors.Is(err, fs.ErrNotExist) {
		return domain.DefaultSettings(), nil
	}
	if err != nil {
		return domain.DefaultSettings(), fmt.Errorf("read settings: %w", err)
	}
	// Decoding over the defaults keeps the default of every field the file does not have.
	loaded := domain.DefaultSettings()
	if err := json.Unmarshal(data, &loaded); err != nil {
		s.quarantine()
		return domain.DefaultSettings(), nil
	}
	return loaded, nil
}

// quarantine moves an unreadable file out of the way so the next Save starts clean.
func (s *Store) quarantine() {
	backup := filepath.Join(filepath.Dir(s.path), backupName)
	_ = os.Remove(backup)
	_ = os.Rename(s.path, backup) // best effort: the defaults are used either way
}

// Save writes the settings atomically (temporary file, then rename).
func (s *Store) Save(settings domain.Settings) error {
	data, err := json.MarshalIndent(settings, "", "  ")
	if err != nil {
		return fmt.Errorf("encode settings: %w", err)
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	return writeAtomically(s.path, data)
}

func writeAtomically(path string, data []byte) error {
	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0o700); err != nil {
		return fmt.Errorf("create settings folder: %w", err)
	}
	temp, err := os.CreateTemp(dir, fileName+".*.tmp")
	if err != nil {
		return fmt.Errorf("create temporary settings file: %w", err)
	}
	_, writeErr := temp.Write(data)
	closeErr := temp.Close()
	if err := errors.Join(writeErr, closeErr); err != nil {
		_ = os.Remove(temp.Name())
		return fmt.Errorf("write settings: %w", err)
	}
	if err := os.Rename(temp.Name(), path); err != nil {
		_ = os.Remove(temp.Name())
		return fmt.Errorf("replace settings file: %w", err)
	}
	return nil
}
