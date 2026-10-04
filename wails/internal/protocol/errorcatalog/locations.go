package errorcatalog

import (
	"path/filepath"
	"regexp"
	"strconv"
	"strings"

	"github.com/codeplai/VizcachaIDE/wails/internal/domain"
)

var windowsAbsolute = regexp.MustCompile(`^[A-Za-z]:[\\/]`)

// ResolvePath keeps absolute paths (Windows or POSIX) as the tool printed them and joins
// relative ones ("./main.go", ".\util\calc.go") to the working directory.
func ResolvePath(rawPath, workingDir string) string {
	if windowsAbsolute.MatchString(rawPath) || strings.HasPrefix(rawPath, "/") || strings.HasPrefix(rawPath, `\`) {
		return rawPath
	}
	relative := filepath.FromSlash(strings.ReplaceAll(rawPath, `\`, "/"))
	if workingDir == "" {
		return filepath.Clean(relative)
	}
	return filepath.Join(workingDir, relative)
}

// LocationFrom builds a SourceLocation from the groups path, line and (optional) column.
func LocationFrom(groups map[string]string, workingDir string) domain.SourceLocation {
	line, _ := strconv.Atoi(groups["line"])
	column := 1
	if parsed, err := strconv.Atoi(groups["column"]); err == nil {
		column = parsed
	}
	return domain.SourceLocation{File: ResolvePath(groups["path"], workingDir), Line: line, Column: column}
}

// NamedGroups returns the named groups of the first match of pattern in text, or nil when there
// is none.
func NamedGroups(pattern *regexp.Regexp, text string) map[string]string {
	values := pattern.FindStringSubmatch(text)
	if values == nil {
		return nil
	}
	groups := map[string]string{}
	for index, name := range pattern.SubexpNames() {
		if name != "" {
			groups[name] = values[index]
		}
	}
	return groups
}

// SplitLines splits tool output into lines, accepting "\r\n" and dropping the empty line that
// a final newline leaves.
func SplitLines(text string) []string {
	lines := strings.Split(strings.ReplaceAll(text, "\r\n", "\n"), "\n")
	if len(lines) > 0 && lines[len(lines)-1] == "" {
		lines = lines[:len(lines)-1]
	}
	return lines
}
