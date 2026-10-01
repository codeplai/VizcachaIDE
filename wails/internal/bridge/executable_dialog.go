package bridge

import (
	"fmt"
	goruntime "runtime"

	"github.com/wailsapp/wails/v2/pkg/runtime"
)

// NewExecutableDialog returns the native "open file" dialog used by SettingsService to
// choose go, dlv or gopls. translate gives the dialog title in the interface language.
// Before the application starts (no context) it returns "" as if the user had cancelled.
func NewExecutableDialog(source ContextSource, translate func(key string, data map[string]any) string) ExecutablePicker {
	return func(tool string) (string, error) {
		ctx := source.Context()
		if ctx == nil {
			return "", nil
		}
		options := runtime.OpenDialogOptions{Title: translate("dialogs.pickTool", map[string]any{"tool": tool})}
		if goruntime.GOOS == "windows" {
			options.Filters = []runtime.FileFilter{{DisplayName: tool + ".exe", Pattern: "*.exe"}}
		}
		path, err := runtime.OpenFileDialog(ctx, options)
		if err != nil {
			return "", fmt.Errorf("open file dialog: %w", err)
		}
		return path, nil
	}
}
