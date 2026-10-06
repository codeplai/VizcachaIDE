package vcpkg

import (
	"context"
	"fmt"
	"sort"
	"strings"

	"github.com/codeplai/VizcachaIDE/wails/internal/app"
	"github.com/codeplai/VizcachaIDE/wails/internal/domain"
	"github.com/codeplai/VizcachaIDE/wails/internal/protocol/packageindex"
)

// Search is the adapter of app.PackageSearch that looks libraries up in the ports of vcpkg. It
// needs no internet: the ports are the files of the vcpkg shipped with the IDE.
type Search struct {
	locator *Locator
	index   portIndex
}

var _ app.PackageSearch = (*Search)(nil)

// NewSearch creates the search. The index of ports is built on the first query and kept in
// <cache>/vcpkg between runs.
func NewSearch(locator *Locator) *Search {
	options := locator.options
	search := &Search{locator: locator}
	search.index.platform = platform{triplet: locator.Triplet(), goos: options.GOOS, goarch: options.GOARCH}
	search.index.cacheDir = locator.CacheRoot()
	return search
}

// Warm builds the index in the background, so the first search of the student finds it ready.
// Reading the 2.900 small files of the ports takes some seconds the first time (the index is then
// kept on disk and a later start reads one file); the application calls Warm when it opens a C++
// project.
func (s *Search) Warm() {
	if root := s.locator.Root(); root != "" {
		go func() { _, _ = s.index.load(root) }()
	}
}

// Search implements app.PackageSearch. Names that start with the query come first, then names
// that contain it, then descriptions that contain it; shorter names win inside each group.
func (s *Search) Search(ctx context.Context, query string) ([]domain.PackageInfo, error) {
	query = strings.ToLower(strings.TrimSpace(query))
	if query == "" {
		return []domain.PackageInfo{}, nil
	}
	root := s.locator.Root()
	if root == "" {
		return nil, fmt.Errorf("%w: vcpkg was not found", app.ErrPackageIndexUnavailable)
	}
	ports, err := s.index.load(root)
	if err != nil {
		return nil, fmt.Errorf("%w: %w", app.ErrPackageIndexUnavailable, err)
	}
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	return rank(ports, query), nil
}

// rank returns the best MaxResults ports for the query.
func rank(ports []port, query string) []domain.PackageInfo {
	type hit struct {
		port  port
		group int
	}
	var hits []hit
	for _, candidate := range ports {
		name := strings.ToLower(candidate.Name)
		switch {
		case name == query:
			hits = append(hits, hit{candidate, 0})
		case strings.HasPrefix(name, query):
			hits = append(hits, hit{candidate, 1})
		case strings.Contains(name, query):
			hits = append(hits, hit{candidate, 2})
		case strings.Contains(strings.ToLower(candidate.Description), query):
			hits = append(hits, hit{candidate, 3})
		}
	}
	sort.Slice(hits, func(a, b int) bool {
		if hits[a].group != hits[b].group {
			return hits[a].group < hits[b].group
		}
		if len(hits[a].port.Name) != len(hits[b].port.Name) {
			return len(hits[a].port.Name) < len(hits[b].port.Name)
		}
		return hits[a].port.Name < hits[b].port.Name
	})
	if len(hits) > packageindex.MaxResults {
		hits = hits[:packageindex.MaxResults]
	}
	found := make([]domain.PackageInfo, 0, len(hits))
	for _, item := range hits {
		found = append(found, domain.PackageInfo{
			Name: item.port.Name, Version: item.port.Version,
			Description: item.port.Description, URL: item.port.Homepage,
		})
	}
	return found
}
