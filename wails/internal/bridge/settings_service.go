package bridge

import (
	"context"
	"fmt"
	"strings"

	"github.com/codeplai/VizcachaIDE/wails/internal/app"
	"github.com/codeplai/VizcachaIDE/wails/internal/domain"
)

// Tool names accepted by PickExecutable.
const (
	toolGo    = "go"
	toolDelve = "dlv"
	toolGopls = "gopls"
)

// ExecutablePicker asks the user for the executable of a tool and returns its path,
// or "" when the user cancels.
type ExecutablePicker func(tool string) (string, error)

// SettingsService reads and saves the user preferences. Emits settings:changed
// after every successful Save.
type SettingsService struct {
	sink      app.EventSink
	store     app.SettingsStore
	language  *LanguageResolver
	toolchain app.Toolchain
	pick      ExecutablePicker
}

// NewSettingsService creates the service. Without UseTools, PickExecutable is unavailable.
func NewSettingsService(sink app.EventSink, store app.SettingsStore, language *LanguageResolver) *SettingsService {
	return &SettingsService{sink: sink, store: store, language: language}
}

// UseTools gives the service what PickExecutable needs: the toolchain that detects the
// tools and the dialog that asks for a file. It returns the service for chaining.
func (s *SettingsService) UseTools(toolchain app.Toolchain, pick ExecutablePicker) *SettingsService {
	s.toolchain, s.pick = toolchain, pick
	return s
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

// ResolvedLanguage is "en" or "es": the language of the settings, or the system's when
// the setting is "auto".
func (s *SettingsService) ResolvedLanguage() string { return s.language.Current() }

// PickExecutable asks the user for the executable of a tool ("go", "dlv" or "gopls"),
// saves it in the settings (emitting settings:changed) and returns the tools detected
// again. If the user cancels, nothing changes.
func (s *SettingsService) PickExecutable(tool string) (domain.ToolchainInfo, error) {
	if s.toolchain == nil || s.pick == nil {
		return domain.ToolchainInfo{}, fmt.Errorf("pick %s: %w", tool, app.ErrToolNotFound)
	}
	if err := s.chooseAndSave(tool); err != nil {
		return domain.ToolchainInfo{}, err
	}
	return s.toolchain.Info(context.Background()), nil
}

func (s *SettingsService) chooseAndSave(tool string) error {
	if tool != toolGo && tool != toolDelve && tool != toolGopls {
		return fmt.Errorf("unknown tool %q", tool)
	}
	path, err := s.pick(tool)
	if err != nil {
		return fmt.Errorf("choose %s: %w", tool, err)
	}
	if strings.TrimSpace(path) == "" {
		return nil
	}
	current, err := s.Get()
	if err != nil {
		return err
	}
	if current.ToolPaths == nil {
		current.ToolPaths = map[string]string{}
	}
	current.ToolPaths[tool] = path
	return s.Save(current)
}
