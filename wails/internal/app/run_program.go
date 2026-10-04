package app

import (
	"errors"
	"strings"
)

// ErrUnclosedQuote means the program arguments have a quote that is never closed.
var ErrUnclosedQuote = errors.New("the program arguments have an unclosed quote")

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
