package vcpkg

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"sync"
)

// port is what the search keeps of one ports/<name>/vcpkg.json.
type port struct {
	Name        string `json:"name"`
	Version     string `json:"version"`
	Description string `json:"description"`
	Homepage    string `json:"homepage"`
}

// manifestOfPort is the part of a port's vcpkg.json the index reads. A description may be a
// string or a list of lines; the version has four spellings.
type manifestOfPort struct {
	Name          string          `json:"name"`
	Version       string          `json:"version"`
	VersionSemver string          `json:"version-semver"`
	VersionDate   string          `json:"version-date"`
	VersionString string          `json:"version-string"`
	Description   json.RawMessage `json:"description"`
	Homepage      string          `json:"homepage"`
	Supports      string          `json:"supports"`
}

// readPort reads one port; ok is false when the file is unreadable or the port does not support
// this platform.
func readPort(file string, platform platform) (port, bool) {
	data, err := os.ReadFile(file)
	if err != nil {
		return port{}, false
	}
	var manifest manifestOfPort
	if json.Unmarshal(data, &manifest) != nil || manifest.Name == "" {
		return port{}, false
	}
	if manifest.Supports != "" && !platform.supports(manifest.Supports) {
		return port{}, false
	}
	return port{
		Name:        manifest.Name,
		Version:     firstNonEmpty(manifest.Version, manifest.VersionSemver, manifest.VersionDate, manifest.VersionString),
		Description: descriptionText(manifest.Description),
		Homepage:    manifest.Homepage,
	}, true
}

func firstNonEmpty(values ...string) string {
	for _, value := range values {
		if value != "" {
			return value
		}
	}
	return ""
}

func descriptionText(raw json.RawMessage) string {
	var text string
	if json.Unmarshal(raw, &text) == nil {
		return strings.TrimSpace(text)
	}
	var lines []string
	if json.Unmarshal(raw, &lines) == nil {
		return strings.TrimSpace(strings.Join(lines, " "))
	}
	return ""
}

// portIndex is the list of ports of one vcpkg root, built once.
type portIndex struct {
	mu       sync.Mutex
	root     string
	platform platform
	cacheDir string // where the index is kept between runs ("" = memory only)
	ports    []port
}

// load returns the ports of root, reading them the first time (or when the root changed).
func (i *portIndex) load(root string) ([]port, error) {
	i.mu.Lock()
	defer i.mu.Unlock()
	if i.ports != nil && i.root == root {
		return i.ports, nil
	}
	cache := i.cacheFile(root)
	if ports, ok := readCache(cache); ok {
		i.root, i.ports = root, ports
		return ports, nil
	}
	ports, err := i.scan(root)
	if err != nil {
		return nil, err
	}
	i.root, i.ports = root, ports
	writeCache(cache, ports)
	return ports, nil
}

// scan reads every port concurrently: there are thousands of small files.
func (i *portIndex) scan(root string) ([]port, error) {
	entries, err := os.ReadDir(filepath.Join(root, "ports"))
	if err != nil {
		return nil, err
	}
	results := make([]*port, len(entries))
	work := make(chan int)
	var group sync.WaitGroup
	for range runtime.NumCPU() * 2 {
		group.Go(func() {
			for index := range work {
				file := filepath.Join(root, "ports", entries[index].Name(), "vcpkg.json")
				if found, ok := readPort(file, i.platform); ok {
					results[index] = &found
				}
			}
		})
	}
	for index, entry := range entries {
		if entry.IsDir() {
			work <- index
		}
	}
	close(work)
	group.Wait()
	ports := make([]port, 0, len(entries))
	for _, found := range results {
		if found != nil {
			ports = append(ports, *found)
		}
	}
	return ports, nil
}

// cacheFile names the cached index of a root. The name carries the modification time of the
// ports folder, so a new snapshot of vcpkg does not read an old index.
func (i *portIndex) cacheFile(root string) string {
	if i.cacheDir == "" {
		return ""
	}
	stamp := root
	if info, err := os.Stat(filepath.Join(root, "ports")); err == nil {
		stamp += info.ModTime().UTC().String()
	}
	stamp += i.platform.key()
	sum := sha256.Sum256([]byte(stamp))
	return filepath.Join(i.cacheDir, "index-"+hex.EncodeToString(sum[:8])+".json")
}

func readCache(file string) ([]port, bool) {
	if file == "" {
		return nil, false
	}
	data, err := os.ReadFile(file)
	if err != nil {
		return nil, false
	}
	var ports []port
	if json.Unmarshal(data, &ports) != nil || len(ports) == 0 {
		return nil, false
	}
	return ports, true
}

// writeCache is best effort: with no cache the next run scans again.
func writeCache(file string, ports []port) {
	if file == "" {
		return
	}
	data, err := json.Marshal(ports)
	if err != nil || os.MkdirAll(filepath.Dir(file), 0o755) != nil {
		return
	}
	temporary := file + ".part"
	if os.WriteFile(temporary, data, 0o644) == nil {
		_ = os.Rename(temporary, file)
	}
}
