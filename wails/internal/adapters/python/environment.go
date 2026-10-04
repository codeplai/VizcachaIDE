package python

import (
	"strings"

	"github.com/codeplai/VizcachaIDE/wails/internal/domain"
)

// Environment returns the variables every Python process runs with (programs, debugpy, pylsp,
// ruff, the console and pip): UTF-8 everywhere, unbuffered output and no __pycache__ folders in
// the student's folders. With the bundled interpreter it also ignores packages that another
// Python installed for the user (PYTHONNOUSERSITE).
func Environment(base []string, interpreter Interpreter) map[string]string {
	env := parseEnvironment(base)
	env["PYTHONUTF8"] = "1"
	env["PYTHONIOENCODING"] = "utf-8"
	env["PYTHONUNBUFFERED"] = "1"
	env["PYTHONDONTWRITEBYTECODE"] = "1"
	if interpreter.Source == domain.ToolBundled {
		env["PYTHONNOUSERSITE"] = "1"
	}
	return env
}

// EnvironmentList turns the map back into "NAME=value" entries for exec.Cmd.
func EnvironmentList(env map[string]string) []string {
	list := make([]string, 0, len(env))
	for name, value := range env {
		list = append(list, name+"="+value)
	}
	return list
}

// parseEnvironment turns "NAME=value" entries into a map. Windows keeps hidden entries such
// as "=C:=C:\dir", whose name starts with "=".
func parseEnvironment(entries []string) map[string]string {
	env := make(map[string]string, len(entries))
	for _, entry := range entries {
		index := strings.Index(entry[min(1, len(entry)):], "=")
		if index < 0 {
			continue
		}
		index++
		env[entry[:index]] = entry[index+1:]
	}
	return env
}

// searchPath returns the PATH of base (Windows usually spells it "Path").
func searchPath(base []string) string {
	for name, value := range parseEnvironment(base) {
		if strings.EqualFold(name, "PATH") {
			return value
		}
	}
	return ""
}
