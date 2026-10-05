package shellterm

import (
	"os"
	"strings"
)

// Environment composes what a shell starts with: the user's environment, the IDE's folders
// first in PATH (the user's own PATH stays after them) and the variables a terminal emulator
// promises (xterm-256color, truecolor). The other entries keep their order.
func Environment(base, folders []string) []string {
	pathName, userPath := "PATH", ""
	env := make([]string, 0, len(base)+3)
	for _, entry := range base {
		name, value := splitEntry(entry)
		switch {
		case strings.EqualFold(name, "PATH"):
			pathName, userPath = name, value
		case name == "TERM" || name == "COLORTERM":
		default:
			env = append(env, entry)
		}
	}
	parts := append(append([]string{}, folders...), userPath)
	joined := strings.Join(nonEmpty(parts), string(os.PathListSeparator))
	return append(env, pathName+"="+joined, "TERM=xterm-256color", "COLORTERM=truecolor")
}

// splitEntry splits "NAME=value". Windows keeps hidden entries such as "=C:=C:\dir", whose
// name starts with "=".
func splitEntry(entry string) (name, value string) {
	skip := min(1, len(entry))
	name, value, _ = strings.Cut(entry[skip:], "=")
	return entry[:skip] + name, value
}

func nonEmpty(values []string) []string {
	kept := values[:0:0]
	for _, value := range values {
		if value != "" {
			kept = append(kept, value)
		}
	}
	return kept
}

// WithVariables returns env with the "NAME=value" entries set, replacing the variables of the
// same name (whatever its case: Windows is case-insensitive).
func WithVariables(env, variables []string) []string {
	if len(variables) == 0 {
		return env
	}
	replaced := map[string]bool{}
	for _, entry := range variables {
		name, _ := splitEntry(entry)
		replaced[strings.ToUpper(name)] = true
	}
	result := make([]string, 0, len(env)+len(variables))
	for _, entry := range env {
		if name, _ := splitEntry(entry); !replaced[strings.ToUpper(name)] {
			result = append(result, entry)
		}
	}
	return append(result, variables...)
}
