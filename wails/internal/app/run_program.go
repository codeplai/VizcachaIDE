package app

import (
	"errors"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/codeplai/VizcachaIDE/wails/internal/domain"
)

// GoModFileName is the file that marks the root of a Go module.
const GoModFileName = "go.mod"

// ErrUnclosedQuote means the program arguments have a quote that is never closed.
var ErrUnclosedQuote = errors.New("the program arguments have an unclosed quote")

var moduleDirective = regexp.MustCompile(`(?m)^\s*module\s+"?([^"\s]+)"?`)

// ParseModulePath returns the module path declared in the text of a go.mod ("" if none).
func ParseModulePath(goModText string) string {
	match := moduleDirective.FindStringSubmatch(goModText)
	if match == nil {
		return ""
	}
	return match[1]
}

// FindGoModule returns the nearest module in startDir or any of its parents, or nil.
func FindGoModule(startDir string) *domain.GoModule {
	dir := startDir
	for {
		info, err := os.Stat(filepath.Join(dir, GoModFileName))
		if err == nil && !info.IsDir() {
			text, _ := os.ReadFile(filepath.Join(dir, GoModFileName))
			return &domain.GoModule{Root: dir, ModulePath: ParseModulePath(string(text))}
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			return nil
		}
		dir = parent
	}
}

// ConfigurationForFile decides how to run the active file: "go run ." in its folder
// when it lives inside a module, "go run file.go" otherwise.
func ConfigurationForFile(path string, programArgs []string) domain.RunConfiguration {
	if absolute, err := filepath.Abs(path); err == nil {
		path = absolute
	}
	config := domain.NewFileRunConfiguration(path, programArgs)
	module := FindGoModule(config.WorkingDir)
	if module == nil {
		return config
	}
	config.Mode = domain.RunPackage
	config.Module = module
	return config
}

// SplitProgramArguments splits text like a shell: spaces separate arguments and
// single or double quotes group them. Backslashes are literal so Windows paths work.
func SplitProgramArguments(text string) ([]string, error) {
	args := []string{}
	var current strings.Builder
	var quote rune
	inArgument := false
	for _, r := range text {
		switch {
		case quote != 0 && r == quote:
			quote = 0
		case quote != 0:
			current.WriteRune(r)
		case r == '"' || r == '\'':
			quote = r
			inArgument = true
		case r == ' ' || r == '\t' || r == '\n' || r == '\r':
			if inArgument {
				args = append(args, current.String())
				current.Reset()
				inArgument = false
			}
		default:
			current.WriteRune(r)
			inArgument = true
		}
	}
	if quote != 0 {
		return nil, ErrUnclosedQuote
	}
	if inArgument {
		args = append(args, current.String())
	}
	return args, nil
}
