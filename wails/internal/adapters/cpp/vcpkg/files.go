package vcpkg

import (
	"os"
	"path/filepath"
	"strings"
)

// Small helpers of the locator: files, the application folder and the environment.

func isFile(path string) bool {
	info, err := os.Stat(path)
	return err == nil && !info.IsDir()
}

func applicationDirectory() string {
	executable, err := os.Executable()
	if err != nil {
		return "."
	}
	return filepath.Dir(executable)
}

// environmentValue returns the value of a variable in a "KEY=value" list, "" if absent.
func environmentValue(environment []string, key string) string {
	for index := len(environment) - 1; index >= 0; index-- {
		name, value, found := strings.Cut(environment[index], "=")
		if found && strings.EqualFold(name, key) {
			return value
		}
	}
	return ""
}
