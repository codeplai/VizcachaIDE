package bridge

import (
	"fmt"
	"path/filepath"
	"strings"

	"github.com/codeplai/VizcachaIDE/wails/internal/app"
	"github.com/wailsapp/wails/v2/pkg/runtime"
)

// Translator gives a text in the interface language.
type Translator func(key string, data map[string]any) string

// NewFilesServiceWithTexts is NewFilesService plus the translator for the titles of the native
// dialogs. It is a function, not a method, so Wails does not expose it to the frontend.
func NewFilesServiceWithTexts(source ContextSource, settings app.SettingsStore, watcher app.FileWatcher, translate Translator) *FilesService {
	service := NewFilesService(source, settings, watcher)
	service.translate = translate
	return service
}

// UseLanguages gives the service the languages whose file extensions the dialogs offer. Without
// it the dialogs only offer "all files". It returns the service for chaining.
func (s *FilesService) UseLanguages(registry *app.LanguageRegistry) *FilesService {
	s.registry = registry
	return s
}

// OpenFileDialog lets the user pick a file. Cancelling returns "" with no error.
func (s *FilesService) OpenFileDialog() (string, error) {
	ctx := s.context.Context()
	if ctx == nil {
		return "", nil
	}
	path, err := withStartFolder(s.lastFolder(), func(start string) (string, error) {
		return runtime.OpenFileDialog(ctx, runtime.OpenDialogOptions{
			Title:            s.text("dialogs.openFile"),
			DefaultDirectory: start,
			Filters:          s.fileFilters(),
		})
	})
	if err != nil {
		return "", fmt.Errorf("open file dialog: %w", err)
	}
	return path, nil
}

// SaveFileDialog asks where to save a file. folder is the suggested folder (empty means the last
// folder). A path without extension gets the extension of the suggested name. Cancelling returns "" with no error.
func (s *FilesService) SaveFileDialog(suggestedName, folder string) (string, error) {
	ctx := s.context.Context()
	if ctx == nil {
		return "", nil
	}
	path, err := withStartFolder(defaultDialogFolder(folder, s.lastFolder()), func(start string) (string, error) {
		return runtime.SaveFileDialog(ctx, runtime.SaveDialogOptions{
			Title:            s.text("dialogs.saveFile"),
			DefaultDirectory: start,
			DefaultFilename:  suggestedName,
			Filters:          s.fileFilters(),
		})
	})
	if err != nil {
		return "", fmt.Errorf("save file dialog: %w", err)
	}
	return withExtensionOf(suggestedName, path), nil
}

// fileFilters offers one filter per language (Go, Python, C++...) and one for all files.
func (s *FilesService) fileFilters() []runtime.FileFilter {
	filters := []runtime.FileFilter{}
	if s.registry != nil {
		for _, profile := range s.registry.Profiles() {
			pattern := patternOf(profile.Extensions)
			filters = append(filters, runtime.FileFilter{
				DisplayName: s.text(profile.NameKey) + " (" + pattern + ")",
				Pattern:     pattern,
			})
		}
	}
	return append(filters, runtime.FileFilter{DisplayName: s.text("dialogs.allFiles"), Pattern: "*.*"})
}

// patternOf turns [".py", ".pyw"] into "*.py;*.pyw".
func patternOf(extensions []string) string {
	patterns := make([]string, 0, len(extensions))
	for _, extension := range extensions {
		patterns = append(patterns, "*"+extension)
	}
	return strings.Join(patterns, ";")
}

func (s *FilesService) text(key string) string {
	if s.translate == nil {
		return key
	}
	return s.translate(key, nil)
}

// withExtensionOf appends the extension of suggestedName to a non-empty path that has none.
func withExtensionOf(suggestedName, path string) string {
	if path == "" || filepath.Ext(path) != "" {
		return path
	}
	return path + filepath.Ext(suggestedName)
}

// defaultDialogFolder prefers the requested folder when it exists and falls back to the last one.
func defaultDialogFolder(folder, last string) string {
	if existingFolder(folder) != "" {
		return folder
	}
	return last
}
