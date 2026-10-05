package app

import (
	"fmt"
	"path/filepath"
	"strings"

	"github.com/codeplai/VizcachaIDE/wails/internal/domain"
)

// LanguageSupport is everything the IDE needs to work with one programming language: its
// profile and one adapter per port. Optional capabilities are nil ports, never empty adapters.
type LanguageSupport struct {
	Profile        domain.LanguageProfile
	Runner         ProgramRunner
	Debugger       Debugger
	LanguageServer LanguageServer
	Explainer      ErrorExplainer
	Console        Console         // nil when Capabilities.Console is false
	Formatter      CodeFormatter   // nil when Capabilities.Format is false
	Checker        CodeChecker     // nil when Capabilities.Check is false
	Packages       PackageManager  // nil when Capabilities.PackageActions is empty
	Scaffold       ProjectScaffold // nil when the language cannot create projects

	// unavailable marks a language whose adapters do not exist yet (see UnavailableSupport).
	unavailable bool
}

// LanguageRegistry maps extensions and ids to supports. It is pure: it creates nothing.
type LanguageRegistry struct {
	supports    []LanguageSupport
	byExtension map[string]int
	byID        map[domain.CodeLanguage]int
	defaultIdx  int
}

// NewLanguageRegistry registers the supports in order. It fails with ErrInconsistentProfile when
// a capability and its port disagree (Console, Format, Check, PackageActions), when two profiles
// claim the same id or extension, or when defaultID is not registered, so the IDE fails at
// start-up and never in the middle of a lesson. The capability check is skipped for supports
// made by UnavailableSupport.
func NewLanguageRegistry(defaultID domain.CodeLanguage, supports ...LanguageSupport) (*LanguageRegistry, error) {
	registry := &LanguageRegistry{
		supports:    make([]LanguageSupport, 0, len(supports)),
		byExtension: map[string]int{},
		byID:        map[domain.CodeLanguage]int{},
	}
	for _, support := range supports {
		if err := registry.add(support); err != nil {
			return nil, err
		}
	}
	index, ok := registry.byID[defaultID]
	if !ok {
		return nil, fmt.Errorf("%w: the default language %q is not registered", ErrInconsistentProfile, defaultID)
	}
	registry.defaultIdx = index
	return registry, nil
}

func (r *LanguageRegistry) add(support LanguageSupport) error {
	profile := support.Profile
	if err := checkSupport(support); err != nil {
		return err
	}
	if _, taken := r.byID[profile.ID]; taken {
		return fmt.Errorf("%w: language %q is registered twice", ErrInconsistentProfile, profile.ID)
	}
	index := len(r.supports)
	for _, extension := range profile.Extensions {
		key := strings.ToLower(extension)
		if owner, taken := r.byExtension[key]; taken {
			return fmt.Errorf("%w: %q is claimed by %q and %q", ErrInconsistentProfile, key, r.supports[owner].Profile.ID, profile.ID)
		}
		r.byExtension[key] = index
	}
	r.byID[profile.ID] = index
	r.supports = append(r.supports, support)
	return nil
}

// checkSupport verifies one support on its own: id, required ports and capability/port pairs.
func checkSupport(support LanguageSupport) error {
	profile := support.Profile
	if profile.ID == "" {
		return fmt.Errorf("%w: a profile has no id", ErrInconsistentProfile)
	}
	if support.unavailable {
		return nil
	}
	if support.Runner == nil || support.Debugger == nil || support.LanguageServer == nil || support.Explainer == nil {
		return fmt.Errorf("%w: %q lacks a runner, debugger, language server or explainer", ErrInconsistentProfile, profile.ID)
	}
	caps := profile.Capabilities
	pairs := []struct {
		name    string
		enabled bool
		present bool
	}{
		{"console", caps.Console, support.Console != nil},
		{"format", caps.Format, support.Formatter != nil},
		{"check", caps.Check, support.Checker != nil},
		{"package actions", len(caps.PackageActions) > 0, support.Packages != nil},
	}
	for _, pair := range pairs {
		if pair.enabled != pair.present {
			return fmt.Errorf("%w: %q declares %s = %t but its port is present = %t",
				ErrInconsistentProfile, profile.ID, pair.name, pair.enabled, pair.present)
		}
	}
	return nil
}

// Profiles returns the profile of every language, in registration order.
func (r *LanguageRegistry) Profiles() []domain.LanguageProfile {
	profiles := make([]domain.LanguageProfile, 0, len(r.supports))
	for _, support := range r.supports {
		profiles = append(profiles, support.Profile)
	}
	return profiles
}

// ForPath finds the language of a file by its extension, ignoring case.
func (r *LanguageRegistry) ForPath(path string) (LanguageSupport, bool) {
	index, ok := r.byExtension[strings.ToLower(filepath.Ext(path))]
	if !ok {
		return LanguageSupport{}, false
	}
	return r.supports[index], true
}

// ForID finds a language by its id.
func (r *LanguageRegistry) ForID(id domain.CodeLanguage) (LanguageSupport, bool) {
	index, ok := r.byID[id]
	if !ok {
		return LanguageSupport{}, false
	}
	return r.supports[index], true
}

// Default is the language used when nothing says otherwise.
func (r *LanguageRegistry) Default() LanguageSupport { return r.supports[r.defaultIdx] }

// All returns every support, in registration order. The caller must not modify the slice.
func (r *LanguageRegistry) All() []LanguageSupport { return r.supports }
