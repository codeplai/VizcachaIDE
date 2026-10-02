// Package windowstate keeps the window geometry in window.json, next to settings.json.
package windowstate

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"sync"

	"github.com/codeplai/VizcachaIDE/wails/internal/app"
	"github.com/codeplai/VizcachaIDE/wails/internal/domain"
)

const (
	folderName = "VizcachaIDE"
	fileName   = "window.json"
)

// Store keeps the window state in one JSON file. It is safe for concurrent use.
type Store struct {
	path string
	mu   sync.Mutex
}

var _ app.WindowStateStore = (*Store)(nil)

// NewStore creates a store that reads and writes the given file.
func NewStore(path string) *Store { return &Store{path: path} }

// NewDefaultStore uses os.UserConfigDir()/VizcachaIDE/window.json.
func NewDefaultStore() (*Store, error) {
	dir, err := os.UserConfigDir()
	if err != nil {
		return nil, fmt.Errorf("find the user config folder: %w", err)
	}
	return NewStore(filepath.Join(dir, folderName, fileName)), nil
}

// Load returns the saved state. A missing or corrupt file gives (defaults, false).
func (s *Store) Load() (domain.WindowState, bool) {
	s.mu.Lock()
	defer s.mu.Unlock()
	data, err := os.ReadFile(s.path)
	if err != nil {
		return domain.DefaultWindowState(), false
	}
	var state domain.WindowState
	if err := json.Unmarshal(data, &state); err != nil {
		return domain.DefaultWindowState(), false
	}
	return state, true
}

// Save writes the state atomically (temporary file, then rename).
func (s *Store) Save(state domain.WindowState) error {
	data, err := json.MarshalIndent(state, "", "  ")
	if err != nil {
		return fmt.Errorf("encode window state: %w", err)
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	dir := filepath.Dir(s.path)
	if err := os.MkdirAll(dir, 0o700); err != nil {
		return fmt.Errorf("create window state folder: %w", err)
	}
	temp, err := os.CreateTemp(dir, fileName+".*.tmp")
	if err != nil {
		return fmt.Errorf("create temporary window file: %w", err)
	}
	_, writeErr := temp.Write(data)
	closeErr := temp.Close()
	if err := errors.Join(writeErr, closeErr); err != nil {
		_ = os.Remove(temp.Name())
		return fmt.Errorf("write window state: %w", err)
	}
	if err := os.Rename(temp.Name(), s.path); err != nil {
		_ = os.Remove(temp.Name())
		return fmt.Errorf("replace window file: %w", err)
	}
	return nil
}
