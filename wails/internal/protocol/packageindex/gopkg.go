package packageindex

import (
	"context"
	"encoding/json"
	"net/http"
	"net/url"
	"strings"

	"github.com/codeplai/VizcachaIDE/wails/internal/domain"
)

// Go searches pkg.go.dev. It has no search API, so the answer joins two sources: the module
// named exactly like the query, when the query looks like a module path (the module proxy,
// reliable), then the results of the search page (HTML, fragile). If the page cannot be read but
// the exact module exists, only that one is returned.
func (i *Index) Go(ctx context.Context, query string) ([]domain.PackageInfo, error) {
	query, ok := clean(query)
	if !ok {
		return nil, nil
	}
	exactAnswer := make(chan *domain.PackageInfo, 1)
	go func() { exactAnswer <- i.goExact(ctx, query) }()
	listed, err := i.goSearch(ctx, query)
	exact := <-exactAnswer
	if err != nil && exact == nil {
		return nil, err
	}
	return mergeGo(exact, listed), nil
}

func (i *Index) goSearch(ctx context.Context, query string) ([]domain.PackageInfo, error) {
	address := i.endpoints.GoSearch + "?m=package&q=" + url.QueryEscape(query)
	body, err := i.getOK(ctx, address, "text/html")
	if err != nil {
		return nil, err
	}
	return parseGoPage(body, i.endpoints.GoPackage)
}

// goExact asks the module proxy for the latest version of the module named like the query. A
// module path has a dot in its first element and a slash ("github.com/google/uuid"). Any
// failure means "none".
func (i *Index) goExact(ctx context.Context, query string) *domain.PackageInfo {
	first, _, hasSlash := strings.Cut(query, "/")
	if !hasSlash || !strings.Contains(first, ".") || strings.ContainsAny(query, " \t?#\\") {
		return nil
	}
	address := i.endpoints.GoProxy + escapeModule(query) + "/@latest"
	body, status, err := i.get(ctx, address, "application/json")
	if err != nil || status != http.StatusOK {
		return nil
	}
	var answer struct{ Version string }
	if json.Unmarshal(body, &answer) != nil || answer.Version == "" {
		return nil
	}
	return &domain.PackageInfo{Name: query, Version: answer.Version, URL: i.endpoints.GoPackage + query}
}

// escapeModule writes a module path the way the module proxy wants it: each capital letter
// becomes "!" and the lower case letter ("Azure" is "!azure").
func escapeModule(path string) string {
	var escaped strings.Builder
	for _, character := range path {
		if character >= 'A' && character <= 'Z' {
			escaped.WriteByte('!')
			character += 'a' - 'A'
		}
		escaped.WriteRune(character)
	}
	return escaped.String()
}

// mergeGo puts the exact module first. The exact one has no description, so when the page also
// lists it, the page's entry (with its synopsis) is kept in first place.
func mergeGo(exact *domain.PackageInfo, listed []domain.PackageInfo) []domain.PackageInfo {
	merged := make([]domain.PackageInfo, 0, MaxResults)
	if exact != nil {
		entry := *exact
		for _, info := range listed {
			if strings.EqualFold(info.Name, exact.Name) {
				entry.Description = info.Description
			}
		}
		merged = append(merged, entry)
	}
	for _, info := range listed {
		if len(merged) == MaxResults {
			break
		}
		if exact != nil && strings.EqualFold(info.Name, exact.Name) {
			continue
		}
		merged = append(merged, info)
	}
	return merged
}
