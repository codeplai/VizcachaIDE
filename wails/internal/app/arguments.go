package app

import (
	"errors"
	"fmt"
	"strings"
	"unicode"
)

// ErrInvalidArgument means a name given to a tool (a module path, a package) cannot be passed
// to it: it is empty, has spaces or could be mistaken for a flag.
var ErrInvalidArgument = errors.New("invalid tool argument")

// SingleWordArgument accepts one non-empty word that cannot be mistaken for a flag. what names
// the value in the error ("module path", "package").
func SingleWordArgument(value, what string) (string, error) {
	value = strings.TrimSpace(value)
	if value == "" {
		return "", fmt.Errorf("%w: the %s is empty", ErrInvalidArgument, what)
	}
	if strings.HasPrefix(value, "-") || strings.IndexFunc(value, unicode.IsSpace) >= 0 {
		return "", fmt.Errorf("%w: %q is not a valid %s", ErrInvalidArgument, value, what)
	}
	return value, nil
}
