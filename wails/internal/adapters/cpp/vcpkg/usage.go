package vcpkg

import (
	"bufio"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
)

// usageCommand matches the CMake commands a "usage" file may ask the project to write.
var usageCommand = regexp.MustCompile(`^(find_package|find_path|find_library|find_file|target_link_libraries|target_include_directories|target_compile_definitions|include_directories|pkg_check_modules|pkg_search_module|include|set)\(`)

// firstArgument matches the target name of "target_*(<target> ...)".
var firstArgument = regexp.MustCompile(`^(target_\w+\()\s*(?:<[^>]*>|[^\s)]+)`)

// InstalledDirectory is where vcpkg installs the libraries of the project of dir, for a triplet.
func InstalledDirectory(dir, triplet string) string {
	return filepath.Join(dir, "build", "vcpkg_installed", triplet)
}

// UsageLines returns the CMake lines that use a port once it is installed in the project of dir:
// the ones vcpkg prints in share/<port>/usage, with the project's target in place of its example
// target; or, for a library without usage text, a find_path of one of its headers. It returns
// nil when the port left nothing to write.
func UsageLines(dir, triplet, port, target string) []string {
	installed := InstalledDirectory(dir, triplet)
	if data, err := os.ReadFile(filepath.Join(installed, "share", port, "usage")); err == nil {
		if lines := parseUsage(string(data), target); len(lines) > 0 {
			return lines
		}
	}
	return headerOnlyLines(dir, triplet, port, target)
}

// parseUsage keeps the first run of CMake commands of a usage text. The text mixes prose,
// comments and alternatives ("# Or use the header-only version"); the first run is the way that
// always works.
func parseUsage(text, target string) []string {
	var lines []string
	var open string // the command being read when it spans several lines
	scanner := bufio.NewScanner(strings.NewReader(text))
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if open != "" {
			open += " " + line
			if balanced(open) {
				lines, open = append(lines, withTarget(open, target)), ""
			}
			continue
		}
		switch {
		case usageCommand.MatchString(line):
			if balanced(line) {
				lines = append(lines, withTarget(line, target))
			} else {
				open = line
			}
		case line == "" || len(lines) == 0:
		default:
			return lines
		}
	}
	return lines
}

func balanced(command string) bool {
	return strings.Count(command, "(") <= strings.Count(command, ")")
}

// withTarget puts the project's target in the first argument of a target_* command.
func withTarget(command, target string) string {
	return firstArgument.ReplaceAllString(command, "${1}"+target)
}

// headerOnlyLines is the fallback for a library without usage text: it looks for one of its
// headers in the list of files vcpkg installed for it.
func headerOnlyLines(dir, triplet, port, target string) []string {
	header := mainHeader(dir, triplet, port)
	if header == "" {
		return nil
	}
	variable := strings.ToUpper(nonIdentifier.ReplaceAllString(port, "_")) + "_INCLUDE_DIRS"
	return []string{
		`find_path(` + variable + ` "` + header + `")`,
		`target_include_directories(` + target + ` PRIVATE ${` + variable + `})`,
	}
}

var nonIdentifier = regexp.MustCompile(`[^A-Za-z0-9]`)

// mainHeader picks the header of a port with the shortest path (<port>.h before <port>/detail/x.h).
func mainHeader(dir, triplet, port string) string {
	lists, _ := filepath.Glob(filepath.Join(dir, "build", "vcpkg_installed", "vcpkg", "info", port+"_*_"+triplet+".list"))
	if len(lists) == 0 {
		return ""
	}
	data, err := os.ReadFile(lists[0])
	if err != nil {
		return ""
	}
	prefix := triplet + "/include/"
	var headers []string
	for _, line := range strings.Split(string(data), "\n") {
		line = strings.TrimSpace(line)
		if strings.HasPrefix(line, prefix) && isHeader(line) {
			headers = append(headers, strings.TrimPrefix(line, prefix))
		}
	}
	if len(headers) == 0 {
		return ""
	}
	sort.Slice(headers, func(a, b int) bool {
		if len(headers[a]) != len(headers[b]) {
			return len(headers[a]) < len(headers[b])
		}
		return headers[a] < headers[b]
	})
	return headers[0]
}

func isHeader(path string) bool {
	switch strings.ToLower(filepath.Ext(path)) {
	case ".h", ".hpp", ".hh", ".hxx":
		return true
	}
	return false
}
