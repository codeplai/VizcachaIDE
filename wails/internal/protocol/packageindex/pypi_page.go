package packageindex

import (
	"net/url"

	"github.com/codeplai/VizcachaIDE/wails/internal/domain"
)

// parsePyPIPage reads the results of https://pypi.org/search/: each result is an
// <a class="package-snippet"> holding .package-snippet__name, .package-snippet__version and
// .package-snippet__description. A page with no results says so in its text. Anything else (a
// challenge page for robots, a new layout) is not readable and gives ErrPackageIndexUnavailable.
func parsePyPIPage(body []byte) ([]domain.PackageInfo, error) {
	root, err := parseDocument(body)
	if err != nil {
		return nil, unavailable(err)
	}
	snippets := withClass(root, "package-snippet")
	if len(snippets) == 0 {
		if mentions(root, "no results") {
			return nil, nil
		}
		return nil, unavailable(errUnexpectedPage)
	}
	found := make([]domain.PackageInfo, 0, MaxResults)
	for _, snippet := range snippets {
		name := textOf(firstWithClass(snippet, "package-snippet__name"))
		if name == "" || len(found) == MaxResults {
			continue
		}
		found = append(found, domain.PackageInfo{
			Name:        name,
			Version:     textOf(firstWithClass(snippet, "package-snippet__version")),
			Description: textOf(firstWithClass(snippet, "package-snippet__description")),
			URL:         "https://pypi.org/project/" + url.PathEscape(name) + "/",
		})
	}
	if len(found) == 0 {
		return nil, unavailable(errUnexpectedPage)
	}
	return found, nil
}
