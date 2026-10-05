package packageindex

import (
	"context"
	"encoding/json"
	"net/http"
	"net/url"
	"strings"

	"github.com/codeplai/VizcachaIDE/wails/internal/domain"
)

// PyPI searches PyPI. It has no search API, so the answer joins two sources: the package whose
// name is exactly the query (the JSON API, reliable) first, then the popular projects whose name
// contains it (pypi_top.go), or the results of the search page (HTML, today blocked for programs)
// when that list is not available. If only the exact name is found, only that one is returned;
// if nothing works the error wraps app.ErrPackageIndexUnavailable.
func (i *Index) PyPI(ctx context.Context, query string) ([]domain.PackageInfo, error) {
	query, ok := clean(query)
	if !ok {
		return nil, nil
	}
	exactAnswer := make(chan *domain.PackageInfo, 1)
	go func() { exactAnswer <- i.pypiExact(ctx, query) }()
	listed, err := i.pypiTop(ctx, query)
	if err != nil || len(listed) == 0 {
		listed, err = i.pypiSearch(ctx, query) // the search page, in case it is readable again
	}
	exact := <-exactAnswer
	if err != nil && exact == nil {
		return nil, err
	}
	return mergeExact(exact, listed), nil
}

// pypiExact asks for the package named exactly like the query. Any failure means "none".
func (i *Index) pypiExact(ctx context.Context, query string) *domain.PackageInfo {
	if strings.ContainsAny(query, " \t/\\?#") {
		return nil
	}
	body, status, err := i.get(ctx, i.endpoints.PyPIJSON+url.PathEscape(query)+"/json", "application/json")
	if err != nil || status != http.StatusOK {
		return nil
	}
	var answer struct {
		Info struct {
			Name    string `json:"name"`
			Version string `json:"version"`
			Summary string `json:"summary"`
		} `json:"info"`
	}
	if json.Unmarshal(body, &answer) != nil || answer.Info.Name == "" {
		return nil
	}
	return &domain.PackageInfo{
		Name:        answer.Info.Name,
		Version:     answer.Info.Version,
		Description: oneLine(answer.Info.Summary),
		URL:         "https://pypi.org/project/" + url.PathEscape(answer.Info.Name) + "/",
	}
}

// pypiSearch reads the search page.
func (i *Index) pypiSearch(ctx context.Context, query string) ([]domain.PackageInfo, error) {
	address := i.endpoints.PyPISearch + "?q=" + url.QueryEscape(query)
	body, err := i.getOK(ctx, address, "text/html")
	if err != nil {
		return nil, err
	}
	return parsePyPIPage(body)
}

// mergeExact puts the exact match first and drops its duplicate from the listed results. PyPI
// treats "-", "_" and "." as the same, so names are compared normalised.
func mergeExact(exact *domain.PackageInfo, listed []domain.PackageInfo) []domain.PackageInfo {
	merged := make([]domain.PackageInfo, 0, MaxResults)
	if exact != nil {
		merged = append(merged, *exact)
	}
	for _, info := range listed {
		if len(merged) == MaxResults {
			break
		}
		if exact != nil && normalised(info.Name) == normalised(exact.Name) {
			continue
		}
		merged = append(merged, info)
	}
	return merged
}

func normalised(name string) string {
	return strings.NewReplacer("_", "-", ".", "-").Replace(strings.ToLower(name))
}
