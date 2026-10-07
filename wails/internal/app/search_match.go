package app

import (
	"errors"
	"fmt"
	"regexp"
	"strconv"
	"strings"
	"unicode"
	"unicode/utf16"
	"unicode/utf8"

	"github.com/codeplai/VizcachaIDE/wails/internal/domain"
)

// ErrInvalidPattern means the regular expression of the search cannot be compiled.
var ErrInvalidPattern = errors.New("search.invalidPattern")

// matcher finds the occurrences of a search in one line of text.
type matcher struct {
	re        *regexp.Regexp
	wholeWord bool
	regex     bool
}

func newMatcher(query string, options domain.SearchOptions) (*matcher, error) {
	pattern := query
	if !options.Regex {
		pattern = regexp.QuoteMeta(query)
	}
	if !options.CaseSensitive {
		pattern = "(?i)" + pattern
	}
	re, err := regexp.Compile(pattern)
	if err != nil {
		return nil, fmt.Errorf("%w: %w", ErrInvalidPattern, err)
	}
	return &matcher{re: re, wholeWord: options.WholeWord, regex: options.Regex}, nil
}

func isWordRune(r rune) bool { return r == '_' || unicode.IsLetter(r) || unicode.IsDigit(r) }

// find returns the submatch indexes of every non-empty occurrence in line.
func (m *matcher) find(line string) [][]int {
	var found [][]int
	for _, index := range m.re.FindAllStringSubmatchIndex(line, -1) {
		if index[0] == index[1] || (m.wholeWord && !m.isWholeWord(line, index[0], index[1])) {
			continue
		}
		found = append(found, index)
	}
	return found
}

func (m *matcher) isWholeWord(line string, start, end int) bool {
	before, _ := utf8.DecodeLastRuneInString(line[:start])
	after, _ := utf8.DecodeRuneInString(line[end:])
	wordBefore := start > 0 && isWordRune(before)
	wordAfter := end < len(line) && isWordRune(after)
	return !wordBefore && !wordAfter
}

// replacement is the text that takes the place of the occurrence. A literal search inserts the
// replacement as it is; a regex one expands $1..$99, $& (the match) and $$ (a dollar), like the
// editor side does.
func (m *matcher) replacement(template, line string, index []int) string {
	if !m.regex || !strings.Contains(template, "$") {
		return template
	}
	groups := len(index)/2 - 1
	var out strings.Builder
	for i := 0; i < len(template); i++ {
		next := byte(0)
		if i+1 < len(template) {
			next = template[i+1]
		}
		switch {
		case template[i] != '$':
			out.WriteByte(template[i])
		case next == '$':
			out.WriteByte('$')
			i++
		case next == '&':
			out.WriteString(line[index[0]:index[1]])
			i++
		case next >= '0' && next <= '9':
			group, used := groupNumber(template[i+1:], groups)
			if group == 0 {
				out.WriteByte('$')
				continue
			}
			if index[2*group] >= 0 {
				out.WriteString(line[index[2*group]:index[2*group+1]])
			}
			i += used
		default:
			out.WriteByte('$')
		}
	}
	return out.String()
}

// groupNumber reads one or two digits that name an existing group; 0 means "not a group".
func groupNumber(digits string, groups int) (group, used int) {
	if len(digits) >= 2 && digits[1] >= '0' && digits[1] <= '9' {
		if two, _ := strconv.Atoi(digits[:2]); two >= 1 && two <= groups {
			return two, 2
		}
	}
	if one := int(digits[0] - '0'); one >= 1 && one <= groups {
		return one, 1
	}
	return 0, 0
}

// utf16Length counts the UTF-16 units of text, the way the editor counts columns.
func utf16Length(text string) int {
	count := 0
	for _, r := range text {
		count += utf16.RuneLen(r)
	}
	return count
}
