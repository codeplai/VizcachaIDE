package packageindex

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/codeplai/VizcachaIDE/wails/internal/domain"
)

// PyPI's search page answers programs with a JavaScript challenge, so partial names are looked up
// in the list of the ~15 000 most downloaded PyPI projects (hugovk/top-pypi-packages, refreshed
// monthly from PyPI's public download data), kept for a week in memory and in the user's cache.
// What a beginner looks for is almost always there; anything else is still found by its exact name.

// topMaxAge is how long the downloaded list is used before asking for it again.
const topMaxAge = 7 * 24 * time.Hour

// topList is the list of popular projects, in download order.
type topList struct {
	mu      sync.Mutex
	names   []string
	fetched time.Time
}

// pypiTop finds the popular projects whose name contains the query and fills in their version and
// summary from the JSON API.
func (i *Index) pypiTop(ctx context.Context, query string) ([]domain.PackageInfo, error) {
	names, err := i.topNames(ctx)
	if err != nil {
		return nil, err
	}
	matched := matchNames(names, query)
	infos := make([]domain.PackageInfo, len(matched))
	var wait sync.WaitGroup
	for n, name := range matched {
		wait.Add(1)
		go func() {
			defer wait.Done()
			infos[n] = domain.PackageInfo{Name: name, URL: "https://pypi.org/project/" + name + "/"}
			if found := i.pypiExact(ctx, name); found != nil {
				infos[n] = *found
			}
		}()
	}
	wait.Wait()
	return infos, nil
}

// matchNames keeps the names that contain the query (compared normalised): the ones that start
// with it first, then the rest, each group in download order.
func matchNames(names []string, query string) []string {
	wanted := normalised(query)
	var starting, containing []string
	for _, name := range names {
		key := normalised(name)
		switch {
		case strings.HasPrefix(key, wanted):
			starting = append(starting, name)
		case strings.Contains(key, wanted):
			containing = append(containing, name)
		}
		if len(starting) >= MaxResults {
			break
		}
	}
	matched := append(starting, containing...)
	return matched[:min(len(matched), MaxResults)]
}

// topNames is the list, from memory, the disk cache or the network (in that order of freshness).
// A stale copy is better than nothing when the network is down.
func (i *Index) topNames(ctx context.Context) ([]string, error) {
	i.top.mu.Lock()
	defer i.top.mu.Unlock()
	if len(i.top.names) > 0 && time.Since(i.top.fetched) < topMaxAge {
		return i.top.names, nil
	}
	cached, cachedAt := i.readTopCache()
	if len(cached) > 0 && time.Since(cachedAt) < topMaxAge {
		i.top.names, i.top.fetched = cached, cachedAt
		return cached, nil
	}
	body, err := i.getOK(ctx, i.endpoints.PyPITop, "application/json")
	if err == nil {
		var names []string
		if names, err = parseTopList(body); err == nil {
			i.top.names, i.top.fetched = names, time.Now()
			i.writeTopCache(body)
			return names, nil
		}
	}
	if len(cached) > 0 {
		i.top.names, i.top.fetched = cached, cachedAt
		return cached, nil
	}
	return nil, err
}

// parseTopList reads {"rows": [{"project": "boto3", ...}, ...]}.
func parseTopList(body []byte) ([]string, error) {
	var list struct {
		Rows []struct {
			Project string `json:"project"`
		} `json:"rows"`
	}
	if err := json.Unmarshal(body, &list); err != nil {
		return nil, unavailable(fmt.Errorf("the list of popular PyPI projects is not readable: %w", err))
	}
	if len(list.Rows) == 0 {
		return nil, unavailable(errors.New("the list of popular PyPI projects is empty"))
	}
	names := make([]string, 0, len(list.Rows))
	for _, row := range list.Rows {
		if row.Project != "" {
			names = append(names, row.Project)
		}
	}
	return names, nil
}

func (i *Index) topCacheFile() string {
	if i.cacheDir == "" {
		return ""
	}
	return filepath.Join(i.cacheDir, "top-pypi-packages.json")
}

func (i *Index) readTopCache() ([]string, time.Time) {
	path := i.topCacheFile()
	info, err := os.Stat(path)
	if path == "" || err != nil {
		return nil, time.Time{}
	}
	body, err := os.ReadFile(path)
	if err != nil {
		return nil, time.Time{}
	}
	names, err := parseTopList(body)
	if err != nil {
		return nil, time.Time{}
	}
	return names, info.ModTime()
}

func (i *Index) writeTopCache(body []byte) {
	path := i.topCacheFile()
	if path == "" || os.MkdirAll(filepath.Dir(path), 0o755) != nil {
		return
	}
	_ = os.WriteFile(path, body, 0o600) // a cache that cannot be written is only slower
}
