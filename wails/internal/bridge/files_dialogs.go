package bridge

import (
	"fmt"
	"path/filepath"

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
// folder). A name without extension gets ".go". Cancelling returns "" with no error.
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
	return withGoExtension(path), nil
}

func (s *FilesService) fileFilters() []runtime.FileFilter {
	return []runtime.FileFilter{
		{DisplayName: s.text("dialogs.goFiles"), Pattern: "*.go"},
		{DisplayName: s.text("dialogs.allFiles"), Pattern: "*.*"},
	}
}

func (s *FilesService) text(key string) string {
	if s.translate == nil {
		return key
	}
	return s.translate(key, nil)
}

// withGoExtension appends ".go" to a non-empty path that has no extension.
func withGoExtension(path string) string {
	if path == "" || filepath.Ext(path) != "" {
		return path
	}
	return path + ".go"
}

// defaultDialogFolder prefers the requested folder when it exists and falls back to the last one.
func defaultDialogFolder(folder, last string) string {
	if existingFolder(folder) != "" {
		return folder
	}
	return last
}
