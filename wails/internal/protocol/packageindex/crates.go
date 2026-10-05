package packageindex

import (
	"context"
	"encoding/json"
	"net/url"
	"strconv"

	"github.com/codeplai/VizcachaIDE/wails/internal/domain"
)

type cratesAnswer struct {
	Crates []struct {
		Name           string `json:"name"`
		MaxStable      string `json:"max_stable_version"`
		NewestVersion  string `json:"newest_version"`
		Description    string `json:"description"`
		DefaultVersion string `json:"default_version"`
	} `json:"crates"`
}

// Crates searches crates.io with its official API. The crate whose name equals the query comes
// first, as the API already orders it.
func (i *Index) Crates(ctx context.Context, query string) ([]domain.PackageInfo, error) {
	query, ok := clean(query)
	if !ok {
		return nil, nil
	}
	address := i.endpoints.Crates + "?per_page=" + strconv.Itoa(MaxResults) + "&q=" + url.QueryEscape(query)
	body, err := i.getOK(ctx, address, "application/json")
	if err != nil {
		return nil, err
	}
	return parseCrates(body)
}

func parseCrates(body []byte) ([]domain.PackageInfo, error) {
	var answer cratesAnswer
	if err := json.Unmarshal(body, &answer); err != nil {
		return nil, unavailable(err)
	}
	found := make([]domain.PackageInfo, 0, len(answer.Crates))
	for _, crate := range answer.Crates {
		if crate.Name == "" || len(found) == MaxResults {
			continue
		}
		found = append(found, domain.PackageInfo{
			Name:        crate.Name,
			Version:     firstNonEmpty(crate.MaxStable, crate.NewestVersion, crate.DefaultVersion),
			Description: oneLine(crate.Description),
			URL:         "https://crates.io/crates/" + url.PathEscape(crate.Name),
		})
	}
	return found, nil
}

func firstNonEmpty(values ...string) string {
	for _, value := range values {
		if value != "" {
			return value
		}
	}
	return ""
}
