package errorcatalog

import (
	"path/filepath"
	"regexp"
	"strconv"
	"strings"

	"github.com/codeplai/VizcachaIDE/wails/internal/domain"
)

var windowsAbsolute = regexp.MustCompile(`^[A-Za-z]:[\\/]`)

// resolveGoPath keeps absolute paths (Windows or POSIX) as Go printed them and joins
// relative ones ("./main.go", ".\util\calc.go") to the working directory.
func resolveGoPath(rawPath, workingDir string) string {
	if windowsAbsolute.MatchString(rawPath) || strings.HasPrefix(rawPath, "/") || strings.HasPrefix(rawPath, `\`) {
		return rawPath
	}
	relative := filepath.FromSlash(strings.ReplaceAll(rawPath, `\`, "/"))
	if workingDir == "" {
		return filepath.Clean(relative)
	}
	return filepath.Join(workingDir, relative)
}

// locationFrom builds a SourceLocation from the groups path, line and (optional) column.
func locationFrom(groups map[string]string, workingDir string) domain.SourceLocation {
	line, _ := strconv.Atoi(groups["line"])
	column := 1
	if parsed, err := strconv.Atoi(groups["column"]); err == nil {
		column = parsed
	}
	return domain.SourceLocation{File: resolveGoPath(groups["path"], workingDir), Line: line, Column: column}
}
