package cmake

import (
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"unicode"

	"golang.org/x/text/runes"
	"golang.org/x/text/transform"
	"golang.org/x/text/unicode/norm"
)

// ListsFile is the file that makes a folder a CMake project.
const ListsFile = "CMakeLists.txt"

// Project is a CMake project: its top folder and the name of its main executable.
type Project struct {
	Root   string
	Target string // ASCII name of the main executable target
}

// reserved are the names CMake does not accept for a target.
var reserved = map[string]bool{
	"all": true, "clean": true, "help": true, "install": true, "package": true, "test": true,
	"edit_cache": true, "rebuild_cache": true, "depend": true, "tests": true, "uninstall": true,
}

// executableCall finds the first add_executable(name ...) with a literal name.
var executableCall = regexp.MustCompile(`(?im)^\s*add_executable\s*\(\s*([A-Za-z0-9_.+-]+)`)

// FindProject finds the CMake project around path (a file or a folder): the topmost folder with
// a CMakeLists.txt going up from it, never above stopAt (the folder the user opened). An empty
// stopAt goes up to the root of the disk; a stopAt that does not hold path is ignored and only
// the path's own folder is looked at.
func FindProject(path, stopAt string) (Project, bool) {
	dir := startFolder(path)
	if stopAt != "" && !contains(stopAt, dir) {
		stopAt = dir
	}
	top := ""
	for {
		if isFile(filepath.Join(dir, ListsFile)) {
			top = dir
		}
		parent := filepath.Dir(dir)
		if parent == dir || (stopAt != "" && sameFolder(dir, stopAt)) {
			break
		}
		dir = parent
	}
	if top == "" {
		return Project{}, false
	}
	return Project{Root: top, Target: targetOf(top)}, true
}

// startFolder is path when it is a folder, otherwise the folder of the file.
func startFolder(path string) string {
	abs, err := filepath.Abs(path)
	if err != nil {
		abs = path
	}
	if info, err := os.Stat(abs); err == nil && info.IsDir() {
		return abs
	}
	return filepath.Dir(abs)
}

// targetOf is the executable the project's CMakeLists.txt declares first, or the name its folder
// would give.
func targetOf(root string) string {
	if text, err := os.ReadFile(filepath.Join(root, ListsFile)); err == nil {
		if found := executableCall.FindSubmatch(text); found != nil {
			return string(found[1])
		}
	}
	return TargetName(filepath.Base(root))
}

// TargetName turns a folder name into the executable's target name: lower case ASCII letters,
// digits and "_", not starting with a digit and not a name CMake reserves ("Ñandú" gives
// "nandu", "Mi Programa" gives "mi_programa").
func TargetName(folderName string) string {
	var name strings.Builder
	for _, letter := range strings.ToLower(withoutAccents(folderName)) {
		if letter < unicode.MaxASCII && (unicode.IsLetter(letter) || unicode.IsDigit(letter)) {
			name.WriteRune(letter)
			continue
		}
		name.WriteRune('_')
	}
	clean := strings.Trim(name.String(), "_")
	for strings.Contains(clean, "__") {
		clean = strings.ReplaceAll(clean, "__", "_")
	}
	if clean == "" {
		return "app"
	}
	if unicode.IsDigit(rune(clean[0])) || reserved[clean] {
		return "app_" + clean
	}
	return clean
}

// withoutAccents turns "Ñandú" into "Nandu": the target name is ASCII but still recognisable.
func withoutAccents(text string) string {
	plain, _, err := transform.String(transform.Chain(norm.NFD, runes.Remove(runes.In(unicode.Mn)), norm.NFC), text)
	if err != nil {
		return text
	}
	return plain
}

func isFile(path string) bool {
	info, err := os.Stat(path)
	return err == nil && !info.IsDir()
}

// relation is where dir is compared with folder: "." for the same folder, otherwise the
// relative path (starting with ".." when it is outside). filepath.Rel ignores case on Windows.
func relation(folder, dir string) string {
	relative, err := filepath.Rel(filepath.Clean(folder), filepath.Clean(dir))
	if err != nil {
		return ".."
	}
	return relative
}

func sameFolder(a, b string) bool { return relation(a, b) == "." }

// contains reports whether dir is folder or inside it.
func contains(folder, dir string) bool {
	relative := relation(folder, dir)
	return relative != ".." && !strings.HasPrefix(relative, ".."+string(filepath.Separator))
}
