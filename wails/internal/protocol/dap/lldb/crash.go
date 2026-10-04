package lldb

import "strings"

// crashPhrases name a crash the same way the runner does, so the Assistant recognises it. The
// patterns are lower case parts of what LLDB says on each system. Stack overflow goes first:
// it is also an access violation on some systems.
var crashPhrases = []struct {
	phrase   string
	patterns []string
}{
	{"Stack overflow", []string{"0xc00000fd"}},
	{"Segmentation fault", []string{"sigsegv", "sigbus", "exc_bad_access", "0xc0000005"}},
	{"Floating point exception", []string{"sigfpe", "exc_arithmetic", "0xc0000094", "0xc0000095"}},
	{"Aborted", []string{"sigabrt", "exc_crash", "0xc0000409"}},
}

// CrashPhrase returns the English phrase of the crash an LLDB description reports, or "".
func CrashPhrase(description string) string {
	lower := strings.ToLower(description)
	for _, crash := range crashPhrases {
		for _, pattern := range crash.patterns {
			if strings.Contains(lower, pattern) {
				return crash.phrase
			}
		}
	}
	return ""
}

// crashDescription is "Segmentation fault: signal SIGSEGV"; unknown signals keep LLDB's text.
func crashDescription(description string) string {
	phrase := CrashPhrase(description)
	switch {
	case phrase == "":
		return description
	case description == "":
		return phrase
	}
	return phrase + ": " + description
}
