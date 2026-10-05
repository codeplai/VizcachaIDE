package vcpkg

import (
	"os"
	"path/filepath"
	"regexp"
	"strings"
)

var (
	addExecutable = regexp.MustCompile(`(?im)^\s*add_executable\(\s*([^\s)$]+)`)
	projectName   = regexp.MustCompile(`(?im)^\s*project\(\s*([^\s)$]+)`)
)

// targetName is the executable the libraries are linked to: the first add_executable of the
// project's CMakeLists.txt, else its project name, else "main".
func targetName(root string) string {
	data, err := os.ReadFile(filepath.Join(root, CMakeFile))
	if err != nil {
		return "main"
	}
	for _, pattern := range []*regexp.Regexp{addExecutable, projectName} {
		if found := pattern.FindSubmatch(data); found != nil {
			return string(found[1])
		}
	}
	return "main"
}

// installedVersions reads the versions of the libraries vcpkg installed in the project, from the
// dpkg style file build/vcpkg_installed/vcpkg/status (a stanza per package). Ports not installed
// yet are absent.
func installedVersions(root string) map[string]string {
	versions := map[string]string{}
	data, err := os.ReadFile(filepath.Join(root, "build", "vcpkg_installed", "vcpkg", "status"))
	if err != nil {
		return versions
	}
	for _, stanza := range strings.Split(strings.ReplaceAll(string(data), "\r\n", "\n"), "\n\n") {
		fields := stanzaFields(stanza)
		if fields["Package"] == "" || fields["Feature"] != "" || fields["Status"] != "install ok installed" {
			continue
		}
		versions[fields["Package"]] = fields["Version"]
	}
	return versions
}

func stanzaFields(stanza string) map[string]string {
	fields := map[string]string{}
	for _, line := range strings.Split(stanza, "\n") {
		if name, value, found := strings.Cut(line, ": "); found && !strings.HasPrefix(line, " ") {
			fields[name] = strings.TrimSpace(value)
		}
	}
	return fields
}
