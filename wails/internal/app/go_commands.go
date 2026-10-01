package app

import (
	"errors"
	"fmt"
	"strings"
	"unicode"
)

// ErrInvalidGoArgument means a module path or package name cannot be passed to the go command.
var ErrInvalidGoArgument = errors.New("invalid go command argument")

// ModInitArguments returns the arguments of "go mod init <modulePath>".
func ModInitArguments(modulePath string) ([]string, error) {
	name, err := singleWord(modulePath, "module path")
	if err != nil {
		return nil, err
	}
	return []string{"mod", "init", name}, nil
}

// GetArguments returns the arguments of "go get <pkg>".
func GetArguments(pkg string) ([]string, error) {
	name, err := singleWord(pkg, "package")
	if err != nil {
		return nil, err
	}
	return []string{"get", name}, nil
}

// ModTidyArguments returns the arguments of "go mod tidy".
func ModTidyArguments() []string { return []string{"mod", "tidy"} }

// singleWord accepts one non-empty word that cannot be mistaken for a flag.
func singleWord(value, what string) (string, error) {
	value = strings.TrimSpace(value)
	if value == "" {
		return "", fmt.Errorf("%w: the %s is empty", ErrInvalidGoArgument, what)
	}
	if strings.HasPrefix(value, "-") || strings.IndexFunc(value, unicode.IsSpace) >= 0 {
		return "", fmt.Errorf("%w: %q is not a valid %s", ErrInvalidGoArgument, value, what)
	}
	return value, nil
}
