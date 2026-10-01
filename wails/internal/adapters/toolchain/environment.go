package toolchain

import (
	"os"
	"sort"
	"strings"
)

// parseEnvironment turns "NAME=value" entries into a map. Windows keeps hidden
// entries such as "=C:=C:\dir", whose name starts with "=".
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

// pathVariableName is PATH as spelled in env (Windows usually spells it "Path").
func pathVariableName(env map[string]string) string {
	for name := range env {
		if strings.EqualFold(name, "PATH") {
			return name
		}
	}
	return "PATH"
}

// buildEnvironment returns the variables Go tools and the user's program run with.
// With the bundled toolchain it sets GOROOT, GOTOOLCHAIN=local and puts its folders
// at the front of PATH, so the app never downloads another Go.
func buildEnvironment(base []string, locator *Locator) map[string]string {
	env := parseEnvironment(base)
	if locator.Locate(ToolGo).Origin != OriginBundled {
		return env
	}
	env["GOROOT"] = locator.BundledGoroot()
	if env["GOTOOLCHAIN"] == "" {
		env["GOTOOLCHAIN"] = "local"
	}
	name := pathVariableName(env)
	parts := locator.BundledBinDirectories()
	if current := env[name]; current != "" {
		parts = append(parts, current)
	}
	env[name] = strings.Join(parts, string(os.PathListSeparator))
	return env
}

// environmentList is env in the "NAME=value" form os/exec wants.
func environmentList(env map[string]string) []string {
	list := make([]string, 0, len(env))
	for name, value := range env {
		list = append(list, name+"="+value)
	}
	sort.Strings(list)
	return list
}
