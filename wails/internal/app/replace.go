package app

import (
	"strings"

	"github.com/codeplai/VizcachaIDE/wails/internal/domain"
)

// ReplaceRequest says what to replace in which files of the folder. With Line > 0 only the
// occurrence that starts at that line and column is replaced (it is looked up again, so a file
// that changed since the search is not damaged).
type ReplaceRequest struct {
	Root        string
	Query       string
	Options     domain.SearchOptions
	Replacement string
	Paths       []string
	Line        int
	Column      int
}

// ReplaceInFolder rewrites the listed files of the folder on disk. Paths outside the folder,
// in skipped folders, or that are not text files are ignored. remember (may be nil) is told the
// new text of each file written.
func ReplaceInFolder(request ReplaceRequest, remember func(path, text string)) (domain.ReplaceResult, error) {
	var result domain.ReplaceResult
	if request.Query == "" {
		return result, nil
	}
	m, err := newMatcher(request.Query, request.Options)
	if err != nil {
		return result, err
	}
	for _, path := range request.Paths {
		if !insideFolder(request.Root, path) {
			continue
		}
		file, ok := readTextFile(path)
		if !ok {
			continue
		}
		text, count := replaceText(m, file.text, request)
		if count == 0 {
			continue
		}
		if file.bom {
			text = utf8BOM + text
		}
		if err := WriteSourceFile(path, text); err != nil {
			return result, err
		}
		if remember != nil {
			remember(path, text)
		}
		result.Files++
		result.Matches += count
	}
	return result, nil
}

// replaceText applies the request to the text of one file and counts the replacements.
func replaceText(m *matcher, text string, request ReplaceRequest) (string, int) {
	lines := strings.Split(text, "\n")
	count := 0
	for i, line := range lines {
		if request.Line > 0 && request.Line != i+1 {
			continue
		}
		body := strings.TrimSuffix(line, "\r")
		rewritten, n := replaceLine(m, body, request)
		if n > 0 {
			lines[i] = rewritten + line[len(body):]
			count += n
		}
	}
	return strings.Join(lines, "\n"), count
}

func replaceLine(m *matcher, body string, request ReplaceRequest) (string, int) {
	var out strings.Builder
	last, count := 0, 0
	for _, index := range m.find(body) {
		if request.Line > 0 && utf16Length(body[:index[0]])+1 != request.Column {
			continue
		}
		out.WriteString(body[last:index[0]])
		out.WriteString(m.replacement(request.Replacement, body, index))
		last = index[1]
		count++
	}
	out.WriteString(body[last:])
	return out.String(), count
}
