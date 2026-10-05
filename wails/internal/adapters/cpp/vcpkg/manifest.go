package vcpkg

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"unicode"

	"golang.org/x/text/runes"
	"golang.org/x/text/transform"
	"golang.org/x/text/unicode/norm"
)

// ManifestFile is the name of the manifest vcpkg reads in manifest mode.
const ManifestFile = "vcpkg.json"

// ErrNoManifest means the project folder has no vcpkg.json.
var ErrNoManifest = errors.New("the project has no vcpkg.json")

// manifest is a vcpkg.json that keeps the order of its keys and the text of every value, so
// editing the dependencies leaves the rest as the student wrote it.
type manifest struct {
	keys   []string
	values map[string]json.RawMessage
}

func readManifest(path string) (*manifest, error) {
	data, err := os.ReadFile(path)
	if errors.Is(err, os.ErrNotExist) {
		return nil, ErrNoManifest
	}
	if err != nil {
		return nil, err
	}
	decoder := json.NewDecoder(bytes.NewReader(data))
	if token, err := decoder.Token(); err != nil || token != json.Delim('{') {
		return nil, fmt.Errorf("vcpkg.json is not a JSON object")
	}
	result := &manifest{values: map[string]json.RawMessage{}}
	for decoder.More() {
		key, err := decoder.Token()
		if err != nil {
			return nil, fmt.Errorf("vcpkg.json: %w", err)
		}
		var value json.RawMessage
		if err := decoder.Decode(&value); err != nil {
			return nil, fmt.Errorf("vcpkg.json: %w", err)
		}
		name, _ := key.(string)
		result.keys = append(result.keys, name)
		result.values[name] = value
	}
	return result, nil
}

// save writes the manifest with two spaces of indentation and a final newline.
func (m *manifest) save(path string) error {
	var out bytes.Buffer
	out.WriteString("{\n")
	for index, key := range m.keys {
		name, _ := json.Marshal(key)
		out.WriteString("  " + string(name) + ": ")
		if err := json.Indent(&out, m.values[key], "  ", "  "); err != nil {
			return err
		}
		if index < len(m.keys)-1 {
			out.WriteString(",")
		}
		out.WriteString("\n")
	}
	out.WriteString("}\n")
	return os.WriteFile(path, out.Bytes(), 0o644)
}

// dependencyList returns the raw entries of "dependencies" (strings or {"name": ...} objects).
func (m *manifest) dependencyList() []json.RawMessage {
	var list []json.RawMessage
	_ = json.Unmarshal(m.values["dependencies"], &list)
	return list
}

func (m *manifest) setDependencies(list []json.RawMessage) {
	if list == nil {
		list = []json.RawMessage{}
	}
	data, _ := json.Marshal(list)
	if _, present := m.values["dependencies"]; !present {
		m.keys = append(m.keys, "dependencies")
	}
	m.values["dependencies"] = data
}

// dependencyName returns the port an entry names.
func dependencyName(entry json.RawMessage) string {
	var name string
	if json.Unmarshal(entry, &name) == nil {
		return name
	}
	var object struct {
		Name string `json:"name"`
	}
	if json.Unmarshal(entry, &object) == nil {
		return object.Name
	}
	return ""
}

// Dependencies returns the ports the project of dir declares in its vcpkg.json.
func Dependencies(dir string) ([]string, error) {
	loaded, err := readManifest(filepath.Join(dir, ManifestFile))
	if err != nil {
		return nil, err
	}
	names := []string{}
	for _, entry := range loaded.dependencyList() {
		if name := dependencyName(entry); name != "" {
			names = append(names, name)
		}
	}
	return names, nil
}

// AddDependency adds a port to the vcpkg.json of dir (creating the file when the folder has
// none). It reports false when the port was already there.
func AddDependency(dir, port string) (bool, error) {
	path := filepath.Join(dir, ManifestFile)
	loaded, err := readManifest(path)
	if errors.Is(err, ErrNoManifest) {
		loaded = newManifest(filepath.Base(dir))
		err = nil
	}
	if err != nil {
		return false, err
	}
	list := loaded.dependencyList()
	for _, entry := range list {
		if dependencyName(entry) == port {
			return false, nil
		}
	}
	name, _ := json.Marshal(port)
	loaded.setDependencies(append(list, name))
	return true, loaded.save(path)
}

// RemoveDependency removes a port from the vcpkg.json of dir. It reports false when it was not
// there.
func RemoveDependency(dir, port string) (bool, error) {
	path := filepath.Join(dir, ManifestFile)
	loaded, err := readManifest(path)
	if err != nil {
		return false, err
	}
	var kept []json.RawMessage
	for _, entry := range loaded.dependencyList() {
		if dependencyName(entry) != port {
			kept = append(kept, entry)
		}
	}
	if len(kept) == len(loaded.dependencyList()) {
		return false, nil
	}
	loaded.setDependencies(kept)
	return true, loaded.save(path)
}

var notManifestName = regexp.MustCompile(`[^a-z0-9]+`)

// newManifest is the vcpkg.json of a folder that had none: vcpkg names are lower case letters,
// digits and hyphens.
func newManifest(folder string) *manifest {
	name := strings.Trim(notManifestName.ReplaceAllString(strings.ToLower(withoutAccents(folder)), "-"), "-")
	if name == "" {
		name = "app"
	}
	quoted, _ := json.Marshal(name)
	return &manifest{
		keys:   []string{"name", "version"},
		values: map[string]json.RawMessage{"name": quoted, "version": json.RawMessage(`"0.1.0"`)},
	}
}

// withoutAccents turns "Ñandú" into "Nandu": vcpkg names are ASCII.
func withoutAccents(text string) string {
	plain, _, err := transform.String(transform.Chain(norm.NFD, runes.Remove(runes.In(unicode.Mn)), norm.NFC), text)
	if err != nil {
		return text
	}
	return plain
}
