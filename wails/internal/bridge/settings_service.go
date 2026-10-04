package bridge

import (
	"context"
	"fmt"
	"strings"

	"github.com/codeplai/VizcachaIDE/wails/internal/app"
	"github.com/codeplai/VizcachaIDE/wails/internal/domain"
)

// ExecutablePicker asks the user for the executable of a tool and returns its path,
// or "" when the user cancels.
type ExecutablePicker func(tool string) (string, error)

// SettingsService reads and saves the user preferences. Emits settings:changed
// after every successful Save.
type SettingsService struct {
	sink     app.EventSink
	store    app.SettingsStore
	language *LanguageResolver
	registry *app.LanguageRegistry
	pick     ExecutablePicker
}

// NewSettingsService creates the service. The registry says which tools exist and detects them.
// Without UseTools, PickExecutable is unavailable.
func NewSettingsService(sink app.EventSink, store app.SettingsStore, language *LanguageResolver, registry *app.LanguageRegistry) *SettingsService {
	return &SettingsService{sink: sink, store: store, language: language, registry: registry}
}

// UseTools gives the service the dialog that asks for a file, which PickExecutable needs.
// It returns the service for chaining.
func (s *SettingsService) UseTools(pick ExecutablePicker) *SettingsService {
	s.pick = pick
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

// PickExecutable asks the user for the executable of a tool (a ToolSpec.ID of any language),
// saves it in the settings (emitting settings:changed) and returns the tools detected again.
// Tools that live inside another one (ProvidedBy) have no path of their own and are rejected.
// If the user cancels, nothing changes.
func (s *SettingsService) PickExecutable(toolID string) ([]domain.ToolStatus, error) {
	if s.pick == nil {
		return nil, fmt.Errorf("pick %s: %w", toolID, app.ErrToolNotFound)
	}
	if err := s.checkPickable(toolID); err != nil {
		return nil, err
	}
	if err := s.chooseAndSave(toolID); err != nil {
		return nil, err
	}
	return toolStatuses(context.Background(), s.registry), nil
}

// checkPickable accepts only the ids of tools that have an executable of their own.
func (s *SettingsService) checkPickable(toolID string) error {
	for _, profile := range s.registry.Profiles() {
		for _, spec := range profile.Tools {
			if spec.ID != toolID {
				continue
			}
			if spec.ProvidedBy != "" {
				return fmt.Errorf("tool %q is provided by %q and cannot be chosen", toolID, spec.ProvidedBy)
			}
			return nil
		}
	}
	return fmt.Errorf("unknown tool %q", toolID)
}

func (s *SettingsService) chooseAndSave(tool string) error {
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
