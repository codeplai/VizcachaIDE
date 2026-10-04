package lldbdap

import (
	"bytes"
	_ "embed" // the LLDB script is embedded data
	"os"
	"path/filepath"
)

// enumsFix makes LLDB read the payload of Rust enums (Some(7), Ok(5)) right on *-gnu targets,
// where LLDB 23 puts it at the wrong offset (data/vizcacha_rust_enums.py; found in the M3 QA).
//
//go:embed data/vizcacha_rust_enums.py
var enumsFix []byte

// enumsFixPath writes the script once to the IDE's cache and returns its path, "" when it cannot
// (the values of enums then show wrong, the rest of debugging works).
func enumsFixPath() string {
	cache, err := os.UserCacheDir()
	if err != nil {
		return ""
	}
	path := filepath.Join(cache, "VizcachaIDE", "lldb", "vizcacha_rust_enums.py")
	if current, err := os.ReadFile(path); err == nil && bytes.Equal(current, enumsFix) {
		return path
	}
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return ""
	}
	if err := os.WriteFile(path, enumsFix, 0o600); err != nil {
		return ""
	}
	return path
}
