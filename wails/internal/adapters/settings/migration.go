package settings

import (
	"encoding/json"
	"strings"

	"github.com/codeplai/VizcachaIDE/wails/internal/domain"
)

// legacyToolPaths are the three keys a 2.0 settings.json used for the Go tools.
type legacyToolPaths struct {
	GoPath    string `json:"goPath"`
	DelvePath string `json:"delvePath"`
	GoplsPath string `json:"goplsPath"`
}

// migrateLegacyToolPaths copies goPath, delvePath and goplsPath of a 2.0 file into
// ToolPaths["go"|"dlv"|"gopls"] and saves the new file, so the old keys disappear. A path that
// ToolPaths already has wins. The caller holds the store's lock. The migration is best effort:
// if saving fails the migrated paths are still returned and the next Load migrates again.
func (s *Store) migrateLegacyToolPaths(data []byte, loaded *domain.Settings) {
	var legacy legacyToolPaths
	if json.Unmarshal(data, &legacy) != nil {
		return
	}
	moved := false
	for tool, path := range map[string]string{"go": legacy.GoPath, "dlv": legacy.DelvePath, "gopls": legacy.GoplsPath} {
		path = strings.TrimSpace(path)
		if path == "" || strings.TrimSpace(loaded.ToolPaths[tool]) != "" {
			continue
		}
		if loaded.ToolPaths == nil {
			loaded.ToolPaths = map[string]string{}
		}
		loaded.ToolPaths[tool] = path
		moved = true
	}
	if !moved && legacy == (legacyToolPaths{}) {
		return
	}
	encoded, err := json.MarshalIndent(loaded, "", "  ")
	if err != nil {
		return
	}
	_ = writeAtomically(s.path, encoded)
}
