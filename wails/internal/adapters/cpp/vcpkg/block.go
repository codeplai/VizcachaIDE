package vcpkg

import (
	"errors"
	"os"
	"path/filepath"
	"strings"
)

// The markers of the block of CMakeLists.txt that the IDE owns (docs/PLAN_CPP_CMAKE.md 4.0).
const (
	LibrariesBegin = "# VizcachaIDE libraries (vcpkg) · begin"
	LibrariesEnd   = "# VizcachaIDE libraries (vcpkg) · end"

	// CMakeFile is the project's CMake script.
	CMakeFile = "CMakeLists.txt"
)

// ErrNoCMakeLists means the project folder has no CMakeLists.txt.
var ErrNoCMakeLists = errors.New("the project has no CMakeLists.txt")

// group is the CMake lines of one library, written under "# <port>". Lines before the first
// header (the student's own) have the empty name and are kept as they are.
type group struct {
	port  string
	lines []string
}

// libraries is the text of CMakeLists.txt around its block.
type libraries struct {
	before, after []string
	groups        []group
	eol           string
	hasBlock      bool // the file already had the markers
}

func readLibraries(dir string) (*libraries, error) {
	data, err := os.ReadFile(filepath.Join(dir, CMakeFile))
	if errors.Is(err, os.ErrNotExist) {
		return nil, ErrNoCMakeLists
	}
	if err != nil {
		return nil, err
	}
	text := string(data)
	result := &libraries{eol: "\n"}
	if strings.Contains(text, "\r\n") {
		result.eol = "\r\n"
	}
	lines := strings.Split(strings.ReplaceAll(text, "\r\n", "\n"), "\n")
	if len(lines) > 0 && lines[len(lines)-1] == "" {
		lines = lines[:len(lines)-1]
	}
	begin, end := markerLines(lines)
	if begin < 0 {
		result.before = lines
		return result, nil
	}
	result.hasBlock = true
	result.before, result.after = lines[:begin], lines[end+1:]
	result.groups = parseGroups(lines[begin+1 : end])
	return result, nil
}

// markerLines returns the index of the begin and end markers, or -1 when the block is missing or
// incomplete.
func markerLines(lines []string) (begin, end int) {
	begin = -1
	for index, line := range lines {
		switch strings.TrimSpace(line) {
		case LibrariesBegin:
			begin = index
		case LibrariesEnd:
			if begin >= 0 {
				return begin, index
			}
		}
	}
	return -1, -1
}

func parseGroups(lines []string) []group {
	var groups []group
	for _, line := range lines {
		trimmed := strings.TrimSpace(line)
		switch {
		case strings.HasPrefix(trimmed, "# "):
			groups = append(groups, group{port: strings.TrimSpace(trimmed[2:])})
		case trimmed == "":
		default:
			if len(groups) == 0 {
				groups = append(groups, group{})
			}
			last := &groups[len(groups)-1]
			last.lines = append(last.lines, line)
		}
	}
	return groups
}

func (l *libraries) save(dir string) error {
	lines := append([]string{}, l.before...)
	lines = append(lines, LibrariesBegin)
	for _, item := range l.groups {
		if item.port != "" {
			lines = append(lines, "# "+item.port)
		}
		lines = append(lines, item.lines...)
	}
	lines = append(lines, LibrariesEnd)
	lines = append(lines, l.after...)
	text := strings.Join(lines, l.eol) + l.eol
	return os.WriteFile(filepath.Join(dir, CMakeFile), []byte(text), 0o644)
}

// SetLibraryLines writes the CMake lines of a port in the block of dir's CMakeLists.txt, replacing
// the ones it had. A CMakeLists.txt without the block gets it at the end.
func SetLibraryLines(dir, port string, lines []string) error {
	loaded, err := readLibraries(dir)
	if err != nil {
		return err
	}
	loaded.groups = withoutPort(loaded.groups, port)
	loaded.groups = append(loaded.groups, group{port: port, lines: lines})
	return loaded.save(dir)
}

// RemoveLibraryLines deletes the lines of a port from the block. It does nothing when the
// CMakeLists.txt has no block.
func RemoveLibraryLines(dir, port string) error {
	loaded, err := readLibraries(dir)
	if err != nil {
		return err
	}
	if !loaded.hasBlock {
		return nil
	}
	loaded.groups = withoutPort(loaded.groups, port)
	return loaded.save(dir)
}

func withoutPort(groups []group, port string) []group {
	kept := make([]group, 0, len(groups))
	for _, item := range groups {
		if item.port != port {
			kept = append(kept, item)
		}
	}
	return kept
}
