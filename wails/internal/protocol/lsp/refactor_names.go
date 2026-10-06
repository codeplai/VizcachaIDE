package lsp

import (
	"strings"
	"unicode"

	"github.com/codeplai/VizcachaIDE/wails/internal/domain"
)

// keywords are never names to rename. Servers disagree: gopls offers to rename the "func" of a
// declaration (it answers a range and the placeholder "func()"), so the editor checks too.
var keywords = map[domain.CodeLanguage]map[string]bool{
	domain.CodeLanguageGo: wordSet("break case chan const continue default defer else fallthrough for func go goto if " +
		"import interface map package range return select struct switch type var"),
	domain.CodeLanguagePython: wordSet("False None True and as assert async await break class continue def del elif else " +
		"except finally for from global if import in is lambda nonlocal not or pass raise return try while with yield"),
	domain.CodeLanguageCpp: wordSet("alignas alignof and asm auto bool break case catch char class concept const consteval " +
		"constexpr constinit continue decltype default delete do double else enum explicit export extern false float for " +
		"friend goto if inline int long mutable namespace new noexcept not nullptr operator or private protected public " +
		"register requires return short signed sizeof static struct switch template this throw true try typedef typeid " +
		"typename union unsigned using virtual void volatile while"),
	domain.CodeLanguageRust: wordSet("as async await break const continue crate dyn else enum extern false fn for if impl " +
		"in let loop match mod move mut pub ref return self static struct super trait true type unsafe use where while"),
}

func wordSet(words string) map[string]bool {
	set := map[string]bool{}
	for _, word := range strings.Fields(words) {
		set[word] = true
	}
	return set
}

func isIdentifier(name string) bool {
	if name == "" {
		return false
	}
	for i, r := range name {
		if r != '_' && !unicode.IsLetter(r) && (i == 0 || !unicode.IsDigit(r)) {
			return false
		}
	}
	return true
}

// wordAround is the identifier that contains or touches the 1-based line and rune column.
func wordAround(text string, line, column int) (name string, start int, ok bool) {
	runes := []rune(lineOf(text, line-1))
	at := min(max(column-1, 0), len(runes))
	isPart := func(i int) bool {
		return i >= 0 && i < len(runes) && (runes[i] == '_' || unicode.IsLetter(runes[i]) || unicode.IsDigit(runes[i]))
	}
	if !isPart(at) && isPart(at-1) {
		at--
	}
	if !isPart(at) {
		return "", 0, false
	}
	from, to := at, at
	for isPart(from - 1) {
		from--
	}
	for isPart(to + 1) {
		to++
	}
	return string(runes[from : to+1]), from + 1, true
}

// vetRenameTarget refuses what is not a name (a keyword, an operator, a literal) whatever the
// server said, and fills the range and name when the server gave none.
func (s *Server) vetRenameTarget(doc document, at domain.SourceLocation, target domain.RenameTarget) domain.RenameTarget {
	if target.Refusal != domain.RenameAllowed {
		return target
	}
	name, start, found := wordAround(doc.text, at.Line, at.Column)
	line, endColumn := at.Line, 0
	if target.Range != nil {
		line, start = target.Range.Start.Line, target.Range.Start.Column
		endColumn = target.Range.End.Column
		name, found = "", target.Range.Start.Line == target.Range.End.Line
		if found {
			runes := []rune(lineOf(doc.text, line-1))
			if start >= 1 && endColumn-1 <= len(runes) && endColumn > start {
				name = string(runes[start-1 : endColumn-1])
			}
		}
	}
	if !found || !isIdentifier(name) || keywords[s.opts.CodeLanguage][name] {
		return domain.RenameTarget{Refusal: domain.RenameNotRenameable}
	}
	if target.Range == nil {
		span := domain.SourceRange{
			Start: domain.SourceLocation{File: at.File, Line: line, Column: start},
			End:   domain.SourceLocation{File: at.File, Line: line, Column: start + len([]rune(name))},
		}
		target.Range = &span
	}
	if !isIdentifier(target.Placeholder) {
		target.Placeholder = name
	}
	return target
}
